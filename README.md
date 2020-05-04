# Go Helm Client

Go client library for accessing [Helm](https://github.com/helm/helm), enabling the user to programmatically change helm charts and releases.
                                                                                                                         
This library is build upon `helm/v3.1.2`
## Usage

```go
import "github.com/mittwald/go-helm-client"
```
Construct a new Helm client, then use the various services on the client to manage helm chart repositories and releases:
```go 
package main

import (
	"github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
)

func main() {
	// Create a client
	helmClient, err := helm.NewClient(&helm.ClientOptions{
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
	})
	if err != nil {
		panic(err)
	}

	// Add needed chart-repos to the client
	err = helmClient.AddOrUpdateChartRepo(repo.Entry{
		Name: "stable",
		URL:  "https://kubernetes-charts.storage.googleapis.com",
	})
	if err != nil {
		panic(err)
	}

	// Define the chart you want to install
	chartSpec := helm.ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	err = helmClient.InstallOrUpgradeChart(&chartSpec)
	if err != nil {
		panic(err)
	}
}
```

Alternatively, you can create a client via REST config: 
```go
helmClient, err := helm.NewClientFromRestConf(&restClientOpts)
```
or via Kubeconfig:

```go
helmClient, err := helm.NewClientFromKubeConf()
```

#### Private chart repository
When working with private repositories, you can utilize the `Username` and `Password` parameters of a chart entry to specify credentials, e.g.:

```go
err := helmClient.AddOrUpdateChartRepo(repo.Entry{
    Name: "stable",
    URL:  "https://private-chart.somedomain.com",
    Username: "foo",
    Password: "bar",
})
```
