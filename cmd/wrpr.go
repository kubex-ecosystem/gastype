package main

import (
	c "github.com/faelmori/gastype/cmd/cli"
	l "github.com/faelmori/gastype/log"
	s "github.com/faelmori/gkbxsrv/services"
	"github.com/spf13/cobra"

	"fmt"
	"os"
	"strings"
)

type GasType struct {
	parentCmdName string
	printBanner   bool
	cfg           s.ConfigService
	certPath      string
	keyPath       string
	configPath    string
}

func (m *GasType) Alias() string {
	return ""
}
func (m *GasType) ShortDescription() string {
	return "GasType: Made to analyse your code for you."
}
func (m *GasType) LongDescription() string {
	return `GasType is a tool that helps you to analyse your code for type errors and other issues.`
}
func (m *GasType) Usage() string {
	return "gastype [command] [args]"
}
func (m *GasType) Examples() []string {
	return []string{"gastype [command] [args]", "gastype check -p ./example"}
}
func (m *GasType) Active() bool {
	return true
}
func (m *GasType) Module() string { return "gastype" }
func (m *GasType) Execute() error {
	dbChanData := make(chan interface{})
	defer close(dbChanData)

	if spyderErr := m.Command().Execute(); spyderErr != nil {
		l.Error(spyderErr.Error(), nil)
		return spyderErr
	} else {
		return nil
	}
}
func (m *GasType) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     m.Module(),
		Aliases: []string{m.Alias()},
		Example: m.concatenateExamples(),
		Annotations: c.GetDescriptions(
			[]string{m.ShortDescription(), m.LongDescription()}, false,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			l.Error("No command specified.", nil)
			return nil
		},
	}

	cmd.AddCommand(c.TypeCheckCmds()...)

	setUsageDefinition(cmd)

	for _, subCmd := range cmd.Commands() {
		setUsageDefinition(subCmd)
		subCmd.PreRunE = m.preRunEMethod
	}

	cmd.PreRunE = m.preRunEMethod

	return cmd
}
func (m *GasType) preRunEMethod(cmd *cobra.Command, args []string) error {
	l.Debug(fmt.Sprintf("PreRunE: %s", cmd.Name()), nil)
	if m.cfg == nil {
		m.cfg = s.NewConfigService(m.configPath, m.keyPath, m.certPath)
	}
	if setupConfigErr := m.cfg.SetupConfig(); setupConfigErr != nil {
		l.Error(setupConfigErr.Error(), nil)
		return setupConfigErr
	}
	return nil
}
func (m *GasType) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}
func (m *GasType) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}

func RegX() *GasType {
	var configPath = os.Getenv("GasType_CONFIGFILE")
	var keyPath = os.Getenv("GasType_KEYFILE")
	var certPath = os.Getenv("GasType_CERTFILE")
	var printBannerV = os.Getenv("GasType_PRINTBANNER")
	if printBannerV == "" {
		printBannerV = "true"
	}

	return &GasType{
		configPath:  configPath,
		keyPath:     keyPath,
		certPath:    certPath,
		printBanner: strings.ToLower(printBannerV) == "true",
	}
}
