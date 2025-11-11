package azuremonitormetricsreceiver

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

const (
	minMetricsFields = 2

	// MetricFieldTimeStamp is timeStamp metric field name.
	MetricFieldTimeStamp = "timeStamp"
	// MetricFieldTotal is total metric field name.
	MetricFieldTotal     = "total"
	// MetricFieldAverage is average metric field name.
	MetricFieldAverage   = "average"
	// MetricFieldCount is count metric field name.
	MetricFieldCount     = "count"
	// MetricFieldMinimum is minimum metric field name.
	MetricFieldMinimum   = "minimum"
	// MetricFieldMaximum is maximum metric field name.
	MetricFieldMaximum   = "maximum"

	// MetricTagSubscriptionID is subscription ID metric tag name.
	MetricTagSubscriptionID = "subscription_id"
	// MetricTagResourceGroup is resource group metric tag name.
	MetricTagResourceGroup  = "resource_group"
	// MetricTagResourceName is resource name metric tag name.
	MetricTagResourceName   = "resource_name"
	// MetricTagNamespace is namespace metric tag name.
	MetricTagNamespace      = "namespace"
	// MetricTagResourceRegion is resource region metric tag name.
	MetricTagResourceRegion = "resource_region"
	// MetricTagUnit is unit metric tag name.
	MetricTagUnit           = "unit"
)

// CollectResourceTargetMetrics collects metrics of a resource target.
func (ammr *AzureMonitorMetricsReceiver) CollectResourceTargetMetrics(target *ResourceTarget) ([]*Metric, []string, error) {
	metricNames := strings.Join(target.Metrics, ",")
	aggregations := strings.Join(target.Aggregations, ",")
	response, err := ammr.AzureClients.MetricsClient.List(ammr.AzureClients.Ctx, target.ResourceID, &armmonitor.MetricsClientListOptions{
		Metricnames: &metricNames,
		Aggregation: &aggregations,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error listing metrics for the resource target %s: %v", target.ResourceID, err)
	}

	metrics, notCollectedMetrics, err := collectMetrics(&response)
	if err != nil {
		return nil, nil, fmt.Errorf("error collecting resource target %s metrics: %v", target.ResourceID, err)
	}

	return metrics, notCollectedMetrics, nil
}

func collectMetrics(response *armmonitor.MetricsClientListResponse) ([]*Metric, []string, error) {
	metrics := make([]*Metric, 0)
	notCollectedMetric := make([]string, 0)

	for _, metric := range response.Value {
		errorMessage, err := getMetricsClientMetricErrorMessage(metric)
		if err != nil {
			return nil, nil, err
		}

		if errorMessage != nil {
			return nil, nil, fmt.Errorf("response error: %s", *errorMessage)
		}

		if len(metric.Timeseries) == 0 {
			metricID, err := getMetricsClientMetricID(metric)
			if err != nil {
				return nil, nil, err
			}

			notCollectedMetric = append(notCollectedMetric, *metricID)
			continue
		}

		timeseries := metric.Timeseries[0]
		if timeseries == nil {
			return nil, nil, fmt.Errorf("metrics client response is bad formatted: metric timeseries is missing")
		}

		if len(timeseries.Data) == 0 {
			metricID, err := getMetricsClientMetricID(metric)
			if err != nil {
				return nil, nil, err
			}

			notCollectedMetric = append(notCollectedMetric, *metricID)
			continue
		}

		metricName, err := createMetricName(metric, response)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating metric name: %v", err)
		}

		metricFields := getMetricFields(timeseries.Data)
		if metricFields == nil {
			metricID, err := getMetricsClientMetricID(metric)
			if err != nil {
				return nil, nil, err
			}

			notCollectedMetric = append(notCollectedMetric, *metricID)
			continue
		}

		metricTags, err := getMetricTags(metric, response)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting metric tags: %v", err)
		}

		metrics = append(metrics, &Metric{
			Name:   *metricName,
			Fields: metricFields,
			Tags:   metricTags,
		})
	}

	return metrics, notCollectedMetric, nil
}

func getMetricsClientMetricErrorMessage(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	if metric.ErrorCode == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ErrorCode is missing")
	}

	if *metric.ErrorCode == "Success" {
		return nil, nil
	}

	errorMessage := fmt.Sprintf("error code %s: %s", *metric.ErrorCode, *metric.ErrorMessage)
	return &errorMessage, nil
}

func getMetricsClientMetricID(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	if metric.ID == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ID is missing")
	}

	return metric.ID, nil
}

func getMetricsClientMetricNameLocalizedValue(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	metricName := metric.Name
	if metricName == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric Name is missing")
	}

	metricNameLocalizedValue := metricName.LocalizedValue
	if metricNameLocalizedValue == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric Name.LocalizedValue is missing")
	}

	return metricNameLocalizedValue, nil
}

func getMetricsClientMetricUnit(metric *armmonitor.Metric) (*string, error) {
	if metric == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric is missing")
	}

	if metric.Unit == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric Unit is missing")
	}

	metricUnit := string(*metric.Unit)
	return &metricUnit, nil
}

func getMetricsClientResponseNamespace(response *armmonitor.MetricsClientListResponse) (*string, error) {
	if response.Namespace == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: reponse Namespace is missing")
	}

	return response.Namespace, nil
}

func getMetricsClientResponseResourceRegion(response *armmonitor.MetricsClientListResponse) (*string, error) {
	if response.Resourceregion == nil {
		return nil, fmt.Errorf("metrics client response is bad formatted: reponse Resourceregion is missing")
	}

	return response.Resourceregion, nil
}

func getMetricsClientMetricValueFields(metricValue *armmonitor.MetricValue) map[string]interface{} {
	if metricValue == nil {
		return nil
	}

	if metricValue.TimeStamp == nil {
		return nil
	}

	metricFields := make(map[string]interface{})
	metricValueFieldsNum := 1

	metricFields[MetricFieldTimeStamp] = metricValue.TimeStamp.Format("2006-01-02T15:04:05Z07:00")

	if metricValue.Total != nil {
		metricFields[MetricFieldTotal] = *metricValue.Total
		metricValueFieldsNum++
	}

	if metricValue.Average != nil {
		metricFields[MetricFieldAverage] = *metricValue.Average
		metricValueFieldsNum++
	}

	if metricValue.Count != nil {
		metricFields[MetricFieldCount] = *metricValue.Count
		metricValueFieldsNum++
	}

	if metricValue.Minimum != nil {
		metricFields[MetricFieldMinimum] = *metricValue.Minimum
		metricValueFieldsNum++
	}

	if metricValue.Maximum != nil {
		metricFields[MetricFieldMaximum] = *metricValue.Maximum
		metricValueFieldsNum++
	}

	if metricValueFieldsNum < minMetricsFields {
		return nil
	}

	return metricFields
}

func createMetricName(metric *armmonitor.Metric, response *armmonitor.MetricsClientListResponse) (*string, error) {
	namespace, err := getMetricsClientResponseNamespace(response)
	if err != nil {
		return nil, err
	}

	name, err := getMetricsClientMetricNameLocalizedValue(metric)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(".", "_", "/", "_", " ", "_", "(", "_", ")", "_")
	metricName := fmt.Sprintf("azure_monitor_%s_%s",
		replacer.Replace(strings.ToLower(*namespace)),
		replacer.Replace(strings.ToLower(*name)))

	return &metricName, nil
}

func getMetricFields(metricValues []*armmonitor.MetricValue) map[string]interface{} {
	for index := len(metricValues) - 1; index >= 0; index-- {
		metricFields := getMetricsClientMetricValueFields(metricValues[index])
		if metricFields == nil {
			continue
		}

		return metricFields
	}

	return nil
}

func getMetricTags(metric *armmonitor.Metric, response *armmonitor.MetricsClientListResponse) (map[string]string, error) {
	tags := make(map[string]string)
	subscriptionID, err := getMetricSubscriptionID(metric)
	if err != nil {
		return nil, err
	}

	tags[MetricTagSubscriptionID] = *subscriptionID

	resourceGroupName, err := getMetricResourceGroupName(metric)
	if err != nil {
		return nil, err
	}

	tags[MetricTagResourceGroup] = *resourceGroupName

	resourceName, err := getMetricResourceName(metric)
	if err != nil {
		return nil, err
	}

	tags[MetricTagResourceName] = *resourceName

	namespace, err := getMetricsClientResponseNamespace(response)
	if err != nil {
		return nil, err
	}

	tags[MetricTagNamespace] = *namespace

	resourceRegion, err := getMetricsClientResponseResourceRegion(response)
	if err != nil {
		return nil, err
	}

	tags[MetricTagResourceRegion] = *resourceRegion

	unit, err := getMetricsClientMetricUnit(metric)
	if err != nil {
		return nil, err
	}

	tags[MetricTagUnit] = *unit

	return tags, nil
}

func getMetricSubscriptionID(metric *armmonitor.Metric) (*string, error) {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := getPartOfMetricID(*metricID, 0, 2, false)
	if err != nil {
		return nil, err
	}

	return subscriptionID, nil
}

func getMetricResourceGroupName(metric *armmonitor.Metric) (*string, error) {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return nil, err
	}

	resourceGroupName, err := getPartOfMetricID(*metricID, 0, 4, false)
	if err != nil {
		return nil, err
	}

	return resourceGroupName, nil
}

func getMetricResourceName(metric *armmonitor.Metric) (*string, error) {
	metricID, err := getMetricsClientMetricID(metric)
	if err != nil {
		return nil, err
	}

	resourceGroupName, err := getPartOfMetricID(*metricID, 1, 2, true)
	if err != nil {
		return nil, err
	}

	return resourceGroupName, nil
}

func getPartOfMetricID(metricID string, partIndex int, partSubPartIndex int, getPartToEnd bool) (*string, error) {
	metricIDParts := strings.Split(metricID, "/providers/")
	if len(metricIDParts) <= partIndex {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ID is bad formatted")
	}

	resourceIDPartSubParts := strings.Split(metricIDParts[partIndex], "/")
	if len(resourceIDPartSubParts) <= partSubPartIndex {
		return nil, fmt.Errorf("metrics client response is bad formatted: metric ID is bad formatted")
	}

	var part string

	if getPartToEnd {
		part = strings.Join(resourceIDPartSubParts[partSubPartIndex:], "/")
	} else {
		part = resourceIDPartSubParts[partSubPartIndex]
	}

	return &part, nil
}
