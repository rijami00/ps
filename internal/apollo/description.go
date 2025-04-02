package apollo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// given a container name, we will return the description
// we read a file called description.json and it contains the node "descriptions"
// which is an array containing two items: "instance" and "description"
// if the name starts with a value from "instance", then we will return the corresponding description

// Description struct to match the JSON structure
type Description struct {
	Instance    string `json:"instance"`
	Description string `json:"description"`
}

// Global variable to store descriptions
var descriptions []Description

// Initialize descriptions at startup
func init() {
	data, err := os.ReadFile("description.json")
	if err != nil {
		fmt.Println("Error reading descriptions:", err)
		descriptions = []Description{} // Empty slice as fallback
		return
	}

	var result struct {
		Descriptions []Description `json:"descriptions"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		fmt.Println("Error parsing descriptions:", err)
		descriptions = []Description{} // Empty slice as fallback
		return
	}

	descriptions = result.Descriptions
}

// GetContainerDescription returns the description for a given container name
func GetContainerDescription(containerName string) string {
	for _, item := range descriptions {
		if strings.HasPrefix(containerName, item.Instance) {
			fmt.Println("Found description for container:", containerName, "->", item.Description)
			return item.Description
		}
	}
	return "" // Return empty string instead of null
}
