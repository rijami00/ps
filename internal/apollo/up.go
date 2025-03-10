package apollo

import (
	"encoding/json"
	"errors"
	"fmt"
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
		CommitHash  string `json:"commitHash"`
		IDProvider  string `json:"idProvider"`
		Stage       string `json:"stage"`
	} `json:"details"`
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
