package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
	"github.com/sploov/uptime/internal/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Optimize SQLite for concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		latency_ns INTEGER NOT NULL,
		status INTEGER NOT NULL,
		error TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_checks_service_ts ON checks(service_id, timestamp);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *Store) AddCheck(r models.CheckResult) error {
	_, err := s.db.Exec(`
		INSERT INTO checks (service_id, timestamp, latency_ns, status, error)
		VALUES (?, ?, ?, ?, ?)
	`, r.ServiceID, r.Timestamp.Unix(), r.Latency.Nanoseconds(), r.Status, r.Error)
	return err
}

// GetRecentChecks returns the checks for the history graph (e.g., last 50)
func (s *Store) GetRecentChecks(serviceID string, limit int) ([]models.CheckResult, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, latency_ns, status, error
		FROM checks
		WHERE service_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, serviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CheckResult
	for rows.Next() {
		var ts int64
		var lat int64
		var status int
		var errStr sql.NullString
		if err := rows.Scan(&ts, &lat, &status, &errStr); err != nil {
			return nil, err
		}
		results = append(results, models.CheckResult{
			ServiceID: serviceID,
			Timestamp: time.Unix(ts, 0),
			Latency:   time.Duration(lat),
			Status:    models.HeartbeatValue(status),
			Error:     errStr.String,
		})
	}
	// Reverse to chronological order if needed, but usually recent first is fine for API, 
	// but the UI might expect chronological. Let's return as is (descending) and let API/UI handle it or reverse here.
	// The prompt asks for "heartbeats" array. Usually that's a timeline.
	// Let's reverse it to be chronological (oldest -> newest).
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results, nil
}

func (s *Store) GetUptime(serviceID string, since time.Time) (float64, error) {
	var up, total int
	err := s.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN status != 2 THEN 1 END),
			COUNT(*)
		FROM checks
		WHERE service_id = ? AND timestamp >= ?
	`, serviceID, since.Unix()).Scan(&up, &total)
	
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return (float64(up) / float64(total)) * 100, nil
}

func (s *Store) GetAverageLatency(serviceID string, since time.Time) (float64, error) {
	var avg sql.NullFloat64
	err := s.db.QueryRow(`
		SELECT AVG(latency_ns)
		FROM checks
		WHERE service_id = ? AND timestamp >= ? AND status != 2
	`, serviceID, since.Unix()).Scan(&avg)
	
	if err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64 / float64(time.Millisecond), nil // Convert ns to ms
}

func (s *Store) Close() error {
	return s.db.Close()
}
