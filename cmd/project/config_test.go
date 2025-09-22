package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProjectToml(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Errorf("failed to restore working directory: %v", chdirErr)
		}
	})

	projectName := "integration-test"
	cudaVersion := "12.1"
	pythonVersion := "3.9"

	generateProjectToml(".", "runpod.toml", projectName, cudaVersion, pythonVersion)

	tomlPath := filepath.Join(tempDir, "runpod.toml")
	t.Cleanup(func() {
		os.Remove(tomlPath)
	})

	contentBytes, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatalf("failed to read generated runpod.toml: %v", err)
	}
	content := string(contentBytes)

	if !strings.Contains(content, fmt.Sprintf("name = \"%s\"", projectName)) {
		t.Fatalf("expected project name %q to appear in runpod.toml", projectName)
	}
	if !strings.Contains(content, fmt.Sprintf("cuda%s", cudaVersion)) {
		t.Fatalf("expected CUDA version %q to appear in runpod.toml", cudaVersion)
	}
	if !strings.Contains(content, fmt.Sprintf("python_version = \"%s\"", pythonVersion)) {
		t.Fatalf("expected python version %q to appear in runpod.toml", pythonVersion)
	}
}

func TestBuildProjectDockerfile(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Errorf("failed to restore working directory: %v", chdirErr)
		}
	})

	tomlContent := `name = "docker-test"

[project]
uuid = "unit-test"
base_image = "example/base:1.0"

  [project.env_vars]
  CUSTOM = "value"

[runtime]
python_version = "3.10"
handler_path = "app/handler.py"
requirements_path = "requirements.txt"
`
	if err := os.WriteFile("runpod.toml", []byte(tomlContent), 0o644); err != nil {
		t.Fatalf("failed to write runpod.toml: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(filepath.Join(tempDir, "runpod.toml"))
	})

	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	t.Cleanup(func() {
		os.Remove(dockerfilePath)
	})

	previousIncludeEnv := includeEnvInDockerfile
	includeEnvInDockerfile = false
	t.Cleanup(func() {
		includeEnvInDockerfile = previousIncludeEnv
	})

	buildProjectDockerfile()

	dockerfileBytes, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("failed to read Dockerfile: %v", err)
	}
	dockerfileContent := string(dockerfileBytes)

	if !strings.Contains(dockerfileContent, "FROM example/base:1.0") {
		t.Fatalf("expected base image to be replaced in Dockerfile")
	}
	if strings.Contains(dockerfileContent, "<<BASE_IMAGE>>") {
		t.Fatalf("base image placeholder was not replaced")
	}
	if strings.Contains(dockerfileContent, "<<HANDLER_PATH>>") {
		t.Fatalf("handler path placeholder was not replaced")
	}
	if !strings.Contains(dockerfileContent, "/app/handler.py") {
		t.Fatalf("expected handler path to be present in Dockerfile")
	}
	if strings.Contains(dockerfileContent, "<<SET_ENV_VARS>>") {
		t.Fatalf("env placeholder was not removed when includeEnvInDockerfile is false")
	}
	if strings.Contains(dockerfileContent, "ENV CUSTOM=value") {
		t.Fatalf("environment variables should not be added when includeEnvInDockerfile is false")
	}

	includeEnvInDockerfile = true
	buildProjectDockerfile()

	dockerfileBytes, err = os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("failed to read Dockerfile after enabling env: %v", err)
	}
	dockerfileContent = string(dockerfileBytes)

	if strings.Contains(dockerfileContent, "<<SET_ENV_VARS>>") {
		t.Fatalf("env placeholder was not replaced when includeEnvInDockerfile is true")
	}
	if !strings.Contains(dockerfileContent, "ENV CUSTOM=value") {
		t.Fatalf("expected custom environment variable to be present in Dockerfile")
	}
	if !strings.Contains(dockerfileContent, "ENV RUNPOD_PROJECT_ID=unit-test") {
		t.Fatalf("expected RUNPOD_PROJECT_ID environment variable to be present in Dockerfile")
	}
}
