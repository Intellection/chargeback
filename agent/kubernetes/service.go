package kubernetes

import (
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
	// var podInfoList []PodInfo

	for _, node := range nodes {
		nodeInfo := NodeInfo{
			name: node.ObjectMeta.Name,
		}

		log.Infof("Node: %+v", node)

		nodeInfoList = append(nodeInfoList, nodeInfo)
	}

	// for _, pod := range pods {
	// 	podInfo := PodInfo{}
	// }

}
