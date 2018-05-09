package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/Intellection/chargeback/agent"
	"github.com/influxdata/influxdb/client/v2"
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
		influxHost, _ := cmd.Flags().GetString("influx-host")
		influxPort, _ := cmd.Flags().GetInt("influx-port")
		influxUsername, _ := cmd.Flags().GetString("influx-username")
		influxPassword, _ := cmd.Flags().GetString("influx-password")

		influxdbClient, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     fmt.Sprintf("%s:%d", influxHost, influxPort),
			Username: influxUsername,
			Password: influxPassword,
		})
		if err != nil {
			log.Fatal(err)
		}

		agent, err := agent.NewAgentFromMode(mode, influxdbClient, interval)
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
	agentCmd.PersistentFlags().String("influx-host", "http://localhost", "the influxdb host name")
	agentCmd.PersistentFlags().Int("influx-port", 8086, "the influxdb port")
	agentCmd.PersistentFlags().String("influx-username", "", "the influxdb username")
	agentCmd.PersistentFlags().String("influx-password", "", "the influxdb password")

	agentCmd.MarkFlagRequired("mode")
}
