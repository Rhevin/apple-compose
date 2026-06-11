package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

const LabelConfigHash = "com.apple-compose.config-hash"

// serviceConfigHash fingerprints all compose fields that affect container run args.
func serviceConfigHash(project string, svc types.ServiceConfig) string {
	h := sha256.New()
	hashWrite(h, svc.Image)

	for _, p := range composePortKeys(svc.Ports) {
		hashWrite(h, "port:"+p)
	}
	for _, kv := range sortedEnvPairs(svc.Environment) {
		hashWrite(h, "env:"+kv)
	}
	for _, path := range sortedEnvFilePaths(svc.EnvFiles) {
		hashWrite(h, "envfile:"+path)
	}
	if svc.Entrypoint != nil {
		hashWrite(h, "entrypoint:"+strings.Join(svc.Entrypoint, " "))
	}
	hashWrite(h, "user:"+svc.User)
	hashWrite(h, "workdir:"+svc.WorkingDir)
	for _, c := range sortedStrings(svc.CapAdd) {
		hashWrite(h, "capadd:"+c)
	}
	for _, c := range sortedStrings(svc.CapDrop) {
		hashWrite(h, "capdrop:"+c)
	}
	for _, t := range sortedStrings(svc.Tmpfs) {
		hashWrite(h, "tmpfs:"+t)
	}
	if svc.ReadOnly {
		hashWrite(h, "readonly:1")
	}
	for _, u := range sortedUlimitKeys(svc.Ulimits) {
		hashWrite(h, "ulimit:"+u)
	}
	if svc.Init != nil && *svc.Init {
		hashWrite(h, "init:1")
	}
	for _, v := range sortedVolumeKeys(project, svc.Volumes) {
		hashWrite(h, "vol:"+v)
	}
	mem, cpus := serviceResourceLimits(svc)
	if mem > 0 {
		hashWrite(h, fmt.Sprintf("mem:%d", mem))
	}
	if cpus > 0 {
		hashWrite(h, fmt.Sprintf("cpus:%.4f", cpus))
	}
	if svc.ShmSize > 0 {
		hashWrite(h, fmt.Sprintf("shm:%d", int64(svc.ShmSize)))
	}
	for _, c := range svc.Command {
		hashWrite(h, "cmd:"+c)
	}
	for _, line := range sortedExtraHostLines(svc.ExtraHosts) {
		hashWrite(h, "extrahost:"+line)
	}
	return hex.EncodeToString(h.Sum(nil)[:8])
}

func hashWrite(h interface{ Write([]byte) (int, error) }, s string) {
	_, _ = h.Write([]byte(s))
	_, _ = h.Write([]byte{0})
}

func sortedEnvFilePaths(files []types.EnvFile) []string {
	paths := make([]string, 0, len(files))
	for _, f := range files {
		if f.Path != "" {
			paths = append(paths, f.Path)
		}
	}
	sort.Strings(paths)
	return paths
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func sortedUlimitKeys(limits map[string]*types.UlimitsConfig) []string {
	keys := make([]string, 0, len(limits))
	for name, limit := range limits {
		if limit == nil {
			continue
		}
		var val string
		if limit.Single != 0 {
			val = fmt.Sprintf("%d", limit.Single)
		} else {
			val = fmt.Sprintf("%d:%d", limit.Soft, limit.Hard)
		}
		keys = append(keys, fmt.Sprintf("%s=%s", name, val))
	}
	sort.Strings(keys)
	return keys
}

func sortedVolumeKeys(project string, vols []types.ServiceVolumeConfig) []string {
	keys := make([]string, 0, len(vols))
	for _, v := range vols {
		switch v.Type {
		case "bind":
			keys = append(keys, fmt.Sprintf("bind:%s:%s", v.Source, v.Target))
		case "volume", "":
			if v.Source != "" {
				keys = append(keys, fmt.Sprintf("named:%s:%s:%s", project, v.Source, v.Target))
			}
		default:
			keys = append(keys, fmt.Sprintf("%s:%s:%s", v.Type, v.Source, v.Target))
		}
	}
	sort.Strings(keys)
	return keys
}

func sortedExtraHostLines(hosts types.HostsList) []string {
	names := make([]string, 0, len(hosts))
	for name := range hosts {
		names = append(names, name)
	}
	sort.Strings(names)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		ips := append([]string(nil), hosts[name]...)
		sort.Strings(ips)
		lines = append(lines, fmt.Sprintf("%s=%s", name, strings.Join(ips, ",")))
	}
	return lines
}

func serviceResourceLimits(svc types.ServiceConfig) (mem types.UnitBytes, cpus float32) {
	mem = svc.MemLimit
	cpus = svc.CPUS
	if svc.Deploy != nil && svc.Deploy.Resources.Limits != nil {
		limits := svc.Deploy.Resources.Limits
		if mem == 0 {
			mem = limits.MemoryBytes
		}
		if cpus == 0 && limits.NanoCPUs > 0 {
			cpus = limits.NanoCPUs.Value()
		}
	}
	return mem, cpus
}

func sortedEnvPairs(env types.MappingWithEquals) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		val := ""
		if v := env[k]; v != nil {
			val = *v
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, val))
	}
	return pairs
}

func composePortKeys(ports []types.ServicePortConfig) []string {
	keys := make([]string, 0, len(ports))
	for _, p := range ports {
		keys = append(keys, fmt.Sprintf("%s:%d", p.Published, p.Target))
	}
	sort.Strings(keys)
	return keys
}

func containerPortKeys(ports []publishedPort) []string {
	keys := make([]string, 0, len(ports))
	for _, p := range ports {
		host := p.HostPort
		if host == 0 {
			continue
		}
		keys = append(keys, fmt.Sprintf("%d:%d", host, p.ContainerPort))
	}
	sort.Strings(keys)
	return keys
}

// configChanged reports whether an existing container no longer matches the compose service.
func configChanged(project string, c appleContainer, svc types.ServiceConfig) bool {
	if hash := c.Configuration.Labels[LabelConfigHash]; hash != "" {
		return hash != serviceConfigHash(project, svc)
	}
	// Legacy containers created before config-hash labels: compare image and ports.
	if !imageMatches(c.Configuration.Image.Reference, svc.Image) {
		return true
	}
	return !portSetsEqual(composePortKeys(svc.Ports), containerPortKeys(c.Configuration.PublishedPorts))
}

func imageMatches(running, desired string) bool {
	if running == desired {
		return true
	}
	if strings.HasPrefix(running, desired+"@") {
		return true
	}
	return false
}

func portSetsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
