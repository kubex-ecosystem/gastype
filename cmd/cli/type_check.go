package cli

import (
	"fmt"
	g "github.com/faelmori/gastype/internal/globals"
	m "github.com/faelmori/gastype/internal/manager"
	"github.com/faelmori/kubex-interfaces/module"
	l "github.com/faelmori/logz"
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
			lgr := l.GetLogger("GasType")
			// Create a new configuration
			lgr.NoticeCtx("Creating configuration", nil)
			if cfg := g.NewConfigWithArgs(dir, workerCount, outputFile, lgr, module.RegX(
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				[]string{},
				true,
				cmd,
				nil,
			)); cfg == nil {
				lgr.ErrorCtx(fmt.Sprintf("Error creating configuration"), nil)
			} else {
				tc := m.NewTypeManager(cfg, lgr)

				tc.SetNotify(true)

				if len(tc.GetActions()) == 0 {
					if prepErr := tc.PrepareActions(); prepErr != nil {
						lgr.ErrorCtx(fmt.Sprintf("Error preparing actions: %s", prepErr.Error()), nil)
						return
					} else {
						if files, filesErr := tc.GetFilesList(true); filesErr != nil {
							lgr.ErrorCtx(fmt.Sprintf("Error getting files list: %s", filesErr.Error()), nil)
							return
						} else {
							lgr.NoticeCtx(fmt.Sprintf("Actions prepared successfully with %d", len(files)), nil)
							tc.SetFiles(files)
						}
					}
				}

				if err := tc.StartChecking(workerCount); err != nil {
					lgr.ErrorCtx(fmt.Sprintf("Error checking Go files: %s", err.Error()), nil)
					return
				}

				lgr.SuccessCtx("Type checking completed successfully", nil)
			}
		},
	}

	// Add flags to the root command
	checkCmd.Flags().StringVarP(&dir, "dir", "d", "./", "Directory containing Go files")
	checkCmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "./type_check_results.json", "Output file for JSON results")
	checkCmd.Flags().StringVarP(&configFile, "config", "c", "./config.json", "Configuration file for email notifications")

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
			lgr := l.GetLogger("GasType")
			if cfg := g.NewConfigWithArgs(dir, workerCount, outputFile, lgr, module.RegX(
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				"gastype",
				[]string{},
				true,
				cmd,
				nil,
			)); cfg == nil {
				lgr.ErrorCtx("Error creating configuration", nil)
				return
			} else {
				// Create a new type manager
				tc := m.NewTypeManager(cfg, lgr)

				// Set the email notifications
				tc.SetEmail(email)
				tc.SetEmailToken(emailToken)
				tc.SetNotify(notify)

				// Start checking the Go files
				if err := tc.StartChecking(workerCount); err != nil {
					lgr.ErrorCtx(fmt.Sprintf("Error checking Go files: %s", err.Error()), nil)
					return
				}
			}
		},
	}

	// Add flags to the watch command
	watch.Flags().StringVarP(&email, "email", "e", "gastype@gmail.com", "Email address for notifications")
	watch.Flags().StringVarP(&emailToken, "token", "t", "123456", "Token for email notifications")
	watch.Flags().BoolVarP(&notify, "notify", "n", false, "Enable email notifications")
	watch.Flags().StringVarP(&dir, "dir", "d", "./", "Directory containing Go files")
	watch.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	watch.Flags().StringVarP(&outputFile, "output", "o", "./type_check_results.json", "Output file for JSON results")
	watch.Flags().StringVarP(&configFile, "config", "c", "./config.json", "Configuration file for email notifications")

	return watch
}
