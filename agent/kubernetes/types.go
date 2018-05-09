package kubernetes

import "github.com/shopspring/decimal"

// TODO: add storage
type NodeInfo struct {
	name              string
	cloudProvider     string
	cost              decimal.Decimal
	externalID        string
	capacityMemory    int64 // Full memory capacity of the node in megabytes
	capacityCPU       int64 // Number of CPUs the node has access to
	allocatableMemory int64 // Full memory capacity minus OS/docker overhead.
	allocatableCPU    int64 // Number of CPUs the Kubernetes has access to (same as capacityCPU)
	utilizedCPU       int   // Total utilized CPU used by scheduled pods on the node
	utilizedMemory    int   // Total utilized memory used by scheduled pods on the node
}

type PodInfo struct {
	name          string
	labels        map[string]string
	cpuRequest    int64
	memoryRequest int64
	nodeName      string
}
