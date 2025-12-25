package monitor

import (
	"testing"
	"time"
)

func TestCheckHTTP(t *testing.T) {
	// Test with a known good URL
	err := checkHTTP("https://www.google.com", 5*time.Second)
	if err != nil {
		t.Errorf("Expected no error for google.com, got %v", err)
	}

	// Test with a non-existent URL
	err = checkHTTP("https://this-is-a-very-unlikely-domain-name.com", 2*time.Second)
	if err == nil {
		t.Error("Expected error for non-existent domain, got nil")
	}
}

func TestCheckTCP(t *testing.T) {
	// Test with a known open port (standard HTTP port on google.com)
	err := checkTCP("google.com:80", 5*time.Second)
	if err != nil {
		t.Errorf("Expected no error for google.com:80, got %v", err)
	}
}
