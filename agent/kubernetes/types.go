package kubernetes

import "github.com/shopspring/decimal"

// TODO: add storage
type nodeInfo struct {
	name               string
	cloudProvider      string
	cost               decimal.Decimal
	externalID         string
	capacityMemory     int64 // Full memory capacity of the node in megabytes
	capacityCPU        int64 // Number of CPUs the node has access to
	allocatableMemory  int64 // Full memory capacity minus OS/docker overhead.
	allocatableCPU     int64 // Number of CPUs the Kubernetes has access to (same as capacityCPU)
	totalRequestCPU    int64 // Total utilized CPU used by scheduled pods on the node
	totalRequestMemory int64 // Total utilized memory used by scheduled pods on the node
}

type podInfo struct {
	name          string
	namespace     string
	labels        map[string]string
	cpuRequest    int64
	memoryRequest int64
	nodeName      string
	cost          decimal.Decimal
}
