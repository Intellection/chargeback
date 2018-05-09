package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Intellection/chargeback/agent"
	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the agent",
	Long:  `Start the agent which will collect cost information about a cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		interval, _ := cmd.Flags().GetDuration("interval")
		agent, err := agent.NewAgentFromMode(mode, interval)
		if err != nil {
			log.Fatal(err)
		}

		agent.Run()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.PersistentFlags().String("mode", "", "the mode the agent should run as (aws, gce or kubernetes)")
	agentCmd.PersistentFlags().Duration("interval", 5*time.Second, "the interval at which to process cost info")

	agentCmd.MarkFlagRequired("mode")
}
