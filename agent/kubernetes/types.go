package kubernetes

import "github.com/shopspring/decimal"

// TODO: add storage
type NodeInfo struct {
	name              string
	cloudProvider     string
	cost              decimal.Decimal
	externalID        string
	allocatableMemory int64
	allocatableCPU    int64
	capacityMemory    int64
	capacityCPU       int64
	utilizedCPU       int
	utilizedMemory    int
}

type PodInfo struct {
	name          string
	labels        map[string]string
	cpuRequest    int64
	memoryRequest int64
	nodeName      string
}
