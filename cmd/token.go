package cmd

import (
	"github.com/spf13/cobra"
)

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get authentication token for scripting",
	Long: `Get authentication token for use in CI/CD pipelines and scripts.

This is an alias for 'owlctl login --output-token'.

The token is printed to stdout for capture in environment variables or scripts.
This is useful for CI/CD workflows where you want to authenticate once and reuse
the token across multiple commands.

Examples:
  # Capture token in environment variable
  export OWLCTL_TOKEN=$(owlctl token)

  # Use token in subsequent commands
  owlctl get jobs
  owlctl job diff --all

  # Capture token in script
  TOKEN=$(owlctl token)
  echo "Token: $TOKEN"

Security Notes:
  - Tokens are short-lived (typically 15 minutes)
  - Store tokens in CI/CD secrets, not in code
  - Never commit tokens to version control
  - Tokens are only valid for the authenticated profile
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set outputToken flag to true
		outputToken = true
		// Call the login function
		loginWithTokenManager()
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.Flags().BoolVar(&debugAuth, "debug-auth", false, "Enable authentication debug logging")
}
