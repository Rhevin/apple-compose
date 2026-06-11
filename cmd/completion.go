package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for apple-compose.

To load completions in zsh:

  source <(apple-compose completion zsh)

To load completions for each session, add the above line to ~/.zshrc.`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func completeServiceNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch cmd.Name() {
	case "exec", "run":
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	case "port":
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	project, err := loadProject()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names := make([]string, 0, len(project.Services))
	for name := range project.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func completeComposeFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if toComplete == "" {
		toComplete = "."
	}
	matches, err := filepath.Glob(toComplete + "*")
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	var out []string
	for _, m := range matches {
		if strings.HasSuffix(m, ".yml") || strings.HasSuffix(m, ".yaml") || strings.HasSuffix(m, "/") {
			out = append(out, m)
		}
	}
	sort.Strings(out)
	return out, cobra.ShellCompDirectiveDefault
}

func registerCompletions() {
	serviceCmds := []*cobra.Command{
		upCmd, downCmd, logsCmd, pullCmd, execCmd, runCmd,
		stopCmd, startCmd, restartCmd, killCmd, rmCmd,
		topCmd, portCmd, statsCmd, imagesCmd,
	}
	for _, c := range serviceCmds {
		c.ValidArgsFunction = completeServiceNames
	}

	_ = rootCmd.RegisterFlagCompletionFunc("file", completeComposeFiles)
}
