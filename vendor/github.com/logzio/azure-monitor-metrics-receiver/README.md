# Azure Monitor Metrics Receiver

An SDK wrapper that uses Azure Monitor SDK. 
Lets you receive metrics of Azure resources using Azure Monitor API in 3 ways:

- Resource Target
- Resource Group Target
- Subscription Target

## Azure Credential

Uses `client_id`, `client_secret` and `tenant_id` for authentication (access token),
and `subscription_id` for accessing Azure resources.

**Here how to find each of them:**

`subscription_id` can be found under **Overview**->**Essentials** in the Azure portal for your application/service.

`client_id` and `client_secret` can be obtained by registering an application under Azure Active Directory.

`tenant_id` can be found under **Azure Active Directory**->**Properties**.

## Resource Target

get metrics of a specific resource.

```go
type ResourceTarget struct {
	ResourceID   string
	Metrics      []string
	Aggregations []string
}
```

`ResourceID` can be found under **Overview**->**Essentials**->**JSON View** (link) in the Azure
portal for your application/service.

Must start with 'resourceGroups/...' ('/subscriptions/xxxxxxxx-xxxx-xxxx-xxx-xxxxxxxxxxxx' must be removed from the 
beginning of Resource ID property value)

`Metrics` is an array of the name of the metrics that you want to collect. 
**Pay attention:** all metrics should be valid metrics of the resource target.

* If the array is empty, all available metrics of the resource target will be collected.

`Aggregations` is an array of the metrics aggregation type value to collect. The available aggregations are:

- Total
- Count
- Average
- Minimum
- Maximum

* If the array is empty, all aggregation types values will be collected for each metric.

## Resource Group Target

get metrics of resources under specific resource group, using resource types.

```go
type ResourceGroupTarget struct {
    resourceGroup string
    resources     []*Resource
}
```

`resourceGroup` is the name of the resource group.

`resources` is an array of resources that are under the same `resourceGroup`.

* Info about `Resource` can be found in Subscription Target section.

## Subscription Target

get metrics of resources under the subscription, using resource types.

```go
type Resource struct {
    resourceType string
    metrics      []string
    aggregations []string
}
```

`resourceType` is the type of resources you want to collect metrics of.

* Info about `metrics` and `aggregations` can be found in Resource Target section.
