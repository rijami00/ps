package apollo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sanity-io/litter"

	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// GetInstances lists all containers via Docker Engine API, and returns
// any whose Compose config path (label) matches `*/apollo/instances/*`.
func GetInstances() ([]Instance, error) {
	// Create a context and Docker client for the remote Docker host
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.WithHost("tcp://192.168.1.17:2376"),
		client.WithAPIVersionNegotiation(),
	)
	const apolloDir = "/home/cdm/apollo/"
	const apolloConfig = apolloDir + ".config"
	var domain = GetDomain(apolloConfig)

	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}
	defer cli.Close()

	// List all containers
	containers, err := cli.ContainerList(ctx, containerTypes.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}

	var instances []Instance

	for _, container := range containers {
		// Docker Compose (v2) typically adds labels like:
		//   com.docker.compose.project
		//   com.docker.compose.project.config_files
		//
		// For example: "com.docker.compose.project.config_files" => "/path/to/docker-compose.yaml"
		workingDir, ok := container.Labels["com.docker.compose.project.working_dir"]
		service, _ := container.Labels["com.docker.compose.service"]
		if !ok {
			// If the container doesn't have this label,
			// it probably wasn't started by Docker Compose v2
			continue
		}
		if service != "app" {
			// If the service isn't "app", it's not an Apollo api server instance
			continue
		}

		// Check if the Compose file path contains "/apollo/instances/"
		if strings.Contains(workingDir, apolloDir) {
			// extract more details:
			name := container.Labels["com.docker.compose.project"]
			port := GetPublicPort(container)
			url := BuildUrl(name, domain, port)
			image := container.Image
			state := container.State
			status := container.Status
			created := time.Unix(container.Created, 0).Format(time.RFC822Z)

			responseUp, err := getUp(url)

			var apiStatus string

			if err != nil {
				apiStatus = unreachableState
			}

			// We could parse container.Ports to get the actual port if needed
			// For now, just fill what we can
			i := Instance{
				WorkingDir:      workingDir,
				Image:           image,
				ContainerState:  state,
				ContainerStatus: status,
				Name:            name,
				Port:            port,
				Url:             url,
				DeploymentTime:  created,
				ApiStatus:       apiStatus,
				// Host, Port, and other fields can be set from container inspection if desired
			}
			if responseUp != nil {
				i.ApiStatus = runningState
				// responseUp.Details can be null
				if responseUp.Details.Version == "" && responseUp.Details.IDProvider == "" {
					fmt.Println("Details are null or not provided in the response.")
				} else {
					i.BackendVersion = responseUp.Details.Version
					i.CdmIdProvider = responseUp.Details.IDProvider
					i.StagingMode = responseUp.Details.Stage
					i.BackendBuildNumber = responseUp.Details.BuildNumber
					i.DbConnectionStatus = responseUp.Details.Database.Status
				}
			}
			instances = append(instances, i)
		}
	}

	for _, instance := range instances {
		litter.Dump(instance)
		//break
	}
	return instances, nil
}
