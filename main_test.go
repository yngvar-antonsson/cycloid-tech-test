package main

import (
	"os"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	content := `
providers:
  openmeteo: {}
  weatherapi:
    APIKey: "some-key"
`
	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Providers == nil {
		t.Fatal("expected non-nil providers")
	}

	if _, ok := cfg.Providers["openmeteo"]; !ok {
		t.Error("expected 'openmeteo' provider")
	}
	if val, ok := cfg.Providers["weatherapi"]["APIKey"]; !ok || val != "some-key" {
		t.Error("expected weatherapi.APIKey = 'some-key'")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := `this: is: not: valid: yaml:`
	tmpFile, err := os.CreateTemp("", "bad-config-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
