package backend

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestPrepareService_PostgresPGDATA(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:17-alpine",
		Volumes: []types.ServiceVolumeConfig{
			{Type: "volume", Source: "pgdata", Target: "/var/lib/postgresql/data"},
		},
	}
	got := PrepareService(svc)
	if got.Environment["PGDATA"] == nil || *got.Environment["PGDATA"] != "/tmp/pgdata" {
		t.Fatalf("expected PGDATA=/tmp/pgdata, got %v", got.Environment["PGDATA"])
	}
}

func TestPrepareService_PostgresPGDATAAlreadySet(t *testing.T) {
	custom := "/data"
	svc := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:16",
		Environment: types.MappingWithEquals{
			"PGDATA": &custom,
		},
		Volumes: []types.ServiceVolumeConfig{
			{Type: "volume", Source: "pgdata", Target: "/var/lib/postgresql/data"},
		},
	}
	got := PrepareService(svc)
	if *got.Environment["PGDATA"] != custom {
		t.Fatalf("should not override existing PGDATA")
	}
}

func TestPrepareService_MySQLWarns(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "db",
		Image: "mysql:8",
		Volumes: []types.ServiceVolumeConfig{
			{Type: "volume", Source: "data", Target: "/var/lib/mysql"},
		},
	}
	got := PrepareService(svc)
	if _, ok := got.Environment["PGDATA"]; ok {
		t.Fatal("mysql should not get PGDATA")
	}
}

func TestIsPostgresImage(t *testing.T) {
	if !isPostgresImage("docker.io/library/postgres:17") {
		t.Fatal("expected postgres match")
	}
	if isPostgresImage("nginx:alpine") {
		t.Fatal("unexpected postgres match")
	}
}

func TestHasNamedDataVolume(t *testing.T) {
	svc := types.ServiceConfig{
		Volumes: []types.ServiceVolumeConfig{
			{Type: "bind", Source: "/x", Target: "/var/lib/postgresql/data"},
			{Type: "volume", Source: "v", Target: "/var/lib/postgresql/data"},
		},
	}
	if !hasNamedDataVolume(svc, postgresDataDirs) {
		t.Fatal("expected named volume match")
	}
	if hasNamedDataVolume(types.ServiceConfig{
		Volumes: []types.ServiceVolumeConfig{
			{Type: "bind", Source: "/x", Target: "/var/lib/postgresql/data"},
		},
	}, postgresDataDirs) {
		t.Fatal("bind mount should not count")
	}
}

func TestPrepareService_NoVolumeNoOp(t *testing.T) {
	svc := types.ServiceConfig{Name: "db", Image: "postgres:16"}
	got := PrepareService(svc)
	if len(got.Environment) != 0 {
		t.Fatalf("unexpected env: %v", got.Environment)
	}
}

func TestIsMySQLImage(t *testing.T) {
	for _, img := range []string{"mysql:8", "mariadb:11", "docker.io/mariadb"} {
		if !isMySQLImage(img) {
			t.Fatalf("expected mysql family match for %q", img)
		}
	}
	if isMySQLImage("postgres:16") {
		t.Fatal("postgres is not mysql")
	}
}
