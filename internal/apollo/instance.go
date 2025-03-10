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
	ContainerDetails   []ContainerDetails
	ApiStatus          string
	BackendVersion     string
	BackendBuildNumber string
	BackendCommitHash  string
	BackendCommitUrl   string
	CdmIdProvider      string
	StagingMode        string
	Uptime             string
	DeploymentTime     string
	DbConnectionStatus string
}

type ContainerDetails struct {
	Id              string
	Name            string
	Service         string
	Image           string
	ContainerState  string
	ContainerStatus string
}

const (
	runningState     = "running"
	unreachableState = "unreachable"
)

func BuildUrl(name string, domain string, port uint16) string {
	return fmt.Sprintf("https://%s.%s:%d", name, domain, port)
}

func GetPublicPort(container types.Container) uint16 {
	// loop on container.Ports
	// find PrivatePort == 443 and then take the PublicPort
	// if not found, return 0
	for _, port := range container.Ports {
		if port.PrivatePort == 443 {
			return port.PublicPort
		}
	}
	if len(container.Ports) > 0 {
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
