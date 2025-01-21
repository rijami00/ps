package apollo

import (
	"bufio"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/haatos/goshipit/internal"
	"os"
	"strings"
)

// Instance is a struct that holds the configuration for instances running
// The variables we want to track are:
// host, port, name, description, backend version, backend build number, cdm id provider, staging mode (development, staging, production), uptime, deployment time, default user/pass for login etc

type Instance struct {
	Id                 string
	WorkingDir         string
	Url                string
	Port               uint16
	Name               string
	Description        string
	Image              string
	ContainerState     string
	ContainerStatus    string
	ApiStatus          string
	BackendVersion     string
	BackendBuildNumber string
	CdmIdProvider      string
	StagingMode        string
	Uptime             string
	DeploymentTime     string
	DbConnectionStatus string
}

const (
	runningState     = "running"
	unreachableState = "unreachable"
)

func BuildUrl(name string, domain string, port uint16) string {
	return fmt.Sprintf("https://%s.%s:%d", name, domain, port)
}

func GetPublicPort(container types.Container) uint16 {
	if len(container.Ports) > 0 && container.Ports[0].PublicPort != 0 {
		return container.Ports[0].PublicPort
	}
	return 0
}

// GetDomain Parse config file to get domain
// config file should contain domain=MY_DOMAIN_HERE as the first line
func GetDomain(configFile string) string {
	var fallback = internal.Settings.InstanceDomain
	// Open the file
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return fallback
	}
	defer file.Close()

	// Read the first line
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2) // Split only once
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "domain" {
			return strings.TrimSpace(parts[1]) // Return the value part
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Return default string if no valid domain found
	return fallback
}

// PrintInstancesHTML calls GetInstances() and returns an HTML table
// as a string, with each Instance printed in a table row.
func PrintInstancesHTML() (string, error) {
	instances, err := GetInstances()
	if err != nil {
		return "", err
	}

	// Start the HTML table
	html := "<table border='1' cellpadding='5' cellspacing='0'>\n"
	html += "  <thead>\n"
	html += "    <tr>\n"
	html += "      <th>Path</th>\n"
	html += "      <th>Host</th>\n"
	html += "      <th>Port</th>\n"
	html += "      <th>Name</th>\n"
	html += "      <th>Description</th>\n"
	html += "      <th>BackendVersion</th>\n"
	html += "      <th>BackendBuildNumber</th>\n"
	html += "      <th>CdmIdProvider</th>\n"
	html += "      <th>StagingMode</th>\n"
	html += "      <th>Uptime</th>\n"
	html += "      <th>DeploymentTime</th>\n"
	html += "    </tr>\n"
	html += "  </thead>\n"
	html += "  <tbody>\n"

	// Populate rows
	for _, inst := range instances {
		html += "    <tr>\n"
		html += fmt.Sprintf("      <td>%s</td>\n", inst.WorkingDir)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.Url)
		html += fmt.Sprintf("      <td>%d</td>\n", inst.Port)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.Name)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.Description)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.BackendVersion)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.BackendBuildNumber)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.CdmIdProvider)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.StagingMode)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.Uptime)
		html += fmt.Sprintf("      <td>%s</td>\n", inst.DeploymentTime)
		html += "    </tr>\n"
	}

	// Close table
	html += "  </tbody>\n"
	html += "</table>\n"

	return html, nil
}
