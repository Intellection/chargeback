package kubernetes

import (
	"github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

type CostService struct {
}

func NewCostService() *CostService {
	return &CostService{}
}

func (cs *CostService) process(nodes []v1.Node, pods []v1.Pod, influxdbClient client.Client) {
	log.Info("processing costs...")

	log.Infof("There are %d nodes in the cluster", len(nodes))
	log.Infof("There are %d pods in the cluster", len(pods))

	var nodeInfoList []NodeInfo

	for _, node := range nodes {

		nodeInfo := NodeInfo{
			name: node.ObjectMeta.Name,
		}

		log.Infof("Node: %+v", nodeInfo)

		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	log.Infof("InfluxdbClient: %+v", influxdbClient)

}
