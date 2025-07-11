package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "envguard",
		Short: "EnvGuard checks your environment for misconfigurations and drift",
		Long:  `A CLI tool to detect issues in Terraform, Kubernetes, AWS, Docker, and Git repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("EnvGuard: No subcommand specified. Try 'envguard scan'")
		},
	}
)

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .envguard.yaml)")

	// Optional global flags here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Bind to Viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".envguard")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	viper.AutomaticEnv() // support ENV vars

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
