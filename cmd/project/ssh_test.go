package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHasChangesDetectsModification(t *testing.T) {
	tempDir := t.TempDir()
	lastSync := time.Now()

	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	future := lastSync.Add(2 * time.Second)
	if err := os.Chtimes(filePath, future, future); err != nil {
		t.Fatalf("adjust file times: %v", err)
	}

	past := lastSync.Add(-1 * time.Second)
	if err := os.Chtimes(tempDir, past, past); err != nil {
		t.Fatalf("adjust dir times: %v", err)
	}

	changed, path := hasChanges(tempDir, lastSync)
	if !changed {
		t.Fatalf("expected changes to be detected")
	}
	if path != filePath {
		t.Fatalf("expected first modified file to be %s, got %s", filePath, path)
	}
}

func TestHasChangesDetectsRemoval(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	filePath := filepath.Join(subDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	lastSync := time.Now()
	past := lastSync.Add(-1 * time.Second)
	if err := os.Chtimes(tempDir, past, past); err != nil {
		t.Fatalf("adjust tempDir times: %v", err)
	}
	if err := os.Chtimes(subDir, past, past); err != nil {
		t.Fatalf("adjust subDir times: %v", err)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove file: %v", err)
	}

	future := lastSync.Add(2 * time.Second)
	if err := os.Chtimes(subDir, future, future); err != nil {
		t.Fatalf("adjust subDir times after removal: %v", err)
	}

	changed, path := hasChanges(tempDir, lastSync)
	if !changed {
		t.Fatalf("expected removal to be detected")
	}
	if path != subDir {
		t.Fatalf("expected first modified path to be %s, got %s", subDir, path)
	}
}
