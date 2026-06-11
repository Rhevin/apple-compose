package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var versionCheckOnce sync.Once

// WarnContainerCLIVersion checks the installed container CLI once per process and
// prints warnings for unsupported versions or unexpected JSON output shapes.
func WarnContainerCLIVersion() {
	versionCheckOnce.Do(func() {
		warnContainerVersion()
		warnContainerJSONShape()
	})
}

func warnContainerVersion() {
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: could not run %q --version: %v\n", bin, err)
		return
	}
	text := strings.TrimSpace(string(out))
	major, _, _, ok := parseContainerVersion(text)
	if !ok {
		fmt.Fprintf(os.Stderr,
			"WARNING: could not parse %q version from %q — expected container CLI 1.0.0+\n",
			bin, text,
		)
		return
	}
	if major < 1 {
		fmt.Fprintf(os.Stderr,
			"WARNING: %q reports %q — apple-compose expects container CLI 1.0.0+ for JSON status output\n",
			bin, text,
		)
	}
}

var containerVersionRE = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

func parseContainerVersion(text string) (major, minor, patch int, ok bool) {
	m := containerVersionRE.FindStringSubmatch(text)
	if len(m) != 4 {
		return 0, 0, 0, false
	}
	major, _ = strconv.Atoi(m[1])
	minor, _ = strconv.Atoi(m[2])
	patch, _ = strconv.Atoi(m[3])
	return major, minor, patch, true
}

func warnContainerJSONShape() {
	out, err := exec.Command(bin, "list", "--all", "--format", "json").Output()
	if err != nil {
		return // runtime may not be started yet
	}
	out = []byte(strings.TrimSpace(string(out)))
	if len(out) == 0 || out[0] != '[' {
		return
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(out, &raw); err != nil || len(raw) == 0 {
		return
	}
	var probe struct {
		Status json.RawMessage `json:"status"`
	}
	if err := json.Unmarshal(raw[0], &probe); err != nil || len(probe.Status) == 0 {
		return
	}
	if probe.Status[0] == '"' {
		fmt.Fprintf(os.Stderr,
			"WARNING: %q list JSON uses legacy string status — upgrade to container CLI 1.0.0+ for reliable parsing\n",
			bin,
		)
	}
}
