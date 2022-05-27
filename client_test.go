package helmclient

import (
	"bytes"
	"context"

	"helm.sh/helm/v3/pkg/action"

	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/client-go/rest"
)

func ExampleNew() {
	var outputBuffer bytes.Buffer

	opt := &Options{
		Namespace:        "default", // Change this to the namespace you wish the client to operate in.
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
		DebugLog:         func(format string, v ...interface{}) {},
		Output:           &outputBuffer, // Not mandatory, leave open for default os.Stdout
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
			Namespace:        "default", // Change this to the namespace you wish the client to operate in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
			},
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
			Namespace:        "default", // Change this to the namespace you wish to install the chart in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
			},
		},
		KubeContext: "",
		KubeConfig:  []byte{},
	}

	helmClient, err := NewClientFromKubeConf(opt, Burst(100), Timeout(10e9))
	if err != nil {
		panic(err)
	}
	_ = helmClient
}

func ExampleHelmClient_AddOrUpdateChartRepo_public() {
	// Define a public chart repository.
	chartRepo := repo.Entry{
		Name: "stable",
		URL:  "https://charts.helm.sh/stable",
	}

	// Add a chart-repository to the client.
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
		// Since helm 3.6.1 it is necessary to pass 'PassCredentialsAll = true'.
		PassCredentialsAll: true,
	}

	// Add a chart-repository to the client.
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

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_InstallOrUpgradeChart_useChartDirectory() {
	// Use an unpacked chart directory.
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "/path/to/stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_InstallOrUpgradeChart_useLocalChartArchive() {
	// Use an archived chart directory.
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "/path/to/stable/etcd-operator.tar.gz",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_InstallOrUpgradeChart_useURL() {
	// Use an archived chart directory via URL.
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "http://helm.whatever.com/repo/etcd-operator.tar.gz",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_InstallOrUpgradeChart_useDefaultRollBackStrategy() {
	// Define the chart to be installed
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	// Use the default rollback strategy offer by HelmClient (revert to the previous version).
	opts := GenericHelmOptions{
		RollBack: helmClient,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, &opts); err != nil {
		panic(err)
	}
}

type customRollBack struct {
	HelmClient
}

var _ RollBack = &customRollBack{}

func (c customRollBack) RollbackRelease(spec *ChartSpec) error {
	client := action.NewRollback(c.ActionConfig)

	client.Force = true

	return client.Run(spec.ReleaseName)
}

func ExampleHelmClient_InstallOrUpgradeChart_useCustomRollBackStrategy() {
	// Define the chart to be installed
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	// Use a custom rollback strategy (customRollBack needs to implement RollBack).
	rollBacker := customRollBack{}

	opts := GenericHelmOptions{
		RollBack: rollBacker,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, &opts); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_LintChart() {
	// Define a chart with custom values to be tested.
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
		ValuesYaml: `deployments:
  etcdOperator: true
  backupOperator: false`,
	}

	if err := helmClient.LintChart(&chartSpec); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_TemplateChart() {
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
		ValuesYaml: `deployments:
  etcdOperator: true
  backupOperator: false`,
	}

	_, err := helmClient.TemplateChart(&chartSpec)
	if err != nil {
		panic(err)
	}
}

func ExampleHelmClient_UpdateChartRepos() {
	// Update the list of chart repositories.
	if err := helmClient.UpdateChartRepos(); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_UninstallRelease() {
	// Define the released chart to be installed.
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	// Uninstall the chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if err := helmClient.UninstallRelease(&chartSpec); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_UninstallReleaseByName() {
	// Uninstall a release by name.
	if err := helmClient.UninstallReleaseByName("etcd-operator"); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_ListDeployedReleases() {
	// List all deployed releases.
	if _, err := helmClient.ListDeployedReleases(); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_GetReleaseValues() {
	// Get the values of a deployed release.
	if _, err := helmClient.GetReleaseValues("etcd-operator", true); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_GetRelease() {
	// Get specific details of a deployed release.
	if _, err := helmClient.GetRelease("etcd-operator"); err != nil {
		panic(err)
	}
}

func ExampleHelmClient_RollbackRelease() {
	// Define the released chart to be installed
	chartSpec := ChartSpec{
		ReleaseName: "etcd-operator",
		ChartName:   "stable/etcd-operator",
		Namespace:   "default",
		UpgradeCRDs: true,
		Wait:        true,
	}

	// Rollback to the previous version of the release.
	if err := helmClient.RollbackRelease(&chartSpec); err != nil {
		return
	}
}
