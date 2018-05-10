package kubernetes

import (
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

// CostService processes kubernetes node and pod info and calculates the costs
// of each pod in the cluster
type CostService struct {
	nodeInfoList   []*nodeInfo
	podInfoList    []*podInfo
	influxdbClient client.Client
}

// NewCostService creates a new CostService
func NewCostService(influxdbClient client.Client) *CostService {
	return &CostService{
		influxdbClient: influxdbClient,
	}
}

func (cs *CostService) processRawData(nodes []v1.Node, pods []v1.Pod) {
	log.Info("processing raw data...")

	for _, node := range nodes {
		allocatableMemory := node.Status.Allocatable["memory"]
		allocatableCPU := node.Status.Allocatable["cpu"]
		capacityMemory := node.Status.Capacity["memory"]
		capacityCPU := node.Status.Capacity["cpu"]

		// mocked cost for now
		cost, _ := decimal.NewFromString("100")

		nodeInfo := &nodeInfo{
			name:              node.ObjectMeta.Name,
			externalID:        node.Spec.ExternalID,
			cloudProvider:     strings.Split(node.Spec.ProviderID, ":")[0],
			allocatableMemory: allocatableMemory.Value(),
			allocatableCPU:    allocatableCPU.MilliValue(),
			capacityMemory:    capacityMemory.Value(),
			capacityCPU:       capacityCPU.MilliValue(),
			cost:              cost,
		}

		cs.nodeInfoList = append(cs.nodeInfoList, nodeInfo)
	}

	for _, pod := range pods {

		var podCPURequest int64
		var podMemoryRequest int64

		// add up the resources of each container
		for _, container := range pod.Spec.Containers {
			containerCPURequest := container.Resources.Requests["cpu"]
			containerMemoryRequest := container.Resources.Requests["memory"]

			podCPURequest += containerCPURequest.MilliValue()
			podMemoryRequest += containerMemoryRequest.Value()
		}

		podInfo := &podInfo{
			name:          pod.ObjectMeta.Name,
			namespace:     pod.ObjectMeta.Namespace,
			nodeName:      pod.Spec.NodeName,
			labels:        pod.ObjectMeta.Labels,
			cpuRequest:    podCPURequest,
			memoryRequest: podMemoryRequest,
		}

		cs.podInfoList = append(cs.podInfoList, podInfo)
	}

	nodeRequestCPUList := make(map[string]int64)
	nodeRequestMemoryList := make(map[string]int64)

	for _, podInfo := range cs.podInfoList {
		nodeRequestCPUList[podInfo.nodeName] += podInfo.cpuRequest
		nodeRequestMemoryList[podInfo.nodeName] += podInfo.memoryRequest
	}

	for _, nodeInfo := range cs.nodeInfoList {
		nodeInfo.totalRequestCPU = nodeRequestCPUList[nodeInfo.name]
		nodeInfo.totalRequestMemory = nodeRequestMemoryList[nodeInfo.name]
	}
}

func (cs *CostService) calculatePodCosts() {
	for _, podInfo := range cs.podInfoList {
		nodeInfo := cs.getPodNodeInfo(podInfo.nodeName)
		cost, _ := calculatePodCost(podInfo, nodeInfo)
		podInfo.cost = cost
	}
}

func (cs *CostService) storePodCosts() {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "chargeback",
		Precision: "s",
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, podInfo := range cs.podInfoList {
		tags := make(map[string]string)
		for key, value := range podInfo.labels {
			tags[key] = value
		}

		tags["pod_name"] = podInfo.name
		tags["node_name"] = podInfo.nodeName
		tags["namespace"] = podInfo.namespace

		podCost, _ := podInfo.cost.Float64()

		fields := map[string]interface{}{
			"monthly_cost": podCost,
		}

		pt, err := client.NewPoint("cost", tags, fields, time.Now())
		if err != nil {
			log.Fatal(err)
		}

		bp.AddPoint(pt)
	}

	// Write the batch
	if err := cs.influxdbClient.Write(bp); err != nil {
		log.Fatal(err)
	}
}

func (cs *CostService) getPodNodeInfo(nodeName string) *nodeInfo {
	for _, nodeInfo := range cs.nodeInfoList {
		if nodeName == nodeInfo.name {
			return nodeInfo
		}
	}

	return nil
}

// calculatePodCost returns a dollar value cost per month which is a fraction of the node's cost
func calculatePodCost(pod *podInfo, node *nodeInfo) (decimal.Decimal, error) {
	nodeMemoryCost := node.cost.Mul(decimal.NewFromFloat(0.5))
	nodeCPUCost := node.cost.Mul(decimal.NewFromFloat(0.5))

	podCPUUtilization := float64(pod.cpuRequest) / float64(node.allocatableCPU)
	podMemoryUtilization := float64(pod.memoryRequest) / float64(node.allocatableMemory)

	podCPUUtilizationCost := nodeCPUCost.Mul(decimal.NewFromFloat(podCPUUtilization))
	podMemoryUtilizationCost := nodeMemoryCost.Mul(decimal.NewFromFloat(podMemoryUtilization))

	podCPUUtilizationFactor := float64(pod.cpuRequest) / float64(node.totalRequestCPU)
	podMemoryUtilizationFactor := float64(pod.memoryRequest) / float64(node.totalRequestMemory)

	podCPUUnderUtilization := 1 - podCPUUtilization
	podMemoryUnderUtilization := 1 - podMemoryUtilization

	podCPUUnderUtilizationCost := podCPUUtilizationFactor * podCPUUnderUtilization
	podMemoryUnderUtilizationCost := podMemoryUtilizationFactor * podMemoryUnderUtilization

	podCPUCost := podCPUUtilizationCost.Add(decimal.NewFromFloat(podCPUUnderUtilizationCost))
	podMemoryCost := podMemoryUtilizationCost.Add(decimal.NewFromFloat(podMemoryUnderUtilizationCost))

	podCost := podCPUCost.Add(podMemoryCost)

	return podCost, nil
}
