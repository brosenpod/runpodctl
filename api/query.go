package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Input struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func Query(input Input) (res *http.Response, err error) {
	jsonValue, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	apiUrl := resolveAPIURL()

	apiKey, err := resolveAPIKey()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonValue))
	if err != nil {
		return
	}

	sanitizedVersion := strings.TrimRight(Version, "\r\n")
	userAgent := "RunPod-CLI/" + sanitizedVersion + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: time.Second * 10}
	return client.Do(req)
}

func resolveAPIURL() string {
	if apiUrl := strings.TrimSpace(os.Getenv("RUNPOD_API_URL")); apiUrl != "" {
		return apiUrl
	}

	if apiUrl := strings.TrimSpace(viper.GetString("apiUrl")); apiUrl != "" {
		return apiUrl
	}

	return "https://api.runpod.io/graphql"
}

func resolveAPIKey() (string, error) {
	candidates := []string{
		os.Getenv("RUNPOD_API_KEY"),
		os.Getenv("RUNPOD_API_TOKEN"),
		viper.GetString("apiKey"),
		viper.GetString("api_key"),
		viper.GetString("RUNPOD_API_KEY"),
	}

	for _, candidate := range candidates {
		apiKey := strings.TrimSpace(candidate)
		if apiKey != "" {
			return apiKey, nil
		}
	}

	fmt.Println("API key not found")
	return "", errors.New("API key not found")
}
