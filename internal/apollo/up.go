package apollo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ResponseUp struct {
	Status  string `json:"status"`
	Details struct {
		Database struct {
			Status string `json:"status"`
		} `json:"database"`
		Version     string `json:"version"`
		BuildNumber string `json:"buildNumber"`
		CommitHash  string `json:"commitHash"`
		IDProvider  string `json:"idProvider"`
		Stage       string `json:"stage"`
	} `json:"details"`
}

type ResponseUpFe struct {
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	CommitHash  string `json:"commitHash"`
}

func getUp(url string) (*ResponseUp, error) {
	// Append /up to the provided URL
	fullURL := url + "/up"

	// Perform HTTP GET request
	// we need to make a get request to the /up endpoint but we also need to add a customer header: api=true

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Add("api", "true")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected HTTP status: " + resp.Status)
	}

	// Decode JSON response
	var response ResponseUp
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// Handle case where details may be null
	if response.Details.Version == "" && response.Details.IDProvider == "" {
		response.Details = struct {
			Database struct {
				Status string `json:"status"`
			} `json:"database"`
			Version     string `json:"version"`
			BuildNumber string `json:"buildNumber"`
			CommitHash  string `json:"commitHash"`
			IDProvider  string `json:"idProvider"`
			Stage       string `json:"stage"`
		}{}
	}

	return &response, nil
}

func getUpFe(url string) (*ResponseUpFe, error) {
	// Append /up to the provided URL
	fullURL := url + "/VERSION"

	// Perform HTTP GET request
	// we need to make a get request to the /VERSION file

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected HTTP status: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	versionStr := strings.TrimSpace(string(bodyBytes))
	// fmt.Println(versionStr)

	// Parse version string like: 1.2.0-SNAPSHOT+196.c1a5781
	parts := strings.Split(versionStr, "+")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}
	version := parts[0]

	buildHashParts := strings.Split(parts[1], ".")
	if len(buildHashParts) != 2 {
		return nil, fmt.Errorf("invalid build.hash format: %s", parts[1])
	}
	build := buildHashParts[0]
	hash := buildHashParts[1]

	response := ResponseUpFe{
		Version:     version,
		BuildNumber: build,
		CommitHash:  hash,
	}

	fmt.Println(response)

	return &response, nil
}
