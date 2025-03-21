package cli

import (
	g "github.com/faelmori/gastype/internal/globals"
	m "github.com/faelmori/gastype/internal/manager"
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
		Use:     "check",
		Short:   "Check code files for type errors",
		Long:    "Check code files for type errors in a given directory",
		Example: `gastype check -d ./example -w 4 -o type_check_results.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := g.NewConfig()

			// Set the configuration values
			if dir == "" {
				cfg.SetDir(dir)
			}
			if workerCount > 0 {
				cfg.SetWorkerLimit(workerCount)
			}
			if outputFile != "" {
				cfg.SetOutputFile(outputFile)
			}

			// Load the configuration
			if cfgErr := cfg.Load(); cfgErr != nil {
				return cfgErr
			}

			// Create a new type manager
			tc := m.NewTypeManager(cfg)

			// Start checking the Go files
			if err := tc.StartChecking(workerCount); err != nil {
				return err
			}

			return nil
		},
	}

	// Add flags to the root command
	checkCmd.Flags().StringVarP(&dir, "dir", "d", "./example", "Directory containing Go files")
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
		Use:     "watch",
		Short:   "Watcher and notifier for type checking Go files",
		Long:    "Watcher and notifier for type checking Go files in a given directory",
		Example: `gastype watch -d ./example -w 4 -o type_check_results.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := g.NewConfig()

			// Set the configuration values
			if dir == "" {
				cfg.SetDir(dir)
			}
			if workerCount > 0 {
				cfg.SetWorkerLimit(workerCount)
			}
			if outputFile != "" {
				cfg.SetOutputFile(outputFile)
			}

			// Load the configuration
			if cfgErr := cfg.Load(); cfgErr != nil {
				return cfgErr
			}

			// Create a new type manager
			tc := m.NewTypeManager(cfg)

			// Set the email notifications
			tc.SetEmail(email)
			tc.SetEmailToken(emailToken)
			tc.SetNotify(notify)

			// Start checking the Go files
			if err := tc.StartChecking(workerCount); err != nil {
				return err
			}

			return nil
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
