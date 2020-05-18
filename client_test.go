package helmclient

import (
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

	// Construct a new Helm client, where '_' is the constructed client.
	// Change this empty assignment to get access to the client's services.
	_, err := New(opt)
	if err != nil {
		panic(err)
	}
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

	// Construct a new Helm client via REST configuration, where '_' is the constructed client.
	// Change this empty assignment to get access to the client's services.
	_, err := NewClientFromRestConf(opt)
	if err != nil {
		panic(err)
	}
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

	// Construct a new Helm client via KubeConf, where '_' is the constructed client.
	// Change this empty assignment to get access to the client's services.
	_, err := NewClientFromKubeConf(opt)
	if err != nil {
		panic(err)
	}
}

func ExampleHelmClient_AddOrUpdateChartRepo_public() {
	// Dummy assignment
	// Construct a real Helm client via New(), NewClientFromRestConf(), or NewClientFromKubeConf()
	helmClient := &HelmClient{}

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
	// Dummy assignment
	// Construct a real Helm client via New(), NewClientFromRestConf(), or NewClientFromKubeConf()
	helmClient := &HelmClient{}

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
	// Dummy assignment
	// Construct a real Helm client via New(), NewClientFromRestConf(), or NewClientFromKubeConf()
	helmClient := &HelmClient{}

	// Define the chart to be installed
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if err := helmClient.InstallOrUpgradeChart(&chartSpec); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_DeleteChartFromCache() {
	// Dummy assignment
	// Construct a real Helm client via New(), NewClientFromRestConf(), or NewClientFromKubeConf()
	helmClient := &HelmClient{}

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
	// Dummy assignment
	// Construct a real Helm client via New(), NewClientFromRestConf(), or NewClientFromKubeConf()
	helmClient := &HelmClient{}

	if err := helmClient.UpdateChartRepos(); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_UninstallRelease() {
	// Dummy assignment
	// Construct a real Helm client via New(), NewClientFromRestConf(), or NewClientFromKubeConf()
	helmClient := &HelmClient{}

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
