package cmd

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/catalystsquad/app-utils-go/logging"
	"github.com/catalystsquad/prometheus-salesforce-exporter/internal/exporter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateServeArgs(); err != nil {
			return err
		}
		if err := exporter.Run(exporterConfig); err != nil {
			logging.Log.WithError(err).Fatal("Failed to run exporter")
			os.Exit(1)
		}
		return nil
	},
}

var exporterConfig = &exporter.SalesforceExporterConfig{}

func init() {
	rootCmd.AddCommand(serveCmd)
	cobra.OnInitialize(initServeConfig)

	serveCmd.Flags().IntVarP(&exporterConfig.HTTPPort, "http-port", "p", 8080, "HTTP Port")
	serveCmd.Flags().DurationVar(&exporterConfig.PollInterval, "poll-interval", 1*time.Minute, "Poll Interval")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforceBaseUrl, "salesforce-base-url", "", "Salesforce Base URL")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforceClientID, "salesforce-client-id", "", "Salesforce Client ID")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforceClientSecret, "salesforce-client-secret", "", "Salesforce Client Secret")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforceUsername, "salesforce-username", "", "Salesforce Username")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforcePassword, "salesforce-password", "", "Salesforce Password")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforceApiVersion, "salesforce-api-version", "58.0", "Salesforce API Version")
	serveCmd.Flags().StringVar(&exporterConfig.SalesforceGrantType, "salesforce-grant-type", "password", "Salesforce Grant Type")
}

func initServeConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	bindFlags(serveCmd)
}

func validateServeArgs() error {
	missingArgs := []string{}
	if exporterConfig.SalesforceBaseUrl == "" {
		missingArgs = append(missingArgs, "salesforce-base-url")
	}
	if exporterConfig.SalesforceClientID == "" {
		missingArgs = append(missingArgs, "salesforce-client-id")
	}
	if exporterConfig.SalesforceClientSecret == "" {
		missingArgs = append(missingArgs, "salesforce-client-secret")
	}
	if exporterConfig.SalesforceUsername == "" {
		missingArgs = append(missingArgs, "salesforce-username")
	}
	if exporterConfig.SalesforcePassword == "" {
		missingArgs = append(missingArgs, "salesforce-password")
	}
	if len(missingArgs) > 0 {
		return errors.New("Missing required flags: " + strings.Join(missingArgs, ", "))
	}
	return nil
}
