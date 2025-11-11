package azuremonitormetricsreceiver

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// AzureMonitorMetricsReceiver is the receiver that gets metrics of Azure resources using Azure Monitor API.
type AzureMonitorMetricsReceiver struct {
	Targets      *Targets
	AzureClients *AzureClients

	subscriptionID string
}

// Targets contains all targets types.
type Targets struct {
	ResourceTargets []*ResourceTarget

	resourceGroupTargets []*ResourceGroupTarget
	subscriptionTargets  []*Resource
}

// ResourceTarget describes an Azure resource by resource ID.
type ResourceTarget struct {
	ResourceID   string
	Metrics      []string
	Aggregations []string
}

// ResourceGroupTarget describes an Azure resource group.
type ResourceGroupTarget struct {
	resourceGroup string
	resources     []*Resource
}

// Resource describes an Azure resource by resource type.
type Resource struct {
	resourceType string
	metrics      []string
	aggregations []string
}

// AzureClients contains all clients that communicate with Azure Monitor API.
type AzureClients struct {
	Ctx                     context.Context
	ResourcesClient         ResourcesClient
	MetricDefinitionsClient MetricDefinitionsClient
	MetricsClient           MetricsClient
}

// Metric is a metric of an Azure resource using Azure Monitor API.
type Metric struct {
	Name   string
	Fields map[string]interface{}
	Tags   map[string]string
}

type azureResourcesClient struct {
	client *armresources.Client
}

// ResourcesClient is an Azure resources client interface.
type ResourcesClient interface {
	List(context.Context, *armresources.ClientListOptions) ([]*armresources.ClientListResponse, error)
	ListByResourceGroup(context.Context, string, *armresources.ClientListByResourceGroupOptions) ([]*armresources.ClientListByResourceGroupResponse, error)
}

// MetricDefinitionsClient is an Azure metric definitions client interface.
type MetricDefinitionsClient interface {
	List(context.Context, string, *armmonitor.MetricDefinitionsClientListOptions) (armmonitor.MetricDefinitionsClientListResponse, error)
}

// MetricsClient is an Azure metrics client interface.
type MetricsClient interface {
	List(context.Context, string, *armmonitor.MetricsClientListOptions) (armmonitor.MetricsClientListResponse, error)
}

// NewAzureMonitorMetricsReceiver lets you create a new receiver.
func NewAzureMonitorMetricsReceiver(subscriptionID string, targets *Targets, azureClients *AzureClients) (*AzureMonitorMetricsReceiver, error) {
	azureMonitorMetricsReceiver := &AzureMonitorMetricsReceiver{
		Targets:        targets,
		AzureClients:   azureClients,
		subscriptionID: subscriptionID,
	}

	if err := azureMonitorMetricsReceiver.checkValidation(); err != nil {
		return nil, fmt.Errorf("got validation error: %v", err)
	}

	azureMonitorMetricsReceiver.addPrefixToResourceTargetsResourceID()
	return azureMonitorMetricsReceiver, nil
}

// NewTargets lets you create a new targets object.
func NewTargets(resourceTargets []*ResourceTarget, resourceGroupTargets []*ResourceGroupTarget, subscriptionTargets []*Resource) *Targets {
	return &Targets{
		ResourceTargets:      resourceTargets,
		resourceGroupTargets: resourceGroupTargets,
		subscriptionTargets:  subscriptionTargets,
	}
}

// NewResourceTarget lets you create a new resource target.
func NewResourceTarget(resourceID string, metrics []string, aggregations []string) *ResourceTarget {
	return &ResourceTarget{
		ResourceID:   resourceID,
		Metrics:      metrics,
		Aggregations: aggregations,
	}
}

// NewResourceGroupTarget lets you create a new resource group target.
func NewResourceGroupTarget(resourceGroup string, resources []*Resource) *ResourceGroupTarget {
	return &ResourceGroupTarget{
		resourceGroup: resourceGroup,
		resources:     resources,
	}
}

// NewResource lets you create a new resource.
func NewResource(resourceType string, metrics []string, aggregations []string) *Resource {
	return &Resource{
		resourceType: resourceType,
		metrics:      metrics,
		aggregations: aggregations,
	}
}
