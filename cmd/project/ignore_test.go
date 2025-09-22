package project

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func withTempChdir(t *testing.T, dir string) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}
	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}
}

func TestGetIgnoreListStripsCommentsAndWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	cleanup := withTempChdir(t, tempDir)
	t.Cleanup(cleanup)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	contents := "# leading comment\n\n custom.log \nassets/\n# trailing comment\n"
	if err := os.WriteFile(filepath.Join(cwd, ".runpodignore"), []byte(contents), 0o644); err != nil {
		t.Fatalf("write ignore file: %v", err)
	}

	ignoreList, err := GetIgnoreList()
	if err != nil {
		t.Fatalf("GetIgnoreList: %v", err)
	}

	expectedPatterns := append([]string{}, EXCLUDE_PATTERNS...)
	expectedPatterns = append(expectedPatterns, "custom.log", "assets/")

	sort.Strings(ignoreList)
	sort.Strings(expectedPatterns)

	if len(ignoreList) != len(expectedPatterns) {
		t.Fatalf("expected %d patterns, got %d", len(expectedPatterns), len(ignoreList))
	}

	for i := range expectedPatterns {
		if ignoreList[i] != expectedPatterns[i] {
			t.Fatalf("pattern mismatch at %d: expected %q, got %q", i, expectedPatterns[i], ignoreList[i])
		}
	}
}

func TestShouldIgnoreRespectsDirectoryPatterns(t *testing.T) {
	tempDir := t.TempDir()
	cleanup := withTempChdir(t, tempDir)
	t.Cleanup(cleanup)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	if err := os.WriteFile(filepath.Join(cwd, ".runpodignore"), []byte("build/\n"), 0o644); err != nil {
		t.Fatalf("write ignore file: %v", err)
	}

	buildDir := filepath.Join(cwd, "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		t.Fatalf("create build directory: %v", err)
	}

	targetFile := filepath.Join(buildDir, "output.txt")
	if err := os.WriteFile(targetFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}

	ignoreList, err := GetIgnoreList()
	if err != nil {
		t.Fatalf("GetIgnoreList: %v", err)
	}

	ignored, err := ShouldIgnore(targetFile, ignoreList)
	if err != nil {
		t.Fatalf("ShouldIgnore: %v", err)
	}
	if !ignored {
		t.Fatalf("expected %s to be ignored", targetFile)
	}

	otherFile := filepath.Join(cwd, "keep.txt")
	if err := os.WriteFile(otherFile, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write keep file: %v", err)
	}

	ignored, err = ShouldIgnore(otherFile, ignoreList)
	if err != nil {
		t.Fatalf("ShouldIgnore other: %v", err)
	}
	if ignored {
		t.Fatalf("did not expect %s to be ignored", otherFile)
	}
}
