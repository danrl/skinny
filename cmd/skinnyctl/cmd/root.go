package cmd

import (
	"fmt"
	"os"

	"github.com/danrl/skinny/config"
	"github.com/spf13/cobra"
)

var (
	err error

	flagConfigFile string
	flagInstance   string

	cfgQuorum          *config.QuorumConfig
	cfgInstances       map[string]string
	cfgDefaultInstance string
)

func init() {
	cfgInstances = make(map[string]string)
	rootCmd.PersistentFlags().StringVar(&flagConfigFile, "config", "quorum.yml", "Skinny quorum configuration file")
}

var rootCmd = &cobra.Command{
	Use: "skinnyctl",
	//	Short: "Skinnyctl is a control tool for Skinny instances",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// load the qorum configuration from file
		cfgQuorum, err = config.NewQuorumConfig(flagConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load config: %v\n", err)
			os.Exit(1)
		}

		// create a hashmap for easier access to instance addresses by name
		// also define the default instance name (the first one in the list)
		for i, in := range cfgQuorum.Instances {
			cfgInstances[in.Name] = in.Address
			if i == 0 {
				cfgDefaultInstance = in.Name
			}
		}
	},
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
