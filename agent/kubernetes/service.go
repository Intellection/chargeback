package kubernetes

import (
	"os"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/shopspring/decimal"
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

func nodeCost(instanceTypeCost decimal.Decimal, diskCost decimal.Decimal) (decimal.Decimal, error) {
	// Given a node's base instance type cost, together with any additional costs such as the SSD
	// nodeCost() will return the total cost per month of the node as:
	// instance_type cost + any SSD mounted on the instance.
	// return (instanceTypeCost + diskCost), nil
	return decimal.NewFromString("100.00") // Mocking out $100 for now, per month
}

func podCost(pod PodInfo, node NodeInfo) (decimal.Decimal, error) {
	// Given a pod's CPU and Memory request, together with the pod's node and it's:
	// Node cost, allocatable CPU, allocatable Memory, Total CPU, Total Memory,
	// podCost returns a dollar value cost per month which is a fraction of the node's cost
	cpuReq := float64(pod.cpuRequest)
	memReq := float64(pod.memRequest)
	nodeCost := node.cost
	allocatableCPU := float64(node.allocatableCPU)
	allocatableMemory := float64(node.allocatableMemory)
	capacityCPU := float64(node.capacityCPU)
	capacityMemory := float64(node.capacityMemory)

	return decimal.NewFromString("1.00")
}
