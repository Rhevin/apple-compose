package backend

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

// UnsupportedServiceKeys returns compose keys set on a service that apple-compose
// does not implement (or only partially implements).
func UnsupportedServiceKeys(svc types.ServiceConfig) []string {
	var keys []string
	add := func(key string, set bool) {
		if set {
			keys = append(keys, key)
		}
	}

	add("annotations", len(svc.Annotations) > 0)
	add("attach", svc.Attach != nil)
	add("blkio_config", svc.BlkioConfig != nil)
	add("build", svc.Build != nil)
	add("cgroup", svc.Cgroup != "")
	add("cgroup_parent", svc.CgroupParent != "")
	add("configs", len(svc.Configs) > 0)
	add("container_name", svc.ContainerName != "")
	add("cpu_count", svc.CPUCount != 0)
	add("cpu_percent", svc.CPUPercent != 0)
	add("cpu_period", svc.CPUPeriod != 0)
	add("cpu_quota", svc.CPUQuota != 0)
	add("cpu_rt_period", svc.CPURTPeriod != 0)
	add("cpu_rt_runtime", svc.CPURTRuntime != 0)
	add("cpuset", svc.CPUSet != "")
	add("cpu_shares", svc.CPUShares != 0)
	add("credential_spec", svc.CredentialSpec != nil)
	add("develop", svc.Develop != nil)
	add("device_cgroup_rules", len(svc.DeviceCgroupRules) > 0)
	add("devices", len(svc.Devices) > 0)
	add("dns", len(svc.DNS) > 0)
	add("dns_opt", len(svc.DNSOpts) > 0)
	add("dns_search", len(svc.DNSSearch) > 0)
	add("domainname", svc.DomainName != "")
	add("external_links", len(svc.ExternalLinks) > 0)
	add("gpus", len(svc.Gpus) > 0)
	add("group_add", len(svc.GroupAdd) > 0)
	add("hostname", svc.Hostname != "")
	add("ipc", svc.Ipc != "")
	add("isolation", svc.Isolation != "")
	add("labels", len(svc.Labels) > 0)
	add("links", len(svc.Links) > 0)
	add("logging", svc.Logging != nil)
	add("log_driver", svc.LogDriver != "")
	add("log_opt", len(svc.LogOpt) > 0)
	add("mac_address", svc.MacAddress != "")
	add("mem_reservation", svc.MemReservation > 0)
	add("memswap_limit", svc.MemSwapLimit > 0)
	add("mem_swappiness", svc.MemSwappiness > 0)
	add("models", len(svc.Models) > 0)
	add("net", svc.Net != "")
	add("network_mode", svc.NetworkMode != "")
	add("networks", hasCustomNetworks(svc.Networks))
	add("restart", svc.Restart != "" && svc.Restart != "no")
	add("pid", svc.Pid != "")
	add("pids_limit", svc.PidsLimit != 0)
	add("platform", svc.Platform != "")
	add("privileged", svc.Privileged)
	add("provider", svc.Provider != nil)
	add("secrets", len(svc.Secrets) > 0)
	add("security_opt", len(svc.SecurityOpt) > 0)
	add("stdin_open", svc.StdinOpen)
	add("sysctls", len(svc.Sysctls) > 0)
	add("tty", svc.Tty)
	add("expose", len(svc.Expose) > 0)

	for _, v := range svc.Volumes {
		switch v.Type {
		case "bind", "volume", "":
		default:
			add("volumes."+v.Type, true)
		}
	}

	keys = append(keys, unsupportedDeployKeys(svc.Deploy)...)

	sort.Strings(keys)
	return keys
}

func unsupportedDeployKeys(d *types.DeployConfig) []string {
	if d == nil {
		return nil
	}
	var keys []string
	add := func(key string, set bool) {
		if set {
			keys = append(keys, key)
		}
	}
	add("deploy.mode", d.Mode != "")
	add("deploy.replicas", d.Replicas != nil)
	add("deploy.labels", len(d.Labels) > 0)
	add("deploy.update_config", d.UpdateConfig != nil)
	add("deploy.rollback_config", d.RollbackConfig != nil)
	add("deploy.restart_policy", d.RestartPolicy != nil)
	add("deploy.placement", !placementEmpty(d.Placement))
	add("deploy.endpoint_mode", d.EndpointMode != "")
	add("deploy.resources.reservations", d.Resources.Reservations != nil)
	return keys
}

// hasCustomNetworks reports whether the service attaches to non-default networks.
// compose-go always injects a "default" network entry; that alone is not unsupported.
func hasCustomNetworks(networks map[string]*types.ServiceNetworkConfig) bool {
	if len(networks) == 0 {
		return false
	}
	if len(networks) == 1 {
		if _, ok := networks["default"]; ok {
			return false
		}
	}
	return true
}

func placementEmpty(p types.Placement) bool {
	return len(p.Constraints) == 0 && len(p.Preferences) == 0 && p.MaxReplicas == 0
}

// WarnUnsupportedKeys prints a stderr warning for each unsupported key found in the project.
func WarnUnsupportedKeys(w io.Writer, project *types.Project) {
	if w == nil || project == nil {
		return
	}
	names := make([]string, 0, len(project.Services))
	for name := range project.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		keys := UnsupportedServiceKeys(project.Services[name])
		if len(keys) == 0 {
			continue
		}
		_, _ = fmt.Fprintf(w, "WARNING: service %q has unsupported key(s): %s\n",
			name, strings.Join(keys, ", "))
	}
}
