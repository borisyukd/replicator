package structs

import (
	"time"

	nomad "github.com/hashicorp/nomad/api"
)

// NomadClient exposes all API methods needed to interact with the Nomad API,
// evaluate cluster capacity and allocations and make scaling decisions.
type NomadClient interface {
	// ClusterAllocationCapacity determines the total cluster capacity and current
	// number of worker nodes.
	ClusterAllocationCapacity(*ClusterCapacity) error

	// ClusterAssignedAllocation determines the consumed capacity across the
	// cluster and tracks the resource consumption of each worker node.
	ClusterAssignedAllocation(*ClusterCapacity) error

	// DrainNode places a worker node in drain mode to stop future allocations and
	// migrate existing allocations to other worker nodes.
	DrainNode(string) error

	// EvaluateClusterCapacity determines if a cluster scaling action is required.
	EvaluateClusterCapacity(*ClusterCapacity, *Config) (bool, error)

	// EvaluateJobScaling compares the consumed resource percentages of a Job group
	// against its scaling policy to determine whether a scaling event is required.
	EvaluateJobScaling([]*JobScalingPolicy)

	// GetAllocationStats discovers the resources consumed by a particular Nomad
	// allocation.
	GetAllocationStats(*nomad.Allocation, *GroupScalingPolicy)

	// GetJobAllocations identifies all allocations for an active job.
	GetJobAllocations([]*nomad.AllocationListStub, *GroupScalingPolicy)

	// IsJobRunning checks to see whether the specified jobID has any currently
	// task groups on the cluster.
	IsJobRunning(string) bool

	// JobScale takes a scaling policy and then attempts to scale the desired job
	// to the appropriate level whilst ensuring the event will not excede any job
	// thresholds set.
	JobScale(*JobScalingPolicy)

	// LeaderCheck determines if the node running replicator is the gossip pool
	// leader.
	LeaderCheck() bool

	// LeaseAllocatedNode determines the worker node consuming the least amount of
	// the cluster's mosted-utilized resource.
	LeastAllocatedNode(*ClusterCapacity) (string, string)

	// MostUtilizedResource calculates which resource is most-utilized across the
	// cluster. The worst-case allocation resource is prioritized when making
	// scaling decisions.
	MostUtilizedResource(*ClusterCapacity)

	// TaskAllocationTotals calculates the allocations required by each running
	// job and what amount of resources required if we increased the count of
	// each job by one. This allows the cluster to proactively ensure it has
	// sufficient capacity for scaling events and deal with potential node failures.
	TaskAllocationTotals(*ClusterCapacity) error

	// VerifyNodeHealth evaluates whether a specified worker node is a healthy
	// member of the Nomad cluster.
	VerifyNodeHealth(string) bool
}

// ScalingState is the central object for managing and storing all cluster
// scaling state information.
// TODO (e.westfall): Information in this struct should be persisted, see
// GH-83 for more details.
type ScalingState struct {
	// FailsafeMode tracks whether the daemon has exceeded the fault threshold
	// while attempting to perform scaling operations. When operating in failsafe
	// mode, the daemon will decline to take scaling actions of any type.
	// TODO (e.westfall): Implement failover mode functionality.
	FailsafeMode bool

	// LastScalingEvent represents the last time the daemon successfully
	// completed a cluster scaling action.
	LastScalingEvent time.Time

	// NodeFailureCount tracks the number of worker nodes that have failed to
	// successfully join the worker pool after a scale-out operation.
	NodeFailureCount int
}

// ClusterCapacity is the central object used to track and evaluate cluster
// capacity, utilization and stores the data required to make scaling
// decisions. All data stored in this object is disposable and is generated
// during each evaluation.
type ClusterCapacity struct {
	// NodeCount is the number of worker nodes in a ready and non-draining state
	// across the cluster.
	NodeCount int

	// ScalingMetric indicates the most-utilized allocation resource across the
	// cluster. The most-utilized resource is prioritized when making scaling
	// decisions like identifying the least-allocated worker node.
	ScalingMetric string

	// MaxAllowedUtilization represents the max allowed cluster utilization after
	// considering node fault-tolerance and task group scaling overhead.
	MaxAllowedUtilization int

	// ClusterTotalAllocationCapacity is the total allocation capacity across
	// the cluster.
	TotalCapacity AllocationResources

	// ClusterUsedAllocationCapacity is the consumed allocation capacity across
	// the cluster.
	UsedCapacity AllocationResources

	// TaskAllocation represents the total allocation requirements of a single
	// instance (count 1) of all running jobs across the cluster. This is used to
	// practively ensure the cluster has sufficient available capacity to scale
	// each task by +1 if an increase in capacity is required.
	TaskAllocation AllocationResources

	// NodeList is a list of all worker nodes in a known good state.
	NodeList []string

	// NodeAllocations is a slice of node allocations.
	NodeAllocations []*NodeAllocation

	// ScalingDirection is the direction in/out of cluster scaling we require
	// after performning the proper evalutation.
	ScalingDirection string
}

// NodeAllocation describes the resource consumption of a specific worker node.
type NodeAllocation struct {
	// NodeID is the unique ID of the worker node.
	NodeID string

	// NodeIP is the private IP of the worker node.
	NodeIP string

	// UsedCapacity represents the percentage of total cluster resources consumed
	// by the worker node.
	UsedCapacity AllocationResources
}

// TaskAllocation describes the resource requirements defined in the job
// specification.
type TaskAllocation struct {
	// TaskName is the name given to the task within the job specficiation.
	TaskName string

	// Resources tracks the resource requirements defined in the job spec and the
	// real-time utilization of those resources.
	Resources AllocationResources
}

// AllocationResources represents the allocation resource utilization.
type AllocationResources struct {
	MemoryMB      int
	CPUMHz        int
	DiskMB        int
	MemoryPercent float64
	CPUPercent    float64
	DiskPercent   float64
}
