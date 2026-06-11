package compose

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleCompose = `
name: testproject

services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: app
    volumes:
      - ./data:/var/lib/postgresql/data

  web:
    image: nginx:alpine
    ports:
      - "8080:80"
    depends_on:
      - db
`

func writeTempCompose(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_ParsesServices(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	project, err := Load([]string{path})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(project.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(project.Services))
	}

	db, err := project.GetService("db")
	if err != nil {
		t.Fatal("service 'db' not found")
	}
	if db.Image != "postgres:16-alpine" {
		t.Errorf("db image: expected postgres:16-alpine, got %s", db.Image)
	}

	web, err := project.GetService("web")
	if err != nil {
		t.Fatal("service 'web' not found")
	}
	if web.Image != "nginx:alpine" {
		t.Errorf("web image: expected nginx:alpine, got %s", web.Image)
	}
}

func TestLoad_DependsOn(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	project, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}

	web, _ := project.GetService("web")
	if _, ok := web.DependsOn["db"]; !ok {
		t.Error("expected web to depend on db")
	}
}

func TestLoad_EnvVars(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	project, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}

	db, _ := project.GetService("db")
	if v, ok := db.Environment["POSTGRES_PASSWORD"]; !ok || *v != "secret" {
		t.Errorf("expected POSTGRES_PASSWORD=secret, got %v", v)
	}
}

func TestLoad_Ports(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	project, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}

	web, _ := project.GetService("web")
	if len(web.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(web.Ports))
	}
	p := web.Ports[0]
	if p.Published != "8080" || p.Target != 80 {
		t.Errorf("expected 8080:80, got %s:%d", p.Published, p.Target)
	}
}

func TestLoad_ProjectName(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	project, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	// compose-go uses the name: field from the file; falls back to directory name.
	// Both are acceptable — just ensure a non-empty name is set.
	if project.Name == "" {
		t.Error("expected non-empty project name")
	}
}

func TestLoad_ProjectName_EnvOverride(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	t.Setenv("COMPOSE_PROJECT_NAME", "myoverride")
	project, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if project.Name != "myoverride" {
		t.Errorf("expected COMPOSE_PROJECT_NAME override 'myoverride', got %q", project.Name)
	}
}

func TestLoad_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "compose.yml")
	override := filepath.Join(dir, "compose.override.yml")
	if err := os.WriteFile(base, []byte(`
services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(override, []byte(`
services:
  web:
    image: haproxy:latest
    ports:
      - "9090:80"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := Load([]string{base, override})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	web, err := project.GetService("web")
	if err != nil {
		t.Fatal("service 'web' not found")
	}
	if web.Image != "haproxy:latest" {
		t.Errorf("expected override image haproxy:latest, got %s", web.Image)
	}
	// compose-go merges list fields; the override port should be present.
	found := false
	for _, p := range web.Ports {
		if p.Published == "9090" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected override port 9090 in %+v", web.Ports)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load([]string{"/nonexistent/docker-compose.yml"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_OrderMatchesDependencies(t *testing.T) {
	path := writeTempCompose(t, sampleCompose)
	project, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}

	order, err := TopologicalOrder(project)
	if err != nil {
		t.Fatal(err)
	}

	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if pos["db"] >= pos["web"] {
		t.Errorf("db (pos %d) must come before web (pos %d)", pos["db"], pos["web"])
	}
}
