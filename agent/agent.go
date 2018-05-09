package agent

import (
	"fmt"
	"time"

	"github.com/Intellection/chargeback/agent/kubernetes"
)

type Agent interface {
	Run()
}

func NewAgentFromMode(mode string, interval time.Duration) (Agent, error) {
	switch mode {
	case "kubernetes":
		return kubernetes.NewKubernetesAgent(interval), nil
	default:
		return nil, fmt.Errorf("no matching mode could be found")
	}
}
