package cli

import (
	"fmt"

	m "github.com/kubex-ecosystem/gastype/internal/manager"
	l "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

// TypeCheckCmds Define the root command
func TypeCheckCmds() []*cobra.Command {
	return []*cobra.Command{
		commandCheckType(),
		commandWatch(),
	}
}

func commandCheckType() *cobra.Command {
	var dir, outputFile, configFile string
	var workerCount int

	checkCmd := &cobra.Command{
		Use: "check",
		Annotations: GetDescriptions([]string{
			"Check code files for type errors in a given directory",
			"Check code files for type errors",
		}, false),
		Example: `gastype check -d ./example -w 4 -o type_check_results.json`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create a new configuration
			// Create a new type manager
			tc := m.NewTypeManager(nil)

			// Load the actions
			if prepareErr := tc.PrepareActions(); prepareErr != nil {
				l.Error(fmt.Sprintf("Error preparing actions: %s", prepareErr.Error()), nil)
				return
			}

			// Start checking the Go files
			if err := tc.StartChecking(workerCount); err != nil {
				l.Error(fmt.Sprintf("Error checking Go files: %s", err.Error()), nil)
				return
			}

		},
	}

	// Add flags to the root command
	checkCmd.Flags().StringVarP(&dir, "dir", "d", "./", "Directory containing Go files")
	checkCmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "type_check_results.json", "Output file for JSON results")
	checkCmd.Flags().StringVarP(&configFile, "config", "c", "config.json", "Configuration file for email notifications")

	return checkCmd
}

func commandWatch() *cobra.Command {
	var dir, outputFile string
	var workerCount int
	var email, emailToken, configFile string
	var notify bool

	watch := &cobra.Command{
		Use: "watch",
		Annotations: GetDescriptions([]string{
			"Watcher and notifier for type checking Go files in a given directory",
			"Watcher and notifier for type checking Go files",
		}, false),
		Example: `gastype watch -d ./example -w 4 -o type_check_results.json`,
		Run: func(cmd *cobra.Command, args []string) {

			// Create a new type manager
			tc := m.NewTypeManager(nil)

			// Set the email notifications
			tc.SetEmail(email)
			tc.SetEmailToken(emailToken)
			tc.SetNotify(notify)

			// Start checking the Go files
			if err := tc.StartChecking(workerCount); err != nil {
				l.Error(fmt.Sprintf("Error checking Go files: %s", err.Error()), nil)
				return
			}

		},
	}

	// Add flags to the watch command
	watch.Flags().StringVarP(&email, "email", "e", "gastype@gmail.com", "Email address for notifications")
	watch.Flags().StringVarP(&emailToken, "token", "t", "123456", "Token for email notifications")
	watch.Flags().BoolVarP(&notify, "notify", "n", false, "Enable email notifications")
	watch.Flags().StringVarP(&dir, "dir", "d", "./example", "Directory containing Go files")
	watch.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	watch.Flags().StringVarP(&outputFile, "output", "o", "type_check_results.json", "Output file for JSON results")
	watch.Flags().StringVarP(&configFile, "config", "c", "config.json", "Configuration file for email notifications")

	return watch
}
