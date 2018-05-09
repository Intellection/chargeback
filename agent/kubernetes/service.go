package kubernetes

import (
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

type CostService struct {
	nodeInfoList   []*NodeInfo
	podInfoList    []*PodInfo
	influxdbClient client.Client
}

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

		nodeInfo := &NodeInfo{
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

		podInfo := &PodInfo{
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
		nodeInfo := cs.getPodNodeInfo(podInfo)
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

func (cs *CostService) getPodNodeInfo(pod *PodInfo) *NodeInfo {
	for _, nodeInfo := range cs.nodeInfoList {
		if pod.nodeName == nodeInfo.name {
			return nodeInfo
		}
	}

	return nil
}

func calculateNodeCost(instanceTypeCost decimal.Decimal, diskCost decimal.Decimal) (decimal.Decimal, error) {
	// Given a node's base instance type cost, together with any additional costs such as the SSD
	// calculateNodeCost() will return the total cost per month of the node as:
	// instance_type cost + any SSD mounted on the instance.
	// return (instanceTypeCost + diskCost), nil
	return decimal.NewFromString("100.00") // Mocking out $100 for now, per month
}

func calculatePodCost(pod *PodInfo, node *NodeInfo) (decimal.Decimal, error) {
	// Given a pod's CPU and Memory request, together with the pod's node and it's:
	// Node cost, allocatable CPU, allocatable Memory, Total CPU, Total Memory,
	// // calculatePodCost returns a dollar value cost per month which is a fraction of the node's cost
	// cpuReq := float64(pod.cpuRequest)
	// memReq := float64(pod.memoryRequest)
	// nodeCost := node.cost
	// allocatableCPU := float64(node.allocatableCPU)
	// allocatableMemory := float64(node.allocatableMemory)
	// capacityCPU := float64(node.capacityCPU)
	// capacityMemory := float64(node.capacityMemory)

	memoryCost := node.cost.Mul(decimal.NewFromFloat(0.5))
	cpuCost := node.cost.Mul(decimal.NewFromFloat(0.5))

	podCPUUtilization := float64(pod.cpuRequest) / float64(node.allocatableCPU)
	podMemoryUtilization := float64(pod.memoryRequest) / float64(node.allocatableMemory)

	podCPUUtilizationCost := cpuCost.Mul(decimal.NewFromFloat(podCPUUtilization))
	podMemoryUtilizationCost := memoryCost.Mul(decimal.NewFromFloat(podMemoryUtilization))

	podCPUUtilizationFactor := float64(pod.cpuRequest) / float64(node.totalRequestCPU)
	podMemoryUtilizationFactor := float64(pod.memoryRequest) / float64(node.totalRequestMemory)

	podCPUUnderUtilization := 1 - podCPUUtilization
	podMemoryUnderUtilization := 1 - podMemoryUtilization

	podCPUUnderUtilizationCost := podCPUUtilizationFactor * podCPUUnderUtilization
	podMemoryUnderUtilizationCost := podMemoryUtilizationFactor * podMemoryUnderUtilization

	// log.Infof("Pod: %s", pod.name)
	// log.Infof("CPU utilization percent: %2f%%", podCPUUtilization*100)
	// log.Infof("Memory utilization percent: %2f%%", podMemoryUtilization*100)
	//
	// log.Infof("CPU utilization cost: %s", podCPUUtilizationCost.String())
	// log.Infof("Memory utilization cost: %s", podMemoryUtilizationCost.String())
	//
	// log.Infof("CPU under-utilization cost: %2f", podCPUUnderUtilizationCost)
	// log.Infof("Memory under-utilization cost: %2f", podMemoryUnderUtilizationCost)

	podCPUCost := podCPUUtilizationCost.Add(decimal.NewFromFloat(podCPUUnderUtilizationCost))
	podMemoryCost := podMemoryUtilizationCost.Add(decimal.NewFromFloat(podMemoryUnderUtilizationCost))

	podCost := podCPUCost.Add(podMemoryCost)

	return podCost, nil
}
