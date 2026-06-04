package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	loginUsername      string
	loginPasswordStdin bool
)

var loginCmd = &cobra.Command{
	Use:   "login [registry]",
	Short: "Log in to a container registry",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cArgs := []string{"registry", "login"}
		if loginUsername != "" {
			cArgs = append(cArgs, "--username", loginUsername)
		}
		if loginPasswordStdin {
			cArgs = append(cArgs, "--password-stdin")
		}
		if len(args) == 1 {
			cArgs = append(cArgs, args[0])
		}
		c := exec.Command("container", cArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout [registry]",
	Short: "Log out from a container registry",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify a registry to log out from")
		}
		c := exec.Command("container", "registry", "logout", args[0])
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username")
	loginCmd.Flags().BoolVar(&loginPasswordStdin, "password-stdin", false, "Read password from stdin")
}
