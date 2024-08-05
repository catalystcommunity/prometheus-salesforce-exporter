# prometheus-salesforce-exporter

The prometheus-salesforce-exporter is a go implementation of a Prometheus
exporter for getting metrics out of Salesforce. Currently, the exporter only
exports the following metrics.

```
salesforce_daily_api_requests_limit
salesforce_daily_api_requests_remaining
salesforce_data_storage_mb_limit
salesforce_data_storage_mb_remaining
```

There are other implementations of a Salesforce exporter out there. We built
this exporter to be concious of the Salesforce API limits, so the exporter
polls metrics from Salesforce independently from when Prometheus scrapes.


## Installing with Helm

A helm chart for this project is maintained [here](https://github.com/catalystcommunity/chart-prometheus-salesforce-exporter)

```
helm repo add catalystcommunity https://raw.githubusercontent.com/catalystcommunity/charts/main
helm install catalystcommunity/prometheus-salesforce-exporter \
  --set salesforce.baseUrl=https://myenv.salesforce.com \
  --set salesforce.clientId=abc123 \
  --set salesforce.clientSecret=ABC123 \
  --set salesforce.username=myuser@example.com \
  --set salesforce.password=securePassword
```


## Contributing

Contributions welcome! Create an issue or a PR!
