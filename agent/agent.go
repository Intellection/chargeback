package agent

import (
	"fmt"
	"time"

	"github.com/Intellection/chargeback/agent/kubernetes"
	"github.com/influxdata/influxdb/client/v2"
)

type Agent interface {
	Run()
}

func NewAgentFromMode(mode string, influxdbClient client.Client, interval time.Duration) (Agent, error) {
	switch mode {
	case "kubernetes":
		return kubernetes.NewKubernetesAgent(influxdbClient, interval), nil
	default:
		return nil, fmt.Errorf("no matching mode could be found")
	}
}
