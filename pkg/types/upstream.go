package types

import (
	"context"
	"github.com/rcrowley/go-metrics"
	"gitlab.alipay-inc.com/afe/mosn/pkg/api/v2"

	"net"
)

//   Below is the basic relation between clusterManager, cluster, hostSet, and hosts:
//
//           1              * | 1                1 | 1                *| 1          *
//   clusterManager --------- cluster  --------- prioritySet --------- hostSet------hosts


// Manage connection pools and load balancing for upstream clusters.
type ClusterManager interface {
	// Add or update a cluster via API.
	AddOrUpdatePrimaryCluster(cluster v2.Cluster) bool
	
	SetInitializedCb(cb func())

	Clusters() map[string]Cluster

	Get(cluster string, context context.Context) ClusterSnapshot

	// temp interface todo: remove it
	UpdateClusterHosts(cluster string, priority uint32, hosts []v2.Host) error

	HttpConnPoolForCluster(cluster string, protocol Protocol, context context.Context) ConnectionPool

	TcpConnForCluster(cluster string, context context.Context) CreateConnectionData

	SofaRpcConnPoolForCluster(cluster string, context context.Context) ConnectionPool

	RemovePrimaryCluster(cluster string) bool

	Shutdown() error

	SourceAddress() net.Addr

	VersionInfo() string

	LocalClusterName() string

	ClusterExist(clusterName string) bool
	
	RemoveClusterHosts(clusterName string, host Host) error
}

// thread-safe cluster snapshot
type ClusterSnapshot interface {
	PrioritySet() PrioritySet

	ClusterInfo() ClusterInfo

	LoadBalancer() LoadBalancer
}

// An upstream cluster (group of hosts).
type Cluster interface {
	Initialize(cb func())

	Info() ClusterInfo

	InitializePhase() InitializePhase
	
	PrioritySet() PrioritySet

	// set the cluster's health checker
	SetHealthChecker(hc HealthChecker)
	
	// return the cluster's health checker
	HealthChecker() HealthChecker

	OutlierDetector() Detector
}

type InitializePhase string

const (
	Primary   InitializePhase = "Primary"
	Secondary InitializePhase = "Secondary"
)

type MemberUpdateCallback func(priority uint32, hostsAdded []Host, hostsRemoved []Host)

// PrioritySet is a hostSet grouped by priority for a given cluster, for ease of load balancing.
type PrioritySet interface {
	
	// Get the hostSet for this priority level, creating it if not exist.
	GetOrCreateHostSet(priority uint32) HostSet

	AddMemberUpdateCb(cb MemberUpdateCallback)

	HostSetsByPriority() []HostSet
}

// HostSet is as set of hosts that contains all of the endpoints for a given
// LocalityLbEndpoints priority level.
type HostSet interface {
	
	// all hosts that make up the set at the current time.
	Hosts() []Host

	HealthyHosts() []Host

	HostsPerLocality() [][]Host

	HealthHostsPerLocality() [][]Host

	UpdateHosts(hosts []Host, healthyHost []Host, hostsPerLocality [][]Host,
		healthyHostPerLocality [][]Host, hostsAdded []Host, hostsRemoved []Host)

	Priority() uint32
}

type HealthFlag int

const (
	// The host is currently failing active health checks.
	FAILED_ACTIVE_HC     HealthFlag = 0x1
	// The host is currently considered an outlier and has been ejected.
	FAILED_OUTLIER_CHECK HealthFlag = 0x02
)

// An upstream host
type Host interface {
	HostInfo

	// Create a connection for this host.
	CreateConnection(context context.Context) CreateConnectionData

	Counters() HostStats

	Gauges() HostStats

	ClearHealthFlag(flag HealthFlag)

	ContainHealthFlag(flag HealthFlag) bool

	SetHealthFlag(flag HealthFlag)

	Health() bool

	SetHealthChecker(healthCheck HealthCheckHostMonitor)

	SetOutlierDetector(outlierDetector DetectorHostMonitor)

	Weight() uint32

	SetWeight(weight uint32)

	Used() bool

	SetUsed(used bool)
}

type HostInfo interface {
	Hostname() string

	Canary() bool

	Metadata() v2.Metadata

	ClusterInfo() ClusterInfo

	OutlierDetector() DetectorHostMonitor

	HealthChecker() HealthCheckHostMonitor

	Address() net.Addr

	AddressString() string

	HostStats() HostStats

	// TODO: add deploy locality
}

type HostStats struct {
	Namespace                                      string
	UpstreamConnectionTotal                        metrics.Counter
	UpstreamConnectionClose                        metrics.Counter
	UpstreamConnectionActive                       metrics.Counter
	UpstreamConnectionTotalHttp1                   metrics.Counter
	UpstreamConnectionTotalHttp2                   metrics.Counter
	UpstreamConnectionTotalSofaRpc                 metrics.Counter
	UpstreamConnectionConFail                      metrics.Counter
	UpstreamConnectionLocalClose                   metrics.Counter
	UpstreamConnectionRemoteClose                  metrics.Counter
	UpstreamConnectionLocalCloseWithActiveRequest  metrics.Counter
	UpstreamConnectionRemoteCloseWithActiveRequest metrics.Counter
	UpstreamConnectionCloseNotify                  metrics.Counter
	UpstreamRequestTotal                           metrics.Counter
	UpstreamRequestActive                          metrics.Counter
	UpstreamRequestLocalReset                      metrics.Counter
	UpstreamRequestRemoteReset                     metrics.Counter
	UpstreamRequestTimeout                         metrics.Counter
	UpstreamRequestFailureEject                    metrics.Counter
	UpstreamRequestPendingOverflow                 metrics.Counter
}

type ClusterInfo interface {
	Name() string

	LbType() LoadBalancerType

	AddedViaApi() bool

	SourceAddress() net.Addr

	ConnectTimeout() int

	ConnBufferLimitBytes() uint32

	Features() int

	Metadata() v2.Metadata

	DiscoverType() string

	MaintenanceMode() bool

	MaxRequestsPerConn() uint64

	Stats() ClusterStats

	ResourceManager() ResourceManager
	
	// protocol used for health checking for this cluster
	HealthCheckProtocol() string
}

type ResourceManager interface {
	ConnectionResource() Resource

	PendingRequests() Resource

	Requests() Resource
}

type Resource interface {
	CanCreate() bool
	Increase()
	Decrease()
	Max() uint64
}

type ClusterStats struct {
	Namespace                                      string
	UpstreamConnectionTotal                        metrics.Counter
	UpstreamConnectionClose                        metrics.Counter
	UpstreamConnectionActive                       metrics.Counter
	UpstreamConnectionTotalHttp1                   metrics.Counter
	UpstreamConnectionTotalHttp2                   metrics.Counter
	UpstreamConnectionTotalSofaRpc                 metrics.Counter
	UpstreamConnectionConFail                      metrics.Counter
	UpstreamConnectionRetry                        metrics.Counter
	UpstreamConnectionLocalClose                   metrics.Counter
	UpstreamConnectionRemoteClose                  metrics.Counter
	UpstreamConnectionLocalCloseWithActiveRequest  metrics.Counter
	UpstreamConnectionRemoteCloseWithActiveRequest metrics.Counter
	UpstreamConnectionCloseNotify                  metrics.Counter
	UpstreamBytesRead                              metrics.Counter
	UpstreamBytesReadCurrent                       metrics.Gauge
	UpstreamBytesWrite                             metrics.Counter
	UpstreamBytesWriteCurrent                      metrics.Gauge
	UpstreamRequestTotal                           metrics.Counter
	UpstreamRequestActive                          metrics.Counter
	UpstreamRequestLocalReset                      metrics.Counter
	UpstreamRequestRemoteReset                     metrics.Counter
	UpstreamRequestRetry                           metrics.Counter
	UpstreamRequestTimeout                         metrics.Counter
	UpstreamRequestFailureEject                    metrics.Counter
	UpstreamRequestPendingOverflow                 metrics.Counter
}

type CreateConnectionData struct {
	Connection ClientConnection
	HostInfo   HostInfo
}

// a simple in mem cluster
type SimpleCluster interface {
	UpdateHosts(newHosts []Host)
}

type LoadBalancerType string

const (
	RoundRobin LoadBalancerType = "RoundRobin"
	Random     LoadBalancerType = "Random"
)

type LoadBalancer interface {
	ChooseHost(context context.Context) Host
}

type ClusterConfigFactoryCb interface {
	UpdateClusterConfig(configs []v2.Cluster) error
}

type ClusterHostFactoryCb interface {
	UpdateClusterHost(cluster string, priority uint32, hosts []v2.Host) error
}

type ClusterManagerFilter interface {
	OnCreated(cccb ClusterConfigFactoryCb, chcb ClusterHostFactoryCb)
}

type RegisterUpstreamUpdateMethodCb interface {
	TriggerClusterUpdate(clusterName string, hosts []v2.Host)
	GetClusterNameByServiceName(serviceName string) string
}
