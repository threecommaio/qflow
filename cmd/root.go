// Copyright © 2018 ThreeComma.io <hello@threecomma.io>

package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/threecommaio/qflow/pkg/qflow"
)

var addr string
var config string
var dataDir string
var VERSION string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "qflow",
	Long: `Replicates traffic from to various http endpoints

This tool helps replicate to multiple http endpoints backed by a
durable disk queue in the event of failures or slowdowns.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		config, err := qflow.ParseConfig(config)
		if err != nil {
			log.Fatal(err)
		}

		qflow.ListenAndServe(config, addr, dataDir)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	VERSION = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "listen addr")
	rootCmd.Flags().StringVarP(&config, "config", "c", "", "config file for the clusters in yaml format")
	rootCmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "data directory for storage")
	rootCmd.MarkFlagRequired("config")
	rootCmd.MarkFlagRequired("data-dir")
}

func initConfig() {
	debug, _ := rootCmd.Flags().GetBool("debug")
	log.SetFormatter(qflow.UTCFormatter{&log.TextFormatter{FullTimestamp: true}})

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	viper.AutomaticEnv() // read in environment variables that match
}
