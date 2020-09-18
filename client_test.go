package helmclient

import (
	"context"

	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/client-go/rest"
)

func ExampleNew() {
	opt := &Options{
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
	}

	helmClient, err := New(opt)
	if err != nil {
		panic(err)
	}
	_ = helmClient
}

func ExampleNewClientFromRestConf() {
	opt := &RestConfClientOptions{
		Options: &Options{
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true,
		},
		RestConfig: &rest.Config{},
	}

	helmClient, err := NewClientFromRestConf(opt)
	if err != nil {
		panic(err)
	}
	_ = helmClient
}

func ExampleNewClientFromKubeConf() {
	opt := &KubeConfClientOptions{
		Options: &Options{
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true,
		},
		KubeContext: "",
		KubeConfig:  []byte{},
	}

	helmClient, err := NewClientFromKubeConf(opt)
	if err != nil {
		panic(err)
	}
	_ = helmClient
}

func ExampleHelmClient_AddOrUpdateChartRepo_public() {
	// Define a public chart repository
	chartRepo := repo.Entry{
		Name: "stable",
		URL:  "https://kubernetes-charts.storage.googleapis.com",
	}

	// Add a chart-repository to the client
	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_AddOrUpdateChartRepo_private() {
	// Define a private chart repository
	chartRepo := repo.Entry{
		Name:     "stable",
		URL:      "https://private-chartrepo.somedomain.com",
		Username: "foo",
		Password: "bar",
	}

	// Add a chart-repository to the client
	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_InstallOrUpgradeChart() {
	// Define the chart to be installed
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_DeleteChartFromCache() {
	// Define the chart to be deleted from the client's cache
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if err := helmClient.DeleteChartFromCache(&chartSpec); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_UpdateChartRepos() {
	if err := helmClient.UpdateChartRepos(); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_UninstallRelease() {
	// Define the released chart to be installed
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if err := helmClient.UninstallRelease(&chartSpec); err != nil {
		panic(err)
	}
}
