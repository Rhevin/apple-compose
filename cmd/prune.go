package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var (
	pruneImages  bool
	pruneAll     bool
	pruneNetwork bool
	pruneVolumes bool
	pruneForce   bool
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stopped containers, unused images, and networks",
	Long: `Remove stopped containers and optionally unused images and networks.

By default only removes stopped containers. Use flags to also prune
images and networks. Use --volumes to delete named volume data for this project.

  # equivalent of: docker system prune -f -a
  apple-compose prune --force -a --volumes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// -a implies --networks (matches docker system prune -a behaviour)
		if pruneAll {
			pruneNetwork = true
		}

		if !pruneForce {
			fmt.Println("This will remove:")
			fmt.Println("  - all stopped containers")
			if pruneNetwork {
				fmt.Println("  - all unused networks")
			}
			if pruneAll {
				fmt.Println("  - all unused images")
			} else if pruneImages {
				fmt.Println("  - dangling images")
			}
			if pruneVolumes {
				fmt.Printf("  - named volume data for project %q\n", resolveProjectName())
			}
			fmt.Print("\nContinue? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Containers
		fmt.Println("Pruning stopped containers...")
		runPruneCmd("container", "prune")

		// Networks
		if pruneNetwork {
			fmt.Println("Pruning unused networks...")
			runPruneCmd("container", "network", "prune")
		}

		// Images
		if pruneImages || pruneAll {
			imgArgs := []string{"image", "prune"}
			if pruneAll {
				imgArgs = append(imgArgs, "--all")
			}
			fmt.Println("Pruning unused images...")
			runPruneCmd("container", imgArgs...)
		}

		// Named volumes
		if pruneVolumes {
			proj := resolveProjectName()
			fmt.Printf("Removing named volumes for project %q...\n", proj)
			if err := backend.RemoveNamedVolumes(proj); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: %v\n", err)
			}
		}

		fmt.Println("Done.")
		return nil
	},
}

func runPruneCmd(name string, args ...string) {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "  warning: %v\n", err)
	}
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneForce, "force", false, "Skip confirmation prompt")
	pruneCmd.Flags().BoolVar(&pruneImages, "images", false, "Remove dangling images")
	pruneCmd.Flags().BoolVarP(&pruneAll, "all", "a", false, "Remove all unused images (also implies --networks)")
	pruneCmd.Flags().BoolVar(&pruneNetwork, "networks", false, "Remove unused networks")
	pruneCmd.Flags().BoolVar(&pruneVolumes, "volumes", false, "Remove named volume data for this project")
}
