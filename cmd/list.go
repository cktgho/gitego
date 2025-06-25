// cmd/list.go

package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/bgreenwell/gitego/config" 
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists all saved user profiles.",
	Long:    `Reads the gitego configuration file and displays a table of all saved profiles, including their associated user name and email.`,
	Aliases: []string{"ls"}, // Users can run 'gitego ls' as a shortcut for 'gitego list'
	Run: func(cmd *cobra.Command, args []string) {
		// Load the existing configuration using our config package.
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		// Check if there are any profiles to display.
		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles found. Use 'gitego add <profile_name>' to create one.")
			return
		}

		// To ensure a consistent output order, we'll get the profile names (keys),
		// sort them alphabetically, and then print them.
		profileNames := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			profileNames = append(profileNames, name)
		}
		sort.Strings(profileNames)

		// Use Go's built-in tabwriter to create a neatly formatted table.
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		defer w.Flush() // The writer buffers output, so we must Flush() to write it out.

		// Print the table header.
		fmt.Fprintln(w, "PROFILE\tNAME\tEMAIL")
		fmt.Fprintln(w, "-------\t----\t-----")

		// Loop through our sorted list of profile names.
		for _, name := range profileNames {
			profile := cfg.Profiles[name]
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, profile.Name, profile.Email)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
