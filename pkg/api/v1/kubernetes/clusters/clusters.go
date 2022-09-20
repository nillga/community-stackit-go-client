// package clusters is used to create and manage STACKIT Kubernetes Enging (SKE) clusters

package clusters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SchwarzIT/community-stackit-go-client/internal/common"
	"github.com/SchwarzIT/community-stackit-go-client/pkg/consts"
	"github.com/SchwarzIT/community-stackit-go-client/pkg/validate"
)

// constants
const (
	apiPath        = consts.API_PATH_SKE_CLUSTERS
	apiPathCluster = consts.API_PATH_SKE_WITH_CLUSTER_ID
)

// New returns a new handler for the service
func New(c common.Client) *KubernetesClusterService {
	return &KubernetesClusterService{
		Client: c,
	}
}

// KubernetesClusterService is the service that handles
// CRUD functionality for SKE clusters
type KubernetesClusterService common.Service

// Cluster is a struct representation of a cluster in STACKIT api
type Cluster struct {
	Name        string       `json:"name"` // 11 lowercase letters, numbers, or hyphens
	Kubernetes  Kubernetes   `json:"kubernetes"`
	Nodepools   []NodePool   `json:"nodepools"`
	Maintenance *Maintenance `json:"maintenance,omitempty"`
	Hibernation *Hibernation `json:"hibernation,omitempty"`
	Extensions  *Extensions  `json:"extensions,omitempty"`
	Status      *Status      `json:"status,omitempty"`
}

// Kubernetes contains the cluster's kubernetes config
type Kubernetes struct {
	Version                   string `json:"version"`
	AllowPrivilegedContainers bool   `json:"allowPrivilegedContainers"`
}

// NodePool is a struct representing a node pool in the cluster
type NodePool struct {
	Name              string            `json:"name,omitempty"`
	Machine           Machine           `json:"machine"`
	Minimum           int               `json:"minimum"`
	Maximum           int               `json:"maximum"`
	MaxSurge          int               `json:"maxSurge"`
	MaxUnavailable    int               `json:"maxUnavailable"`
	Volume            Volume            `json:"volume"`
	Labels            map[string]string `json:"labels"`
	Taints            []Taint           `json:"taints"`
	CRI               CRI               `json:"cri"`
	AvailabilityZones []string          `json:"availabilityZones"`
}

// Machine contains information of the machine in the node pool
type Machine struct {
	Type  string       `json:"type"`
	Image MachineImage `json:"image"`
}

// MachineImage contains information of the machine's image
type MachineImage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Volume is the node pool volume information
type Volume struct {
	Type string `json:"type"`
	Size int    `json:"size"`
}

// Taint is a taint of the node pool
type Taint struct {
	Effect string `json:"effect"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// CRI is the container runtime interface of the node pool
type CRI struct {
	Name string `json:"name"`
}

// Maintenance is the node pool's maintenance window
type Maintenance struct {
	AutoUpdate MaintenanceAutoUpdate `json:"autoUpdate"`
	TimeWindow MaintenanceTimeWindow `json:"timeWindow"`
}

// MaintenanceAutoUpdate is the auto update confguration
type MaintenanceAutoUpdate struct {
	KubernetesVersion   bool `json:"kubernetesVersion"`
	MachineImageVersion bool `json:"machineImageVersion"`
}

// MaintenanceTimeWindow is when the maintenance window should happen
type MaintenanceTimeWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Hibernation schedule
type Hibernation struct {
	Schedules []HibernationScedule `json:"schedules"`
}

// HibernationScedule is the schedule for hibernation
type HibernationScedule struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Timezone string `json:"timezone"`
}

// Extensions represent SKE extensions
type Extensions struct {
	Argus *ArgusExtension `json:"argus,omitempty"`
}

// ArgusExtension is Argus extension
type ArgusExtension struct {
	Enabled         bool   `json:"enabled"`
	ArgusInstanceID string `json:"argusInstanceId"`
}

// Status is the cluster status
type Status struct {
	Hibernated bool   `json:"hibernated"`
	Aggregated string `json:"aggregated"`
	Error      struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
	} `json:"error,omitempty"`
}

// List returns the clusters in the project
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_ListClusters
func (svc *KubernetesClusterService) List(ctx context.Context, projectID string) (res []Cluster, err error) {
	req, err := svc.Client.Request(ctx, http.MethodGet, fmt.Sprintf(apiPath, projectID), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// Get returns the a cluster by project ID and cluster name
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_ListClusters
func (svc *KubernetesClusterService) Get(ctx context.Context, projectID, clusterName string) (res Cluster, err error) {
	req, err := svc.Client.Request(ctx, http.MethodGet, fmt.Sprintf(apiPathCluster, projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// Create creates a new SKE cluster
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_CreateOrUpdateCluster
func (svc *KubernetesClusterService) Create(
	ctx context.Context,
	projectID string,
	clusterName string,
	clusterConfig Kubernetes,
	nodePools []NodePool,
	maintenance *Maintenance,
	hibernation *Hibernation,
	extensions *Extensions,
) (res Cluster, err error) {
	if err = ValidateCluster(
		clusterName,
		clusterConfig,
		nodePools,
		maintenance,
		hibernation,
		extensions,
	); err != nil {
		err = validate.WrapError(err)
		return
	}

	body, _ := svc.buildCreateRequest(
		projectID,
		clusterName,
		clusterConfig,
		nodePools,
		maintenance,
		hibernation,
		extensions,
	)
	req, err := svc.Client.Request(ctx, http.MethodPut, fmt.Sprintf(apiPathCluster, projectID, clusterName), body)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

func (svc *KubernetesClusterService) buildCreateRequest(
	projectID string,
	clusterName string,
	clusterConfig Kubernetes,
	nodePools []NodePool,
	maintenance *Maintenance,
	hibernation *Hibernation,
	extensions *Extensions,
) ([]byte, error) {
	return json.Marshal(Cluster{
		Name:        clusterName,
		Kubernetes:  clusterConfig,
		Nodepools:   nodePools,
		Maintenance: maintenance,
		Hibernation: hibernation,
		Extensions:  extensions,
	})
}

// Update updates an SKE cluster or creates a new one if it doesn't exist
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_CreateOrUpdateCluster
func (svc *KubernetesClusterService) Update(
	ctx context.Context,
	projectID string,
	clusterName string,
	clusterConfig Kubernetes,
	nodePools []NodePool,
	maintenance *Maintenance,
	hibernation *Hibernation,
	extensions *Extensions,
) (res Cluster, err error) {
	return svc.Create(
		ctx,
		projectID,
		clusterName,
		clusterConfig,
		nodePools,
		maintenance,
		hibernation,
		extensions,
	)
}

// Delete deletes an SKE cluster
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_DeleteCluster
func (svc *KubernetesClusterService) Delete(ctx context.Context, projectID, clusterName string) (err error) {
	req, err := svc.Client.Request(ctx, http.MethodDelete, fmt.Sprintf(apiPathCluster, projectID, clusterName), nil)
	if err != nil {
		return
	}
	_, err = svc.Client.Do(req, nil)
	return err
}

// Hibernate triggers cluster hibernation
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_TriggerClusterHibernation
func (svc *KubernetesClusterService) Hibernate(ctx context.Context, projectID, clusterName string) (res Cluster, err error) {
	req, err := svc.Client.Request(ctx, http.MethodPost, fmt.Sprintf(apiPathCluster+"/hibernate", projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// Maintenance triggers cluster maintenance
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_TriggerClusterMaintenance
func (svc *KubernetesClusterService) Maintenance(ctx context.Context, projectID, clusterName string) (res Cluster, err error) {
	req, err := svc.Client.Request(ctx, http.MethodPost, fmt.Sprintf(apiPathCluster+"/maintenance", projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// Reconcile triggers cluster reconciliation
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_TriggerClusterReconciliation
func (svc *KubernetesClusterService) Reconcile(ctx context.Context, projectID, clusterName string) (res Cluster, err error) {
	req, err := svc.Client.Request(ctx, http.MethodPost, fmt.Sprintf(apiPathCluster+"/reconcile", projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// Wakeup triggers cluster wakeup
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_TriggerClusterWakeup
func (svc *KubernetesClusterService) Wakeup(ctx context.Context, projectID, clusterName string) (res Cluster, err error) {
	req, err := svc.Client.Request(ctx, http.MethodPost, fmt.Sprintf(apiPathCluster+"/wakeup", projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// Credentials is the struct response for cluster credentils
type Credentials struct {
	Server                   string `json:"server"`
	Kubeconfig               string `json:"kubeconfig"`
	CertificateAuthorityData string `json:"certificateAuthorityData"`
	Token                    string `json:"token"`
}

// GetCredential returns the a credentials for the cluster
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#tag/Credentials
func (svc *KubernetesClusterService) GetCredential(ctx context.Context, projectID, clusterName string) (res Credentials, err error) {
	req, err := svc.Client.Request(ctx, http.MethodGet, fmt.Sprintf(apiPathCluster+"/credentials", projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, &res)
	return
}

// RotateCredentials triggers cluster credentials rotation
// See also https://api.stackit.schwarz/ske-service/openapi.v1.html#operation/SkeService_TriggerClusterCredentialRotation
func (svc *KubernetesClusterService) RotateCredentials(ctx context.Context, projectID, clusterName string) (err error) {
	req, err := svc.Client.Request(ctx, http.MethodPost, fmt.Sprintf(apiPathCluster+"/rotate-credentials", projectID, clusterName), nil)
	if err != nil {
		return
	}

	_, err = svc.Client.Do(req, nil)
	return
}
