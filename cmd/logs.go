package cmd

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var follow bool

var logsCmd = &cobra.Command{
	Use:   "logs [service...]",
	Short: "Fetch logs from service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)

		if len(targets) == 0 {
			return fmt.Errorf("no services found — specify a service name or use -f to point to a compose file")
		}

		if len(targets) == 1 {
			return backend.Logs(backend.ContainerName(proj, targets[0]), follow)
		}

		// Multiple services: fan out goroutines, prefix each line with service name
		var wg sync.WaitGroup
		for _, name := range targets {
			wg.Add(1)
			go func(svc string) {
				defer wg.Done()
				streamLogs(backend.ContainerName(proj, svc), svc, follow)
			}(name)
		}
		wg.Wait()
		return nil
	},
}

func streamLogs(containerName, prefix string, follow bool) {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	args = append(args, containerName)

	c := exec.Command("container", args...)
	stdout, err := c.StdoutPipe()
	if err != nil {
		fmt.Printf("[%s] error: %v\n", prefix, err)
		return
	}
	c.Stderr = c.Stdout
	if err := c.Start(); err != nil {
		fmt.Printf("[%s] error: %v\n", prefix, err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Printf("[%s] %s\n", prefix, scanner.Text())
	}
	_ = c.Wait()
}

func init() {
	logsCmd.Flags().BoolVar(&follow, "follow", false, "Follow log output")
}
