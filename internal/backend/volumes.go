package backend

import (
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

var (
	postgresDataDirs = []string{"/var/lib/postgresql/data", "/var/lib/postgresql"}
	mysqlDataDirs    = []string{"/var/lib/mysql"}
)

// PrepareService applies runtime workarounds (e.g. virtiofs DB volumes) before RunArgs.
func PrepareService(svc types.ServiceConfig) types.ServiceConfig {
	applyDBVolumeWorkaround(&svc)
	return svc
}

func applyDBVolumeWorkaround(svc *types.ServiceConfig) {
	if isPostgresImage(svc.Image) && hasNamedDataVolume(*svc, postgresDataDirs) && !envIsSet(*svc, "PGDATA") {
		val := "/tmp/pgdata"
		if svc.Environment == nil {
			svc.Environment = types.MappingWithEquals{}
		}
		svc.Environment["PGDATA"] = &val
		fmt.Fprintf(os.Stderr,
			"  NOTE: service %q: auto-setting PGDATA=/tmp/pgdata (virtiofs named-volume workaround)\n",
			svc.Name,
		)
		return
	}
	if isMySQLImage(svc.Image) && hasNamedDataVolume(*svc, mysqlDataDirs) {
		fmt.Fprintf(os.Stderr,
			"  WARNING: service %q: named volume on /var/lib/mysql may fail (virtiofs chown).\n"+
				"           Workaround: remove the named volume and add tmpfs, e.g.:\n"+
				"             tmpfs:\n"+
				"               - /var/lib/mysql\n",
			svc.Name,
		)
	}
}

func isPostgresImage(image string) bool {
	return strings.Contains(strings.ToLower(image), "postgres")
}

func isMySQLImage(image string) bool {
	ref := strings.ToLower(image)
	return strings.Contains(ref, "mysql") || strings.Contains(ref, "mariadb")
}

func hasNamedDataVolume(svc types.ServiceConfig, dirs []string) bool {
	for _, v := range svc.Volumes {
		if v.Type == "bind" {
			continue
		}
		for _, d := range dirs {
			if v.Target == d || strings.HasPrefix(v.Target, d+"/") {
				return true
			}
		}
	}
	return false
}

func envIsSet(svc types.ServiceConfig, key string) bool {
	_, ok := svc.Environment[key]
	return ok
}
