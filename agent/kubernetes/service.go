package kubernetes

import (
	"os"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

type CostService struct {
	influxdbClient client.Client
}

func NewCostService(influxdbClient client.Client) *CostService {
	return &CostService{
		influxdbClient: influxdbClient,
	}
}

func (cs *CostService) processRawData(nodes []v1.Node, pods []v1.Pod) {
	log.Info("processing raw data...")

	var nodeInfoList []NodeInfo
	var podInfoList []PodInfo

	for _, node := range nodes {
		allocatableMemory := node.Status.Allocatable["memory"]
		allocatableCPU := node.Status.Allocatable["cpu"]
		capacityMemory := node.Status.Capacity["memory"]
		capacityCPU := node.Status.Capacity["cpu"]

		nodeInfo := NodeInfo{
			name:              node.ObjectMeta.Name,
			externalID:        node.Spec.ExternalID,
			cloudProvider:     strings.Split(node.Spec.ProviderID, ":")[0],
			allocatableMemory: allocatableMemory.Value(),
			allocatableCPU:    allocatableCPU.Value(),
			capacityMemory:    capacityMemory.Value(),
			capacityCPU:       capacityCPU.Value(),
		}

		nodeInfoList = append(nodeInfoList, nodeInfo)
	}

	for _, pod := range pods {

		log.Infof("Pod: %+v", pod)

		podInfo := PodInfo{}

		podInfoList = append(podInfoList, podInfo)

		os.Exit(0)
	}
}
