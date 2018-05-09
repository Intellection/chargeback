package kubernetes

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

type CostService struct {
}

func NewCostService() *CostService {
	return &CostService{}
}

func (cs *CostService) process(nodes []v1.Node, pods []v1.Pod) {
	log.Info("processing costs...")

	log.Infof("There are %d nodes in the cluster", len(nodes))
	log.Infof("There are %d pods in the cluster", len(pods))
}
