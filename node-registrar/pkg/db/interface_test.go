package db

import (
	"encoding/json"
	"testing"
)

func TestInterfaceUnmarshalJSON_OldFormat(t *testing.T) {
	// Test old format where ips is a string with "/" separators
	jsonData := []byte(`{
		"name": "eth0",
		"mac": "00:11:22:33:44:55",
		"ips": "192.168.1.1/10.0.0.1"
	}`)

	var iface Interface
	err := json.Unmarshal(jsonData, &iface)
	if err != nil {
		t.Fatalf("Failed to unmarshal old format: %v", err)
	}

	if iface.Name != "eth0" {
		t.Errorf("Expected name 'eth0', got '%s'", iface.Name)
	}
	if len(iface.IPs) != 2 {
		t.Fatalf("Expected 2 IPs, got %d", len(iface.IPs))
	}
	if iface.IPs[0] != "192.168.1.1" {
		t.Errorf("Expected first IP '192.168.1.1', got '%s'", iface.IPs[0])
	}
	if iface.IPs[1] != "10.0.0.1" {
		t.Errorf("Expected second IP '10.0.0.1', got '%s'", iface.IPs[1])
	}
}

func TestInterfaceUnmarshalJSON_NewFormat(t *testing.T) {
	// Test new format where ips is an array of strings
	jsonData := []byte(`{
		"name": "eth0",
		"mac": "00:11:22:33:44:55",
		"ips": ["192.168.1.1", "10.0.0.1"]
	}`)

	var iface Interface
	err := json.Unmarshal(jsonData, &iface)
	if err != nil {
		t.Fatalf("Failed to unmarshal new format: %v", err)
	}

	if iface.Name != "eth0" {
		t.Errorf("Expected name 'eth0', got '%s'", iface.Name)
	}
	if len(iface.IPs) != 2 {
		t.Fatalf("Expected 2 IPs, got %d", len(iface.IPs))
	}
	if iface.IPs[0] != "192.168.1.1" {
		t.Errorf("Expected first IP '192.168.1.1', got '%s'", iface.IPs[0])
	}
	if iface.IPs[1] != "10.0.0.1" {
		t.Errorf("Expected second IP '10.0.0.1', got '%s'", iface.IPs[1])
	}
}

func TestInterfaceUnmarshalJSON_EmptyString(t *testing.T) {
	// Test empty string for ips
	jsonData := []byte(`{
		"name": "eth0",
		"mac": "00:11:22:33:44:55",
		"ips": ""
	}`)

	var iface Interface
	err := json.Unmarshal(jsonData, &iface)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty string: %v", err)
	}

	if len(iface.IPs) != 0 {
		t.Errorf("Expected 0 IPs for empty string, got %d", len(iface.IPs))
	}
}

func TestInterfaceUnmarshalJSON_EmptyArray(t *testing.T) {
	// Test empty array for ips
	jsonData := []byte(`{
		"name": "eth0",
		"mac": "00:11:22:33:44:55",
		"ips": []
	}`)

	var iface Interface
	err := json.Unmarshal(jsonData, &iface)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty array: %v", err)
	}

	if len(iface.IPs) != 0 {
		t.Errorf("Expected 0 IPs for empty array, got %d", len(iface.IPs))
	}
}
