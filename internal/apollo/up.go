package apollo

import (
	"encoding/json"
	"errors"
	"fmt"
	ago "github.com/SerhiiCho/timeago/v3"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ResponseUp struct {
	Status  string `json:"status"`
	Details struct {
		Database struct {
			Status string `json:"status"`
		} `json:"database"`
		Version        string `json:"version"`
		BuildNumber    string `json:"buildNumber"`
		CommitHash     string `json:"commitHash"`
		CommitDateTime string `json:"commitDateTime"`
		IDProvider     string `json:"idProvider"`
		Stage          string `json:"stage"`
	} `json:"details"`
}

type ResponseUpFe struct {
	Version        string `json:"version"`
	BuildNumber    string `json:"buildNumber"`
	CommitHash     string `json:"commitHash"`
	CommitDateTime string `json:"commitDateTime"`
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
			Version        string `json:"version"`
			BuildNumber    string `json:"buildNumber"`
			CommitHash     string `json:"commitHash"`
			CommitDateTime string `json:"commitDateTime"`
			IDProvider     string `json:"idProvider"`
			Stage          string `json:"stage"`
		}{}
	}

	// Format commit timestamp (must come after details overwrite check)
	response.Details.CommitDateTime = unixToTime(response.Details.CommitDateTime)

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

	build := buildHashParts[0]
	hash := "UNKNOWN"
	commitDateTime := "UNKNOWN"

	if len(buildHashParts) >= 2 {
		hash = buildHashParts[1]
	}
	if len(buildHashParts) >= 3 {
		commitDateTime = buildHashParts[2]
		commitDateTime = unixToTime(commitDateTime)
	}

	response := ResponseUpFe{
		Version:        version,
		BuildNumber:    build,
		CommitHash:     hash,
		CommitDateTime: commitDateTime,
	}

	fmt.Println(response)

	return &response, nil
}

// Take a unix timestamp in string, convert it to int64, and return a human-readable date string
func unixToTime(unixTimestamp string) string {
	sec, err := strconv.ParseInt(unixTimestamp, 10, 64)
	if err != nil {
		return "UNKNOWN"
	}
	t := time.Unix(sec, 0).Local()
	return t.Format("2006-01-02 15:04:05")
}

func prettyTime(unixTimestamp string) string {
	sec, err := strconv.Atoi(unixTimestamp)
	if err != nil {
		return "UNKNOWN"
	}
	prettyTime, _ := ago.Parse(sec)
	return prettyTime
}
