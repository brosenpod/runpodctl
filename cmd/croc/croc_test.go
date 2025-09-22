package croc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetRelays(t *testing.T) {
	expected := Response{
		Relays: []Relay{
			{Address: "relay1.example", Password: "pass1", Ports: "1234"},
			{Address: "relay2.example", Password: "pass2", Ports: "5678"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if err := json.NewEncoder(w).Encode(expected); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	originalRelayURL := relayUrl
	relayUrl = server.URL
	defer func() { relayUrl = originalRelayURL }()

	relays, err := getRelays()
	if err != nil {
		t.Fatalf("getRelays returned error: %v", err)
	}

	if !reflect.DeepEqual(relays, expected.Relays) {
		t.Fatalf("unexpected relays: %#v", relays)
	}
}

func TestGetFilesInfo(t *testing.T) {
	tempDir := t.TempDir()
	rootDir := filepath.Join(tempDir, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("failed to create root directory: %v", err)
	}

	rootFilePath := filepath.Join(rootDir, "file1.txt")
	if err := os.WriteFile(rootFilePath, []byte("root"), 0o644); err != nil {
		t.Fatalf("failed to create root file: %v", err)
	}

	subDir := filepath.Join(rootDir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	subFilePath := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(subFilePath, []byte("sub"), 0o644); err != nil {
		t.Fatalf("failed to create subdir file: %v", err)
	}

	emptyDir := filepath.Join(rootDir, "empty")
	if err := os.Mkdir(emptyDir, 0o755); err != nil {
		t.Fatalf("failed to create empty dir: %v", err)
	}

	filesInfo, emptyFolders, totalFolders, err := GetFilesInfo([]string{rootDir}, false)
	if err != nil {
		t.Fatalf("GetFilesInfo returned error: %v", err)
	}

	if len(filesInfo) != 2 {
		t.Fatalf("expected 2 files, got %d", len(filesInfo))
	}

	if totalFolders != 3 {
		t.Fatalf("expected 3 folders, got %d", totalFolders)
	}

	if len(emptyFolders) != 1 {
		t.Fatalf("expected 1 empty folder, got %d", len(emptyFolders))
	}

	emptyFolder := emptyFolders[0]
	if emptyFolder.FolderRemote != "root/empty/" {
		t.Fatalf("unexpected empty folder remote path: %q", emptyFolder.FolderRemote)
	}

	infoByName := make(map[string]FileInfo, len(filesInfo))
	for _, info := range filesInfo {
		infoByName[info.Name] = info
	}

	rootInfo, ok := infoByName["file1.txt"]
	if !ok {
		t.Fatalf("missing root file info")
	}
	if rootInfo.FolderRemote != "root/" {
		t.Fatalf("unexpected root FolderRemote: %q", rootInfo.FolderRemote)
	}
	if rootInfo.FolderSource != rootDir {
		t.Fatalf("unexpected root FolderSource: %q", rootInfo.FolderSource)
	}
	if rootInfo.TempFile {
		t.Fatalf("root file should not be marked as TempFile")
	}

	subInfo, ok := infoByName["file2.txt"]
	if !ok {
		t.Fatalf("missing subdir file info")
	}
	if subInfo.FolderRemote != "root/subdir/" {
		t.Fatalf("unexpected subdir FolderRemote: %q", subInfo.FolderRemote)
	}
	if subInfo.FolderSource != subDir {
		t.Fatalf("unexpected subdir FolderSource: %q", subInfo.FolderSource)
	}
	if subInfo.TempFile {
		t.Fatalf("subdir file should not be marked as TempFile")
	}
}

func TestGetFilesInfoZipFolder(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	tempWD := t.TempDir()
	if err := os.Chdir(tempWD); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	actualWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to resolve working directory: %v", err)
	}

	sourceDir := filepath.Join(actualWD, "data")
	if err := os.Mkdir(sourceDir, 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	sourceFile := filepath.Join(sourceDir, "file.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	filesInfo, emptyFolders, totalFolders, err := GetFilesInfo([]string{sourceDir}, true)
	if err != nil {
		t.Fatalf("GetFilesInfo returned error: %v", err)
	}

	if len(emptyFolders) != 0 {
		t.Fatalf("expected no empty folders, got %d", len(emptyFolders))
	}

	if totalFolders != 0 {
		t.Fatalf("expected total folders to be 0, got %d", totalFolders)
	}

	if len(filesInfo) != 1 {
		t.Fatalf("expected 1 file, got %d", len(filesInfo))
	}

	info := filesInfo[0]
	if info.Name != "data.zip" {
		t.Fatalf("unexpected zip name: %q", info.Name)
	}
	if info.FolderRemote != "./" {
		t.Fatalf("unexpected FolderRemote: %q", info.FolderRemote)
	}
	if info.FolderSource != actualWD {
		t.Fatalf("unexpected FolderSource: %q", info.FolderSource)
	}
	if !info.TempFile {
		t.Fatalf("expected TempFile to be true")
	}

	zipPath := filepath.Join(actualWD, "data.zip")
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("expected zip file to exist: %v", err)
	}
	if err := os.Remove(zipPath); err != nil {
		t.Fatalf("failed to remove zip file: %v", err)
	}
}
