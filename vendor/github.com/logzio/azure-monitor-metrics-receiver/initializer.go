package azuremonitormetricsreceiver

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

const (
	// MaxMetricsPerRequest is max metrics per request to Azure Monitor API.
	MaxMetricsPerRequest = 20
)

type metricDefWrapper struct {
	client *armmonitor.MetricDefinitionsClient
}

type AzureClientOptions struct {
	clientOptions *azcore.ClientOptions
}

func (w *metricDefWrapper) List(ctx context.Context, resourceID string, options *armmonitor.MetricDefinitionsClientListOptions) (armmonitor.MetricDefinitionsClientListResponse, error) {
	var response armmonitor.MetricDefinitionsClientListResponse

	pager := w.client.NewListPager(resourceID, options)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return armmonitor.MetricDefinitionsClientListResponse{}, err
		}
		response.MetricDefinitionCollection.Value = append(response.MetricDefinitionCollection.Value, page.Value...)
	}
	return response, nil
}

// CreateAzureClients creates Azure clients with service principal credentials
func CreateAzureClients(subscriptionID string, clientID string, clientSecret string, tenantID string, clientOptions ...func(*AzureClientOptions)) (*AzureClients, error) {
	var options *azidentity.ClientSecretCredentialOptions = nil
	azureClientOptions := checkOptionalClientParameters(clientOptions, &AzureClientOptions{})

	if azureClientOptions != nil {
		options = &azidentity.ClientSecretCredentialOptions{ClientOptions: *azureClientOptions.clientOptions}
	}

	credential, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, options)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure client credential: %w", err)
	}

	if azureClientOptions != nil {
		return CreateAzureClientsWithCreds(subscriptionID, credential, WithAzureClientOptions(azureClientOptions.clientOptions))
	} else {
		return CreateAzureClientsWithCreds(subscriptionID, credential)
	}
}

// CreateAzureClientsWithCreds creates Azure clients with provided TokenCredential
func CreateAzureClientsWithCreds(subscriptionID string, credential azcore.TokenCredential, clientOptions ...func(*AzureClientOptions)) (*AzureClients, error) {
	var armClientOptions *arm.ClientOptions = nil
	azureClientOptions := checkOptionalClientParameters(clientOptions, &AzureClientOptions{})

	if azureClientOptions != nil {
		armClientOptions = &arm.ClientOptions{ClientOptions: *azureClientOptions.clientOptions}
	}

	metricClient, err := armmonitor.NewMetricsClient(subscriptionID, credential, armClientOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure metric client: %w", err)
	}
	defClient, err := armmonitor.NewMetricDefinitionsClient(subscriptionID, credential, armClientOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure definitions client: %w", err)
	}

	resClient, err := armresources.NewClient(subscriptionID, credential, armClientOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure definitions client: %w", err)
	}

	return &AzureClients{
		Ctx:                     context.Background(),
		ResourcesClient:         &azureResourcesClient{client: resClient},
		MetricsClient:           metricClient,
		MetricDefinitionsClient: &metricDefWrapper{client: defClient},
	}, nil
}

type ClientOptions func(*AzureClientOptions)

func WithAzureClientOptions(options *azcore.ClientOptions) ClientOptions {
	return func(azureClientOptions *AzureClientOptions) {
		azureClientOptions.clientOptions = options
	}
}

func checkOptionalClientParameters(clientOptions []func(*AzureClientOptions), azureClientOptions *AzureClientOptions) *AzureClientOptions {
	if clientOptions != nil {
		for _, optArgs := range clientOptions {
			optArgs(azureClientOptions)
		}

		if &azureClientOptions != nil {
			return azureClientOptions
		}
	}
	return nil
}

func (ammr *AzureMonitorMetricsReceiver) checkValidation() error {
	if ammr.subscriptionID == "" {
		return fmt.Errorf("subscription ID is empty or missing")
	}

	if len(ammr.Targets.ResourceTargets) == 0 && len(ammr.Targets.resourceGroupTargets) == 0 && len(ammr.Targets.subscriptionTargets) == 0 {
		return fmt.Errorf("no target to collect metrics from")
	}

	if err := ammr.checkResourceTargetsValidation(); err != nil {
		return err
	}

	if err := ammr.checkResourceGroupTargetsValidation(); err != nil {
		return err
	}

	return ammr.checkSubscriptionTargetValidation()
}

func (ammr *AzureMonitorMetricsReceiver) checkResourceTargetsValidation() error {
	for index, target := range ammr.Targets.ResourceTargets {
		if target.ResourceID == "" {
			return fmt.Errorf(
				"resource target #%d resource ID is empty or missing", index+1)
		}

		if len(target.Aggregations) > 0 {
			if !areTargetAggregationsValid(target.Aggregations) {
				return fmt.Errorf("resource target #%d aggregations contain invalid aggregation/s. "+
					"The valid aggregations are: %s", index, strings.Join(getPossibleAggregations(), ", "))
			}
		}
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) checkResourceGroupTargetsValidation() error {
	for resourceGroupIndex, target := range ammr.Targets.resourceGroupTargets {
		if target.resourceGroup == "" {
			return fmt.Errorf(
				"resource group target #%d resource group is empty or missing",
				resourceGroupIndex+1)
		}

		if len(target.resources) == 0 {
			return fmt.Errorf("resource group target #%d has no resources", resourceGroupIndex+1)
		}

		for resourceIndex, resource := range target.resources {
			if resource.resourceType == "" {
				return fmt.Errorf(
					"resource group target #%d resource #%d resource_type is empty or missing. Please check your configuration",
					resourceGroupIndex+1, resourceIndex+1)
			}

			if len(resource.aggregations) > 0 {
				if !areTargetAggregationsValid(resource.aggregations) {
					return fmt.Errorf("resource group target #%d resource #%d aggregations contain invalid aggregation/s. "+
						"The valid aggregations are: %s", resourceGroupIndex, resourceIndex, strings.Join(getPossibleAggregations(), ", "))
				}
			}
		}
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) checkSubscriptionTargetValidation() error {
	for index, target := range ammr.Targets.subscriptionTargets {
		if target.resourceType == "" {
			return fmt.Errorf(
				"subscription target #%d resource_type is empty or missing. Please check your configuration", index+1)
		}

		if len(target.aggregations) > 0 {
			if !areTargetAggregationsValid(target.aggregations) {
				return fmt.Errorf("subscription target #%d aggregations contain invalid aggregation/s. "+
					"The valid aggregations are: %s", index, strings.Join(getPossibleAggregations(), ", "))
			}
		}
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) addPrefixToResourceTargetsResourceID() {
	for _, target := range ammr.Targets.ResourceTargets {
		target.ResourceID = "/subscriptions/" + ammr.subscriptionID + "/" + target.ResourceID
	}
}

// CreateResourceTargetsFromResourceGroupTargets creates resource targets from resource group targets.
func (ammr *AzureMonitorMetricsReceiver) CreateResourceTargetsFromResourceGroupTargets() error {
	if len(ammr.Targets.resourceGroupTargets) == 0 {
		return nil
	}

	for _, target := range ammr.Targets.resourceGroupTargets {
		if err := ammr.createResourceTargetFromResourceGroupTarget(target); err != nil {
			return fmt.Errorf("error creating resource targets from resource group target %s: %v", target.resourceGroup, err)
		}
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) createResourceTargetFromResourceGroupTarget(target *ResourceGroupTarget) error {
	resourceTargetsCreatedNum := 0
	filter := createClientResourcesFilter(target.resources)
	responses, err := ammr.AzureClients.ResourcesClient.ListByResourceGroup(ammr.AzureClients.Ctx, target.resourceGroup,
		&armresources.ClientListByResourceGroupOptions{Filter: &filter})
	if err != nil {
		return err
	}

	for _, response := range responses {
		currentResourceTargetsCreatedNum, err := ammr.createResourceTargetFromTargetResources(response.Value, target.resources)
		if err != nil {
			return fmt.Errorf("error creating resource target from resource group target resources: %v", err)
		}

		resourceTargetsCreatedNum += currentResourceTargetsCreatedNum
	}

	return nil
}

// CreateResourceTargetsFromSubscriptionTargets creates resource targets from subscription targets.
func (ammr *AzureMonitorMetricsReceiver) CreateResourceTargetsFromSubscriptionTargets() error {
	if len(ammr.Targets.subscriptionTargets) == 0 {
		return nil
	}

	resourceTargetsCreatedNum := 0
	filter := createClientResourcesFilter(ammr.Targets.subscriptionTargets)
	responses, err := ammr.AzureClients.ResourcesClient.List(ammr.AzureClients.Ctx, &armresources.ClientListOptions{Filter: &filter})
	if err != nil {
		return err
	}

	for _, response := range responses {
		currentResourceTargetsCreatedNum, err := ammr.createResourceTargetFromTargetResources(response.Value, ammr.Targets.subscriptionTargets)
		if err != nil {
			return fmt.Errorf("error creating resource target from subscription targets: %v", err)
		}

		resourceTargetsCreatedNum += currentResourceTargetsCreatedNum
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) createResourceTargetFromTargetResources(resources []*armresources.GenericResourceExpanded, targetResources []*Resource) (int, error) {
	resourceTargetsCreatedNum := 0

	for _, targetResource := range targetResources {
		isResourceTargetCreated := false

		for _, resource := range resources {
			resourceID, err := getResourcesClientResourceID(resource)
			if err != nil {
				return resourceTargetsCreatedNum, err
			}

			resourceType, err := getResourcesClientResourceType(resource)
			if err != nil {
				return resourceTargetsCreatedNum, err
			}

			if *resourceType != targetResource.resourceType {
				continue
			}

			ammr.Targets.ResourceTargets = append(ammr.Targets.ResourceTargets, NewResourceTarget(*resourceID, targetResource.metrics, targetResource.aggregations))
			isResourceTargetCreated = true
			resourceTargetsCreatedNum++
		}

		if !isResourceTargetCreated {
			return resourceTargetsCreatedNum, fmt.Errorf("could not find resources with resource type %s", targetResource.resourceType)
		}
	}

	return resourceTargetsCreatedNum, nil
}

// CheckResourceTargetsMetricsValidation checks resource targets metrics validation.
func (ammr *AzureMonitorMetricsReceiver) CheckResourceTargetsMetricsValidation() error {
	for _, target := range ammr.Targets.ResourceTargets {
		if len(target.Metrics) > 0 {
			response, err := ammr.getMetricDefinitionsResponse(target.ResourceID)
			if err != nil {
				return fmt.Errorf("error getting metric definitions response for resource target %s: %v", target.ResourceID, err)
			}

			if err = target.checkMetricsValidation(response.Value); err != nil {
				return fmt.Errorf("error checking resource target %s metrics: %v", target.ResourceID, err)
			}
		}
	}

	return nil
}

// SetResourceTargetsMetrics sets resource targets metrics if their metrics array is empty.
func (ammr *AzureMonitorMetricsReceiver) SetResourceTargetsMetrics() error {
	for _, target := range ammr.Targets.ResourceTargets {
		if len(target.Metrics) > 0 {
			continue
		}

		response, err := ammr.getMetricDefinitionsResponse(target.ResourceID)
		if err != nil {
			return fmt.Errorf("error getting metric definitions response for resource target %s: %v", target.ResourceID, err)
		}

		if err = target.setMetrics(response.Value); err != nil {
			return fmt.Errorf("error setting resource target %s metrics: %v", target.ResourceID, err)
		}
	}

	ammr.changeResourceTargetsMetricsWithComma()
	return nil
}

func (ammr *AzureMonitorMetricsReceiver) changeResourceTargetsMetricsWithComma() {
	for _, target := range ammr.Targets.ResourceTargets {
		target.changeMetricsWithComma()
	}
}

// SplitResourceTargetsMetricsByMinTimeGrain splits resource targets metrics by min time grain.
func (ammr *AzureMonitorMetricsReceiver) SplitResourceTargetsMetricsByMinTimeGrain() error {
	for _, target := range ammr.Targets.ResourceTargets {
		if err := ammr.splitResourceTargetMetricsByMinTimeGrain(target); err != nil {
			return fmt.Errorf("error checking resource target %s metrics min time grain: %v", target.ResourceID, err)
		}
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) splitResourceTargetMetricsByMinTimeGrain(target *ResourceTarget) error {
	response, err := ammr.getMetricDefinitionsResponse(target.ResourceID)
	if err != nil {
		return fmt.Errorf("error getting metric definitions response for resource target %s: %v", target.ResourceID, err)
	}

	timeGrainsMetricsMap, err := target.createResourceTargetTimeGrainsMetricsMap(response.Value)
	if err != nil {
		return fmt.Errorf("error creating resource target time grains metrics map: %v", err)
	}

	if len(timeGrainsMetricsMap) == 1 {
		return nil
	}

	var firstTimeGrain string

	for timeGrain := range timeGrainsMetricsMap {
		firstTimeGrain = timeGrain
		break
	}

	for timeGrain, metrics := range timeGrainsMetricsMap {
		if timeGrain == firstTimeGrain {
			target.Metrics = metrics
			continue
		}

		newTargetAggregations := make([]string, 0)
		newTargetAggregations = append(newTargetAggregations, target.Aggregations...)
		ammr.Targets.ResourceTargets = append(ammr.Targets.ResourceTargets, NewResourceTarget(target.ResourceID, metrics, newTargetAggregations))
	}

	return nil
}

func (ammr *AzureMonitorMetricsReceiver) getMetricDefinitionsResponse(resourceID string) (*armmonitor.MetricDefinitionsClientListResponse, error) {
	response, err := ammr.AzureClients.MetricDefinitionsClient.List(ammr.AzureClients.Ctx, resourceID, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing metric definitions for the resource target %s: %v", resourceID, err)
	}

	if len(response.Value) == 0 {
		return nil, fmt.Errorf("metric definitions response is bad formatted: Value is empty")
	}

	return &response, nil
}

// SplitResourceTargetsWithMoreThanMaxMetrics splits resource targets with more than max metrics.
func (ammr *AzureMonitorMetricsReceiver) SplitResourceTargetsWithMoreThanMaxMetrics() {
	for _, target := range ammr.Targets.ResourceTargets {
		if len(target.Metrics) <= MaxMetricsPerRequest {
			continue
		}

		for start := MaxMetricsPerRequest; start < len(target.Metrics); start += MaxMetricsPerRequest {
			end := start + MaxMetricsPerRequest

			if end > len(target.Metrics) {
				end = len(target.Metrics)
			}

			newTargetMetrics := target.Metrics[start:end]
			newTargetAggregations := make([]string, 0)
			newTargetAggregations = append(newTargetAggregations, target.Aggregations...)
			newTarget := NewResourceTarget(target.ResourceID, newTargetMetrics, newTargetAggregations)
			ammr.Targets.ResourceTargets = append(ammr.Targets.ResourceTargets, newTarget)
		}

		target.Metrics = target.Metrics[:MaxMetricsPerRequest]
	}
}

// SetResourceTargetsAggregations sets resource targets aggregations if their aggregations array is empty.
func (ammr *AzureMonitorMetricsReceiver) SetResourceTargetsAggregations() {
	for _, target := range ammr.Targets.ResourceTargets {
		if len(target.Aggregations) == 0 {
			target.setAggregations()
		}
	}
}

func (arc *azureResourcesClient) List(ctx context.Context, options *armresources.ClientListOptions) ([]*armresources.ClientListResponse, error) {
	responses := make([]*armresources.ClientListResponse, 0)
	pager := arc.client.NewListPager(options)

	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &response)
	}

	return responses, nil
}

func (arc *azureResourcesClient) ListByResourceGroup(
	ctx context.Context,
	resourceGroup string,
	options *armresources.ClientListByResourceGroupOptions,
) ([]*armresources.ClientListByResourceGroupResponse, error) {
	responses := make([]*armresources.ClientListByResourceGroupResponse, 0)
	pager := arc.client.NewListByResourceGroupPager(resourceGroup, options)

	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &response)
	}

	return responses, nil
}

func (rt *ResourceTarget) setMetrics(metricDefinitions []*armmonitor.MetricDefinition) error {
	for _, metricDefinition := range metricDefinitions {
		metricNameValue, err := getMetricDefinitionsClientMetricNameValue(metricDefinition)
		if err != nil {
			return err
		}

		rt.Metrics = append(rt.Metrics, *metricNameValue)
	}

	return nil
}

func (rt *ResourceTarget) setAggregations() {
	rt.Aggregations = append(rt.Aggregations, getPossibleAggregations()...)
}

func (rt *ResourceTarget) checkMetricsValidation(metricDefinitions []*armmonitor.MetricDefinition) error {
	for _, metric := range rt.Metrics {
		isMetricExist := false

		for _, metricDefinition := range metricDefinitions {
			metricNameValue, err := getMetricDefinitionsClientMetricNameValue(metricDefinition)
			if err != nil {
				return err
			}

			if metric == *metricNameValue {
				isMetricExist = true
				break
			}
		}

		if !isMetricExist {
			return fmt.Errorf("resource target has invalid metric %s. Please check your resource targets, "+
				"resource group targets and subscription targets in your configuration", metric)
		}
	}

	return nil
}

func (rt *ResourceTarget) createResourceTargetTimeGrainsMetricsMap(metricDefinitions []*armmonitor.MetricDefinition) (map[string][]string, error) {
	timeGrainsMetrics := make(map[string][]string)

	for _, metric := range rt.Metrics {
		for _, metricDefinition := range metricDefinitions {
			metricNameValue, err := getMetricDefinitionsClientMetricNameValue(metricDefinition)
			if err != nil {
				return nil, err
			}

			if metric == *metricNameValue {
				metricMinTimeGrain, err := getMetricDefinitionsMetricMinTimeGrain(metricDefinition)
				if err != nil {
					return nil, err
				}

				if _, found := timeGrainsMetrics[*metricMinTimeGrain]; !found {
					timeGrainsMetrics[*metricMinTimeGrain] = []string{metric}
				} else {
					timeGrainsMetrics[*metricMinTimeGrain] = append(timeGrainsMetrics[*metricMinTimeGrain], metric)
				}
			}
		}
	}

	return timeGrainsMetrics, nil
}

func (rt *ResourceTarget) changeMetricsWithComma() {
	for index := 0; index < len(rt.Metrics); index++ {
		rt.Metrics[index] = strings.Replace(rt.Metrics[index], ",", "%2", -1)
	}
}

func getPossibleAggregations() []string {
	possibleAggregations := make([]string, 0)

	for _, aggregation := range armmonitor.PossibleAggregationTypeEnumValues() {
		possibleAggregations = append(possibleAggregations, string(aggregation))
	}

	return possibleAggregations
}

func areTargetAggregationsValid(targetAggregations []string) bool {
	for _, targetAggregation := range targetAggregations {
		isTargetAggregationValid := false

		for _, aggregation := range getPossibleAggregations() {
			if targetAggregation == aggregation {
				isTargetAggregationValid = true
				break
			}
		}

		if !isTargetAggregationValid {
			return false
		}
	}

	return true
}

func createClientResourcesFilter(resources []*Resource) string {
	var filter string
	resourcesSize := len(resources)

	for index, resource := range resources {
		if index+1 == resourcesSize {
			filter += "resourceType eq " + "'" + resource.resourceType + "'"
		} else {
			filter += "resourceType eq " + "'" + resource.resourceType + "'" + " or "
		}
	}

	return filter
}

func getResourcesClientResourceID(resource *armresources.GenericResourceExpanded) (*string, error) {
	if resource == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource is missing")
	}

	if resource.ID == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource ID is missing")
	}

	return resource.ID, nil
}

func getResourcesClientResourceType(resource *armresources.GenericResourceExpanded) (*string, error) {
	if resource == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource is missing")
	}

	if resource.Type == nil {
		return nil, fmt.Errorf("resources client response is bad formatted: resource Type is missing")
	}

	return resource.Type, nil
}

func getMetricDefinitionsClientMetricNameValue(metricDefinition *armmonitor.MetricDefinition) (*string, error) {
	if metricDefinition == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition is missing")
	}

	metricName := metricDefinition.Name
	if metricName == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition Name is missing")
	}

	metricNameValue := metricName.Value
	if metricNameValue == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition Name.Value is missing")
	}

	return metricNameValue, nil
}

func getMetricDefinitionsMetricMinTimeGrain(metricDefinition *armmonitor.MetricDefinition) (*string, error) {
	if metricDefinition == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition is missing")
	}

	if len(metricDefinition.MetricAvailabilities) == 0 {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition MetricAvailabilities is empty")
	}

	metricAvailability := metricDefinition.MetricAvailabilities[0]
	if metricDefinition.MetricAvailabilities[0] == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition MetricAvailabilities[0] is missing")
	}

	timeGrain := metricAvailability.TimeGrain
	if timeGrain == nil {
		return nil, fmt.Errorf("metric definitions client response is bad formatted: metric definition MetricAvailabilities[0].TimeGrain is missing")
	}

	return timeGrain, nil
}
