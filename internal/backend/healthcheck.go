package backend

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
)

// HealthcheckExecArgs returns the command to run inside a container for a healthcheck test.
func HealthcheckExecArgs(hc *types.HealthCheckConfig) ([]string, error) {
	if hc == nil || hc.Disable || len(hc.Test) == 0 {
		return nil, fmt.Errorf("no healthcheck defined")
	}
	test := hc.Test
	switch test[0] {
	case "NONE":
		return nil, fmt.Errorf("healthcheck disabled")
	case "CMD":
		if len(test) < 2 {
			return nil, fmt.Errorf("invalid healthcheck CMD")
		}
		return test[1:], nil
	case "CMD-SHELL":
		if len(test) < 2 {
			return nil, fmt.Errorf("invalid healthcheck CMD-SHELL")
		}
		return []string{"sh", "-c", test[1]}, nil
	default:
		return test, nil
	}
}

// WaitForDependency blocks until a dependency satisfies its depends_on condition.
func WaitForDependency(project string, dep types.ServiceConfig, condition string, timeout time.Duration) error {
	name := ContainerName(project, dep.Name)

	switch condition {
	case "", types.ServiceConditionStarted:
		return waitRunning(name, dep.Name, timeout)
	case types.ServiceConditionHealthy:
		if dep.HealthCheck == nil || dep.HealthCheck.Disable || len(dep.HealthCheck.Test) == 0 {
			fmt.Fprintf(os.Stderr,
				"  WARNING: service %q has depends_on condition service_healthy but no healthcheck; waiting for running state\n",
				dep.Name,
			)
			return waitRunning(name, dep.Name, timeout)
		}
		return waitHealthcheck(name, dep.Name, dep.HealthCheck, timeout)
	case types.ServiceConditionCompletedSuccessfully:
		return waitCompleted(name, dep.Name, timeout)
	default:
		fmt.Fprintf(os.Stderr,
			"  WARNING: unknown depends_on condition %q for %q; treating as service_started\n",
			condition, dep.Name,
		)
		return waitRunning(name, dep.Name, timeout)
	}
}

func waitRunning(name, service string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := containerStatus(name)
		if err == nil && status.State == "running" {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service %q did not start within %s", service, timeout)
}

func waitHealthcheck(name, service string, hc *types.HealthCheckConfig, timeout time.Duration) error {
	interval := healthcheckInterval(hc)
	startPeriod := healthcheckStartPeriod(hc)
	deadline := time.Now().Add(timeout)
	started := time.Now()

	for time.Now().Before(deadline) {
		status, err := containerStatus(name)
		if err != nil || status.State != "running" {
			time.Sleep(interval)
			continue
		}
		if time.Since(started) < startPeriod {
			time.Sleep(interval)
			continue
		}
		if err := runHealthcheck(name, hc); err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("service %q did not become healthy within %s", service, timeout)
}

func waitCompleted(name, service string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := containerStatus(name)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		switch status.State {
		case "running":
			time.Sleep(500 * time.Millisecond)
		default:
			return nil
		}
	}
	return fmt.Errorf("service %q did not complete within %s", service, timeout)
}

func runHealthcheck(containerName string, hc *types.HealthCheckConfig) error {
	args, err := HealthcheckExecArgs(hc)
	if err != nil {
		return err
	}
	execArgs := append([]string{"exec", containerName}, args...)
	cmd := exec.Command(bin, execArgs...)
	return cmd.Run()
}

func healthcheckInterval(hc *types.HealthCheckConfig) time.Duration {
	if hc.Interval != nil {
		return time.Duration(*hc.Interval)
	}
	return 30 * time.Second
}

func healthcheckStartPeriod(hc *types.HealthCheckConfig) time.Duration {
	if hc.StartPeriod != nil {
		return time.Duration(*hc.StartPeriod)
	}
	return 0
}
