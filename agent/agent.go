package agent

import (
	"fmt"
	"time"

	"github.com/Intellection/chargeback/agent/kubernetes"
	"github.com/influxdata/influxdb/client/v2"
)

// Agent represents a cost information collection agent
type Agent interface {
	Run()
}

// NewAgentFromMode creates a new Agent depending on the mode being used
func NewAgentFromMode(mode string, influxdbClient client.Client, interval time.Duration) (Agent, error) {
	switch mode {
	case "kubernetes":
		return kubernetes.NewAgent(influxdbClient, interval), nil
	default:
		return nil, fmt.Errorf("no matching mode could be found")
	}
}
