package project

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestCreateEnvVarsAndFormatting(t *testing.T) {
	configData := `
name = "sample"
[project]
uuid = "project-123"
  [project.env_vars]
  API_KEY = "secret"
  TIMEOUT = "30"
`

	tree, err := toml.Load(configData)
	if err != nil {
		t.Fatalf("toml.Load: %v", err)
	}

	env := createEnvVars(tree)
	expected := map[string]string{
		"API_KEY":           "secret",
		"TIMEOUT":           "30",
		"RUNPOD_PROJECT_ID": "project-123",
	}

	if len(env) != len(expected) {
		t.Fatalf("expected %d env vars, got %d", len(expected), len(env))
	}

	for key, want := range expected {
		if got, ok := env[key]; !ok || got != want {
			t.Fatalf("env[%q] = %q, want %q", key, got, want)
		}
	}

	podEnv := mapToApiEnv(env)
	if len(podEnv) != len(expected) {
		t.Fatalf("expected %d pod env entries, got %d", len(expected), len(podEnv))
	}

	for _, item := range podEnv {
		if want, ok := expected[item.Key]; !ok || want != item.Value {
			t.Fatalf("unexpected pod env entry: %v", item)
		}
	}

	dockerEnv := formatAsDockerEnv(env)
	for key, value := range expected {
		if !strings.Contains(dockerEnv, "ENV "+key+"="+value+"\n") {
			t.Fatalf("docker env missing entry for %s", key)
		}
	}
}

func TestBaseDockerImage(t *testing.T) {
	got := baseDockerImage("12.2")
	want := "runpod/base:0.4.4-cuda12.2"
	if got != want {
		t.Fatalf("baseDockerImage mismatch: got %q, want %q", got, want)
	}
}

func TestGetDefaultModelName(t *testing.T) {
	cases := map[string]string{
		"LLM":              "google/flan-t5-base",
		"Stable_Diffusion": "stabilityai/sdxl-turbo",
		"Text_to_Audio":    "facebook/musicgen-small",
		"Unknown":          "",
	}

	for input, want := range cases {
		if got := getDefaultModelName(input); got != want {
			t.Fatalf("getDefaultModelName(%q) = %q, want %q", input, got, want)
		}
	}
}
