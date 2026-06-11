package cmd

import (
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/rhevin/apple-compose/internal/compose"
	"github.com/spf13/cobra"
)

var (
	composeFiles []string
	projectName  string
	envFile      string
	profiles     []string
)

var rootCmd = &cobra.Command{
	Use:   "apple-compose",
	Short: "Docker Compose-compatible orchestrator for Apple Containers",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return
		}
		backend.WarnContainerCLIVersion()
	},
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

// composeFilesOrDefault returns explicit -f paths, or docker-compose.yml when none were given.
func composeFilesOrDefault() []string {
	if len(composeFiles) > 0 {
		return composeFiles
	}
	return []string{"docker-compose.yml"}
}

// loadProject loads the compose file honoring all global flags.
func loadProject() (*types.Project, error) {
	return compose.Load(composeFilesOrDefault(), compose.Options{
		ProjectName: projectName,
		EnvFile:     envFile,
		Profiles:    profiles,
	})
}

func init() {
	rootCmd.PersistentFlags().StringArrayVarP(&composeFiles, "file", "f", nil, "Compose file(s) to use")
	rootCmd.PersistentFlags().StringVarP(&projectName, "project-name", "p", "", "Project name (overrides compose file name)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env-file", "", "Path to an env file")
	rootCmd.PersistentFlags().StringArrayVar(&profiles, "profile", nil, "Enable service profiles")

	rootCmd.AddCommand(
		completionCmd,
		upCmd, downCmd, psCmd, logsCmd,
		pullCmd, execCmd, stopCmd, startCmd, restartCmd,
		killCmd, rmCmd, runCmd,
		configCmd, lsCmd, topCmd, portCmd, statsCmd, imagesCmd, cpCmd,
		loginCmd, logoutCmd,
		pauseCmd, unpauseCmd, eventsCmd, waitCmd, scaleCmd,
		pruneCmd,
	)

	registerCompletions()
}
