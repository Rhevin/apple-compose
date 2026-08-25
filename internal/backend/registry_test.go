package backend

import (
	"strings"
	"testing"
)

func TestValidateRegistryScheme(t *testing.T) {
	for _, scheme := range []string{"", "http", "https"} {
		if err := ValidateRegistryScheme(scheme); err != nil {
			t.Errorf("scheme %q: unexpected error: %v", scheme, err)
		}
	}
	if err := ValidateRegistryScheme("auto"); err == nil || !strings.Contains(err.Error(), "1.3.0") {
		t.Fatalf("expected 1.3.0 auto rejection, got %v", err)
	}
	if err := ValidateRegistryScheme("ftp"); err == nil {
		t.Fatal("expected error for unknown scheme")
	}
}

func TestImagePullArgs(t *testing.T) {
	args, err := ImagePullArgs("nginx:alpine", "linux/amd64", "http")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"image", "pull", "--platform", "linux/amd64", "--scheme", "http", "nginx:alpine"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", args, want)
	}
}

func TestImagePullArgs_Defaults(t *testing.T) {
	args, err := ImagePullArgs("nginx:alpine", "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"image", "pull", "nginx:alpine"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", args, want)
	}
}

func TestRegistryLoginArgs(t *testing.T) {
	args, err := RegistryLoginArgs("localhost:5000", "user", "http", true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"registry", "login", "--username", "user", "--password-stdin", "--scheme", "http", "localhost:5000"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", args, want)
	}
}
