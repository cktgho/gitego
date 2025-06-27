// cmd/completion.go

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const completionLongDesc = `To load completions:

Bash:
  $ source <(gitego completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ gitego completion bash > /etc/bash_completion.d/gitego
  # macOS:
  $ gitego completion bash > /usr/local/etc/bash_completion.d/gitego

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ gitego completion zsh > "${fpath[1]}/_gitego"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ gitego completion fish | source

  # To load completions for each session, execute once:
  $ gitego completion fish > ~/.config/fish/completions/gitego.fish

PowerShell:
  PS> gitego.exe completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> gitego.exe completion powershell > gitego.ps1
  # and source this file from your PowerShell profile.
`

// completionCmd represents the completion command.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script for your shell",
	Long:  completionLongDesc,
	// Disables file completion for this command
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	//Args:                  cobra.ExactValidArgs(1),
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
