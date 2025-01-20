package apollo

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ResponseUp struct {
	Status  string `json:"status"`
	Details struct {
		Database struct {
			Status string `json:"status"`
		} `json:"database"`
		Version     string `json:"version"`
		BuildNumber string `json:"buildNumber"`
		IDProvider  string `json:"idProvider"`
		Stage       string `json:"stage"`
	} `json:"details"`
}

func getUp(url string) (*ResponseUp, error) {
	// Append /up to the provided URL
	fullURL := url + "/up"

	// Perform HTTP GET request
	resp, err := http.Get(fullURL)
	if err != nil {
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
			IDProvider  string `json:"idProvider"`
			Stage       string `json:"stage"`
		}{}
	}

	return &response, nil
}
