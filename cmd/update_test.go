package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	t.Parallel()

	binaryContent := []byte{0x00, 0x01, 0x02, 0x03}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(binaryContent); err != nil {
				panic(err)
			}
		case "/bad":
			http.Error(w, "error", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "binary")

	file, err := DownloadFile(server.URL+"/ok", destPath)
	if err != nil {
		t.Fatalf("DownloadFile returned unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if file != nil {
			file.Close()
		}
	})

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed reading downloaded file: %v", err)
	}
	if !bytes.Equal(data, binaryContent) {
		t.Fatalf("downloaded content mismatch: got %v want %v", data, binaryContent)
	}

	if _, err := DownloadFile(server.URL+"/bad", filepath.Join(tempDir, "bad")); err == nil {
		t.Fatalf("expected error for non-200 response, got nil")
	}
}

func TestGetJson(t *testing.T) {
	t.Parallel()

	payload := GithubApiResponse{
		Version: "v1.2.3",
		Assets: []Asset{{
			Url:  "https://example.com/runpodctl",
			Name: "runpodctl-linux-amd64",
		}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			panic(err)
		}
	}))
	t.Cleanup(server.Close)

	resp, err := GetJson(server.URL)
	if err != nil {
		t.Fatalf("GetJson returned unexpected error: %v", err)
	}

	if resp.Version != payload.Version {
		t.Fatalf("unexpected version: got %q want %q", resp.Version, payload.Version)
	}

	if len(resp.Assets) != len(payload.Assets) {
		t.Fatalf("unexpected assets length: got %d want %d", len(resp.Assets), len(payload.Assets))
	}

	if resp.Assets[0] != payload.Assets[0] {
		t.Fatalf("unexpected asset: got %#v want %#v", resp.Assets[0], payload.Assets[0])
	}
}
