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

// Load parses one or more Compose files (merged left-to-right) and returns the project.
func Load(files []string, opts ...Options) (*types.Project, error) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	if len(files) == 0 {
		files = []string{"docker-compose.yml"}
	}

	absFiles := make([]string, len(files))
	for i, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		absFiles[i] = abs
	}

	name := o.ProjectName
	if name == "" {
		name = projectName(absFiles[0])
	}

	popts := []cli.ProjectOptionsFn{
		cli.WithWorkingDirectory(filepath.Dir(absFiles[0])),
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

	projectOpts, err := cli.NewProjectOptions(absFiles, popts...)
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
