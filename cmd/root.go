package cmd

import (
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rhevin/apple-compose/internal/compose"
	"github.com/spf13/cobra"
)

var (
	composeFile string
	projectName string
	envFile     string
	profiles    []string
)

var rootCmd = &cobra.Command{
	Use:   "apple-compose",
	Short: "Docker Compose-compatible orchestrator for Apple Containers",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func SetVersion(v string) {
	rootCmd.Version = v
}

// loadProject loads the compose file honoring all global flags.
func loadProject() (*types.Project, error) {
	return compose.Load(composeFile, compose.Options{
		ProjectName: projectName,
		EnvFile:     envFile,
		Profiles:    profiles,
	})
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&composeFile, "file", "f", "docker-compose.yml", "Compose file to use")
	rootCmd.PersistentFlags().StringVarP(&projectName, "project-name", "p", "", "Project name (overrides compose file name)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env-file", "", "Path to an env file")
	rootCmd.PersistentFlags().StringArrayVar(&profiles, "profile", nil, "Enable service profiles")

	rootCmd.AddCommand(
		upCmd, downCmd, psCmd, logsCmd,
		pullCmd, execCmd, stopCmd, startCmd, restartCmd,
		killCmd, rmCmd, runCmd,
		configCmd, lsCmd, topCmd, portCmd, statsCmd, imagesCmd, cpCmd,
		loginCmd, logoutCmd,
		pauseCmd, unpauseCmd, eventsCmd, waitCmd, scaleCmd,
		pruneCmd,
	)
}
