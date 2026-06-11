package backend

import (
	"context"
	"testing"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
)

func TestComposeLoad_DeployLimitsNotFlattened(t *testing.T) {
	t.Parallel()

	p, err := loader.LoadWithContext(context.Background(), types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{{
			Filename: "compose.yaml",
			Content: []byte(`name: test
services:
  web:
    image: nginx:alpine
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: 512M
`),
		}},
	}, func(o *loader.Options) { o.SetProjectName("test", true) })
	if err != nil {
		t.Fatal(err)
	}

	svc := p.Services["web"]
	if svc.MemLimit != 0 || svc.CPUS != 0 {
		t.Fatalf("compose-go does not flatten deploy limits into top-level fields: mem_limit=%d cpus=%f", svc.MemLimit, svc.CPUS)
	}
	if svc.Deploy == nil || svc.Deploy.Resources.Limits == nil {
		t.Fatal("expected deploy.resources.limits to be populated")
	}
	limits := svc.Deploy.Resources.Limits
	if limits.MemoryBytes != 512*1024*1024 {
		t.Fatalf("memory limit: got %d", limits.MemoryBytes)
	}
	if limits.NanoCPUs.Value() != 2 {
		t.Fatalf("cpu limit: got %f", limits.NanoCPUs.Value())
	}

	args, err := RunArgs("test", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--memory", "536870912")
	assertContains(t, args, "--cpus", "2.00")
}
