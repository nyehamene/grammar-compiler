package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var updateFlag = os.Getenv("UPDATE_SNAPSHOTS") == "true"

func getModuleRoot() string {
	cwd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return "."
}

var moduleRoot = getModuleRoot()

func SnapshotPath(t interface {
	Fatalf(format string, args ...any)
}, name string) string {
	ts := t.(interface {
		Name() string
	})
	testName := ts.Name()

	var testFile string
	callers := strings.Split(testName, "/")
	if len(callers) > 0 {
		testFile = callers[len(callers)-1]
	}

	snapshotDir := filepath.Join(moduleRoot, "testdata", "snapshots", testFile)
	snapshotFile := filepath.Join(snapshotDir, name+".json")

	if updateFlag {
		os.MkdirAll(snapshotDir, 0755)
		return snapshotFile
	}

	if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
		t.Fatalf("snapshot %s does not exist. Run with UPDATE_SNAPSHOTS=true to create it.", snapshotFile)
	}

	return snapshotFile
}

func LoadSnapshot(t interface {
	Fatalf(format string, args ...any)
}, name string) []byte {
	snapshotFile := SnapshotPath(t, name)
	data, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	return data
}

func SaveSnapshot(t interface {
	Fatalf(format string, args ...any)
}, name string, data []byte) {
	snapshotFile := SnapshotPath(t, name)
	err := os.WriteFile(snapshotFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to write snapshot: %v", err)
	}
}

func AssertSnapshotJSON(t interface {
	Fatalf(format string, args ...any)
	Name() string
	Logf(format string, args ...any)
}, name string, got any) {
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal got: %v", err)
	}

	snapshotFile := SnapshotPath(t, name)

	if updateFlag {
		os.MkdirAll(filepath.Dir(snapshotFile), 0755)
		err := os.WriteFile(snapshotFile, gotJSON, 0644)
		if err != nil {
			t.Fatalf("failed to write snapshot: %v", err)
		}
		t.Logf("snapshot %s updated", snapshotFile)
		return
	}

	expected, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	if string(gotJSON) != string(expected) {
		t.Fatalf("snapshot mismatch for %s:\n\ngot:\n%s\n\nexpected:\n%s\n", name, gotJSON, expected)
	}
}

func AssertSnapshotText(t interface {
	Fatalf(format string, args ...any)
	Name() string
	Logf(format string, args ...any)
}, name string, got string) {
	snapshotFile := SnapshotPath(t, name)

	if updateFlag {
		os.MkdirAll(filepath.Dir(snapshotFile), 0755)
		err := os.WriteFile(snapshotFile, []byte(got), 0644)
		if err != nil {
			t.Fatalf("failed to write snapshot: %v", err)
		}
		t.Logf("snapshot %s updated", snapshotFile)
		return
	}

	expected, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	if got != string(expected) {
		t.Fatalf("snapshot mismatch for %s:\n\ngot:\n%s\n\nexpected:\n%s\n", name, got, expected)
	}
}

func SnapshotFilename(t interface{ Name() string }, name string) string {
	ts := t.(interface {
		Name() string
	})
	testName := ts.Name()

	callers := strings.Split(testName, "/")
	testFile := "default"
	if len(callers) > 0 {
		testFile = callers[len(callers)-1]
	}

	ext := filepath.Ext(name)
	if ext == "" {
		ext = ".json"
	}

	snapshotDir := filepath.Join("testdata", "snapshots", testFile)
	os.MkdirAll(snapshotDir, 0755)

	baseName := strings.TrimSuffix(name, ext)
	return filepath.Join(snapshotDir, baseName+ext)
}

func MustMarshalJSON(v any) string {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return string(bytes)
}
