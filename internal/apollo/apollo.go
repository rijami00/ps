package apollo

import (
	"cmp"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/haatos/goshipit/internal"
	"slices"
	"sort"
	"strings"
	"time"

	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// GetInstances lists all containers via Docker Engine API, and returns
// any whose Compose config path (label) matches `*/apollo/instances/*`.
func GetInstances() ([]Instance, error) {
	// Create a context and Docker client for the remote Docker host
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.WithHost(internal.Settings.ApolloDocker),
		client.WithAPIVersionNegotiation(),
	)
	var apolloDir = internal.Settings.ApolloDir
	// var apolloConfig = apolloDir + ".config"
	// var domain = GetDomain(apolloConfig)
	var domain = internal.Settings.Domain

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

	// first iterate over all containers and filter out the ones we want
	// we only want the ones where service label is proxy
	// if proxy is not available, then app of the same `com.docker.compose.project`
	// if that is not available, then we skip the container
	// then we will iterate over the filtered containers and get the details, create instance

	// let's declare a map (key value) so we can store containers against their project name
	// here we only want the proxy container
	entries := make(map[string]types.Container)

	// project -> service -> container map
	fullMap := make(map[string]map[string]types.Container)

	for _, container := range containers {
		service, _ := container.Labels["com.docker.compose.service"]
		project, _ := container.Labels["com.docker.compose.project"]

		if _, exists := fullMap[project]; !exists {
			fullMap[project] = make(map[string]types.Container)
		}
		fullMap[project][service] = container

		if service == "proxy" {
			entries[project] = container
		} else if service == "app" {
			// Only add an "app" container if no "proxy" container has been stored for this project
			if _, exists := entries[project]; !exists {
				entries[project] = container
			}
		}
	}

	for project, container := range entries {
		// Docker Compose (v2) typically adds labels like:
		//   com.docker.compose.project
		//   com.docker.compose.project.config_files
		//
		// For example: "com.docker.compose.project.config_files" => "/path/to/docker-compose.yaml"
		fmt.Println("Project: ", project)
		workingDir, ok := container.Labels["com.docker.compose.project.working_dir"]
		// service, _ := container.Labels["com.docker.compose.service"]
		if !ok {
			// If the container doesn't have this label,
			// it probably wasn't started by Docker Compose v2
			continue
		}

		// Check if the Compose file path contains "/apollo/instances/"
		if strings.Contains(workingDir, apolloDir) {
			// extract more details:
			id := container.ID
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
			instance := Instance{
				Id:              id,
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

			// now assign all the service status in the ContainerDetails of this project
			for _, services := range fullMap[project] {
				details := ContainerDetails{
					Id:              services.ID,
					Name:            services.Names[0],
					Image:           services.Image,
					Service:         services.Labels["com.docker.compose.service"],
					ContainerState:  services.State,
					ContainerStatus: services.Status,
				}
				instance.ContainerDetails = append(instance.ContainerDetails, details)

				// custom sorting
				sortOrder := map[string]int{
					"proxy": 1,
					"app":   2,
					"db":    3,
					"web":   4,
				}

				sort.Slice(instance.ContainerDetails, func(i, j int) bool {
					getPriority := func(service string) int {
						if priority, exists := sortOrder[service]; exists {
							return priority
						}
						return len(sortOrder) + 1 // Anything not in the map comes last in lexicographic order
					}

					pi, pj := getPriority(instance.ContainerDetails[i].Service), getPriority(instance.ContainerDetails[j].Service)
					if pi != pj {
						return pi < pj
					}
					return instance.ContainerDetails[i].Service < instance.ContainerDetails[j].Service
				})
			}

			if responseUp != nil {
				instance.ApiStatus = runningState
				// responseUp.Details can be null
				if responseUp.Details.Version == "" && responseUp.Details.IDProvider == "" {
					fmt.Println("Details are null or not provided in the response.")
				} else {
					instance.BackendVersion = responseUp.Details.Version
					instance.BackendCommitHash = responseUp.Details.CommitHash
					instance.BackendCommitUrl = fmt.Sprintf(internal.Settings.GitHubInstance + "/commit/" + responseUp.Details.CommitHash)
					instance.CdmIdProvider = responseUp.Details.IDProvider
					instance.StagingMode = responseUp.Details.Stage
					instance.BackendBuildNumber = responseUp.Details.BuildNumber
					instance.DbConnectionStatus = responseUp.Details.Database.Status
				}
			}
			instances = append(instances, instance)
		}
	}

	//for _, instance := range instances {
	//	litter.Dump(instance)
	//	break
	//}

	// sort the instances by port number

	slices.SortFunc(instances, func(i, j Instance) int {
		return cmp.Compare(i.Name, j.Name)
	})

	return instances, nil
}
