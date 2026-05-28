package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/egermano/balde/cli"
)

func TestInitCmd_CreatesDB(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"init"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	dbPath := filepath.Join(dir, "balde.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected balde.db to be created")
	}

	configPath := filepath.Join(dir, "balde.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected balde.json config to be created")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if config["frequency"] != "monthly" {
		t.Errorf("expected frequency=monthly, got %v", config["frequency"])
	}
	if config["currency_symbol"] != "$" {
		t.Errorf("expected currency_symbol=$, got %v", config["currency_symbol"])
	}
	if config["decimal_separator"] != "." {
		t.Errorf("expected decimal_separator=., got %v", config["decimal_separator"])
	}
	if config["thousands_separator"] != "," {
		t.Errorf("expected thousands_separator=,, got %v", config["thousands_separator"])
	}
}
