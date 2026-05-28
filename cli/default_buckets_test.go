package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/egermano/balde/cli"
)

func TestInitCmd_CreatesDefaultBuckets(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	setupInitBudget(t)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"view", "buckets", "--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(os.Stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("view buckets failed: %v", err)
	}

	var buckets []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(buf.Bytes(), &buckets); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, buf.String())
	}

	if len(buckets) != 6 {
		t.Errorf("expected 6 default buckets, got %d\noutput: %s", len(buckets), buf.String())
	}

	expectedNames := map[string]bool{
		"financial freedom": false,
		"fixed costs":       false,
		"pleasures":         false,
		"comfort":           false,
		"knowledge":         false,
		"goals":             false,
	}

	for _, b := range buckets {
		if _, ok := expectedNames[b.Name]; ok {
			expectedNames[b.Name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("missing default bucket: %s", name)
		}
	}
}
