package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// verifyGvisorInstallation checks if gVisor is properly installed
func verifyGvisorInstallation(cli *client.Client) error {
	// Check if runsc binary exists
	if _, err := os.Stat("/usr/local/bin/runsc"); err != nil {
		return fmt.Errorf("gVisor runtime (runsc) not found in /usr/local/bin/runsc")
	}

	// Verify runtime in Docker
	info, err := cli.Info(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get Docker info: %v", err)
	}

	runtime, exists := info.Runtimes["runsc"]
	if !exists {
		return fmt.Errorf("runsc runtime not configured in Docker")
	}

	// Verify runtime path matches
	if runtime.Path != "/usr/local/bin/runsc" {
		return fmt.Errorf("runsc runtime path mismatch")
	}

	return nil
}

// checks if the container is running with gVisor runtime
func monitorContainer(ctx context.Context, cli *client.Client, containerID string, done chan<- error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			inspect, err := cli.ContainerInspect(ctx, containerID)
			if err != nil {
				done <- fmt.Errorf("container inspection failed: %v", err)
				return
			}
			if inspect.HostConfig.Runtime != "runsc" {
				done <- fmt.Errorf("security violation: runtime changed from gVisor")
				return
			}
		}
	}
}

func main() {
	// Ensure we're running on Linux
	if runtime.GOOS != "linux" {
		panic("This program must run on Linux")
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Verify gVisor installation before proceeding
	if err := verifyGvisorInstallation(cli); err != nil {
		panic(fmt.Sprintf("gVisor verification failed: %v", err))
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", image.PullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	// Create container with enforced gVisor runtime and security options
	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: "alpine",
			Cmd:   []string{"echo", "hello world"},
		},
		&container.HostConfig{
			Runtime: "runsc",
			SecurityOpt: []string{
				"no-new-privileges=true",
			},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		panic(err)
	}

	// Panic if the container is missing
	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		panic(fmt.Sprintf("failed to inspect container: %v", err))
	}

	// Panic if the container is not using the gVisor runtime
	if inspect.HostConfig.Runtime != "runsc" {
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		panic("security violation: container not using gVisor runtime")
	}

	// Start monitoring the container for security violations
	monitorDone := make(chan error)
	go monitorContainer(ctx, cli, resp.ID, monitorDone)

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	// Monitor security while waiting for the container to finish
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	case err := <-monitorDone:
		// If monitor detected a security violation
		panic(err)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out) // later, this will be encrypted and returned, not printed
}
