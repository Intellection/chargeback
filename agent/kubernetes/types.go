package kubernetes

import "github.com/shopspring/decimal"

// TODO: add storage
type NodeInfo struct {
	name              string
	cloudProvider     string
	cost              decimal.Decimal
	externalID        string
	allocatableMemory string
	allocatableCPU    int
	capacityMemory    string
	capacityCPU       int
	utilizedCPU       int
	utilizedMemory    int
}

type PodInfo struct {
	name          string
	labels        map[string]string
	cpuRequest    int
	memoryRequest string
	nodeName      string
}
