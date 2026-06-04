package compose

import (
	"context"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

// Options controls how a compose file is loaded.
type Options struct {
	ProjectName string
	EnvFile     string
	Profiles    []string
}

// Load parses a Compose file and returns the project.
func Load(file string, opts ...Options) (*types.Project, error) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	name := o.ProjectName
	if name == "" {
		name = projectName(absFile)
	}

	popts := []cli.ProjectOptionsFn{
		cli.WithWorkingDirectory(filepath.Dir(absFile)),
		cli.WithName(name),
		cli.WithDotEnv,
		cli.WithOsEnv,
	}

	if o.EnvFile != "" {
		popts = append(popts, cli.WithEnvFiles(o.EnvFile))
	}

	if len(o.Profiles) > 0 {
		popts = append(popts, cli.WithProfiles(o.Profiles))
	}

	projectOpts, err := cli.NewProjectOptions([]string{absFile}, popts...)
	if err != nil {
		return nil, err
	}

	return projectOpts.LoadProject(context.Background())
}

func projectName(composeFile string) string {
	if name := os.Getenv("COMPOSE_PROJECT_NAME"); name != "" {
		return name
	}
	return filepath.Base(filepath.Dir(composeFile))
}
