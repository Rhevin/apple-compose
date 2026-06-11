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

// serviceConfigHash fingerprints image, ports, and environment for drift detection.
func serviceConfigHash(svc types.ServiceConfig) string {
	h := sha256.New()
	h.Write([]byte(svc.Image))
	h.Write([]byte{0})
	for _, p := range composePortKeys(svc.Ports) {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	for _, kv := range sortedEnvPairs(svc.Environment) {
		h.Write([]byte(kv))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)[:8])
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
func configChanged(c appleContainer, svc types.ServiceConfig) bool {
	if hash := c.Configuration.Labels[LabelConfigHash]; hash != "" {
		return hash != serviceConfigHash(svc)
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
	// Resolved image refs may include a digest suffix.
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
