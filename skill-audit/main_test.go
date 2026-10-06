package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtures = "internal/audit/testdata"

func TestCLI(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"-no-color", fixtures}, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "Summary") {
		t.Errorf("no summary:\n%s", out.String())
	}

	out.Reset()
	if code := run([]string{"-strict", "-format", "json", fixtures}, &out, &errb); code != 1 {
		t.Errorf("strict exit = %d, want 1", code)
	}
	var reps []map[string]any
	if err := json.Unmarshal(out.Bytes(), &reps); err != nil || len(reps) != 2 {
		t.Errorf("json: %v (%d reports)", err, len(reps))
	}

	if code := run([]string{"-strict", "-strict-warn", filepath.Join(fixtures, "good-skill")}, &out, &errb); code != 0 {
		t.Errorf("good skill strict exit = %d", code)
	}

	dir := t.TempDir()
	out.Reset()
	if code := run([]string{"-prompt", "-prompt-dir", dir, fixtures}, &out, &errb); code != 0 {
		t.Fatalf("prompt exit %d", code)
	}
	if !strings.Contains(out.String(), "# Skill fixes (1 skill(s))") {
		t.Errorf("prompt header:\n%.300s", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "Bad_Skill.prompt.md")); err != nil {
		t.Error(err)
	}

	if code := run([]string{"-only", "nope", fixtures}, &out, &errb); code != 2 {
		t.Errorf("bad -only exit = %d", code)
	}
	out.Reset()
	if code := run([]string{"-list-checks"}, &out, &errb); code != 0 || strings.Count(out.String(), "\n") < 11 {
		t.Errorf("list-checks: %d\n%s", code, out.String())
	}
}
