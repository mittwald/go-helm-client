package helmclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/spf13/pflag"

	"helm.sh/helm/v3/pkg/getter"

	"helm.sh/helm/v3/pkg/cli"

	"helm.sh/helm/v3/pkg/downloader"

	"helm.sh/helm/v3/pkg/chart"

	"helm.sh/helm/v3/pkg/storage/driver"

	"helm.sh/helm/v3/pkg/chart/loader"

	"helm.sh/helm/v3/pkg/action"

	"helm.sh/helm/v3/pkg/repo"
)

var storage = repo.File{}

const (
	defaultCachePath            = "/tmp/.helmcache"
	defaultRepositoryConfigPath = "/tmp/.helmrepo"
)

// New
//
// Wrapper function returning a new Helm client
func New(options *ClientOptions) (*Client, error) {
	settings := cli.New()

	err := setEnvSettings(options, settings)
	if err != nil {
		return nil, err
	}

	return newClient(options, settings.RESTClientGetter(), settings)
}

// NewClientFromKubeConf
//
// Wrapper function returning a new Helm client constructed with the provided kubeconfig options
func NewClientFromKubeConf(options *KubeConfClientOptions) (*Client, error) {
	settings := cli.New()
	if options.KubeConfig == nil {
		return nil, fmt.Errorf("kubeconfig missing")
	}

	clientGetter := NewRESTClientGetter(options.Namespace, options.KubeConfig, nil)
	err := setEnvSettings(options.ClientOptions, settings)
	if err != nil {
		return nil, err
	}

	if options.KubeContext != "" {
		settings.KubeContext = options.KubeContext
	}

	return newClient(options.ClientOptions, clientGetter, settings)
}

// NewClientFromRestConf
//
// Wrapper function returning a new Helm client constructed with the provided REST config options
func NewClientFromRestConf(options *RestConfClientOptions) (*Client, error) {
	settings := cli.New()

	clientGetter := NewRESTClientGetter(options.Namespace, nil, options.RestConfig)

	err := setEnvSettings(options.ClientOptions, settings)
	if err != nil {
		return nil, err
	}

	return newClient(options.ClientOptions, clientGetter, settings)
}

// newClient
//
// Return a new Helm client
func newClient(options *ClientOptions, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (*Client, error) {
	err := setEnvSettings(options, settings)
	if err != nil {
		return nil, err
	}

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		clientGetter,
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {
			log.Printf(format, v)
		},
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		Settings:     settings,
		Providers:    getter.All(settings),
		storage:      &storage,
		ActionConfig: actionConfig,
		linting:      options.Linting,
	}, nil
}

// setEnvSettings
//
// Set the client's environment settings based on the provided client configuration
func setEnvSettings(options *ClientOptions, settings *cli.EnvSettings) error {
	if options == nil {
		options = &ClientOptions{
			RepositoryConfig: defaultRepositoryConfigPath,
			RepositoryCache:  defaultCachePath,
			Linting:          true,
		}
	}

	// set the namespace with this ugly workaround because cli.EnvSettings.namespace is private
	// thank you helm!
	if options.Namespace != "" {
		pflags := pflag.NewFlagSet("", pflag.ContinueOnError)
		settings.AddFlags(pflags)
		err := pflags.Parse([]string{"-n", options.Namespace})
		if err != nil {
			return err
		}
	}

	if options.RepositoryConfig == "" {
		options.RepositoryConfig = defaultRepositoryConfigPath
	}

	if options.RepositoryCache == "" {
		options.RepositoryCache = defaultCachePath
	}

	settings.RepositoryCache = options.RepositoryCache
	settings.RepositoryConfig = defaultRepositoryConfigPath
	settings.Debug = options.Debug

	return nil
}

// AddOrUpdateChartRepo
//
// Add or update the provided helm chart repository
func (c *Client) AddOrUpdateChartRepo(entry repo.Entry) error {
	chartRepo, err := repo.NewChartRepository(&entry, c.Providers)
	if err != nil {
		return err
	}

	chartRepo.CachePath = c.Settings.RepositoryCache

	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return err
	}

	if c.storage.Has(entry.Name) {
		log.Printf("WARNING: repository name %q already exists", entry.Name)
		return nil
	}

	c.storage.Update(&entry)
	err = c.storage.WriteFile(c.Settings.RepositoryConfig, 0644)
	if err != nil {
		return err
	}

	return nil
}

// UpdateChartRepos
//
// Update the list of chart repositories stored in the client's cache
func (c *Client) UpdateChartRepos() error {
	for _, entry := range c.storage.Repositories {
		chartRepo, err := repo.NewChartRepository(entry, c.Providers)
		if err != nil {
			return err
		}

		chartRepo.CachePath = c.Settings.RepositoryCache
		_, err = chartRepo.DownloadIndexFile()
		if err != nil {
			return err
		}

		c.storage.Update(entry)
	}

	return c.storage.WriteFile(c.Settings.RepositoryConfig, 0644)
}

// InstallOrUpgradeChart
//
// Triggers the installation of the provided chart. If the chart is already installed, trigger an upgrade instead
func (c *Client) InstallOrUpgradeChart(spec *ChartSpec) error {
	installed, err := c.chartIsInstalled(spec.ReleaseName)
	if err != nil {
		return err
	}

	if installed {
		return c.upgrade(spec)
	}
	return c.install(spec)
}

// DeleteChartFromCache
//
// Wrapper function for deleting the provided chart from the client's cache
func (c *Client) DeleteChartFromCache(spec *ChartSpec) error {
	return c.deleteChartFromCache(spec)
}

// UninstallRelease
//
// Wrapper function for uninstalling the provided release
func (c *Client) UninstallRelease(spec *ChartSpec) error {
	return c.uninstallRelease(spec)
}

// install
//
// Lint and install the provided chart
func (c *Client) install(spec *ChartSpec) error {
	client := action.NewInstall(c.ActionConfig)
	c.mergeInstallOptions(spec, client)

	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return err
	}

	if helmChart.Metadata.Type != "" && helmChart.Metadata.Type != "application" {
		return fmt.Errorf(
			"chart %q has an unsupported type and is not installable: %q",
			helmChart.Metadata.Name,
			helmChart.Metadata.Type,
		)
	}

	if req := helmChart.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(helmChart, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					ChartPath:        chartPath,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          c.Providers,
					RepositoryConfig: c.Settings.RepositoryConfig,
					RepositoryCache:  c.Settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	values, err := spec.GetValuesMap()
	if err != nil {
		return err
	}

	if c.linting {
		err = c.lint(chartPath, values)
		if err != nil {
			return err
		}
	}

	rel, err := client.Run(helmChart, values)
	if err != nil {
		return err
	}

	log.Printf("release installed successfully: %s/%s-%s", rel.Name, rel.Name, rel.Chart.Metadata.Version)

	return nil
}

// upgrade
//
// Upgrade a chart and CRDs
func (c *Client) upgrade(spec *ChartSpec) error {
	client := action.NewUpgrade(c.ActionConfig)
	c.mergeUpgradeOptions(spec, client)

	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return err
	}

	if req := helmChart.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(helmChart, req); err != nil {
			return err
		}
	}

	values, err := spec.GetValuesMap()
	if err != nil {
		return err
	}

	if c.linting {
		err = c.lint(chartPath, values)
		if err != nil {
			return err
		}
	}

	if !spec.SkipCRDs && spec.UpgradeCRDs {
		log.Printf("updating crds")
		err = c.upgradeCRDs(helmChart)
		if err != nil {
			return err
		}
	}

	rel, err := client.Run(spec.ReleaseName, helmChart, values)
	if err != nil {
		return err
	}

	log.Printf("release upgrade successfully: %s/%s-%s", rel.Name, rel.Name, rel.Chart.Metadata.Version)

	return nil
}

// deleteChartFromCache
//
// Delete the provided chart from the client's cache
func (c *Client) deleteChartFromCache(spec *ChartSpec) error {
	client := action.NewChartRemove(c.ActionConfig)

	helmChart, _, err := c.getChart(spec.ChartName, &action.ChartPathOptions{})
	if err != nil {
		return err
	}

	var deleteOutputBuffer bytes.Buffer
	err = client.Run(&deleteOutputBuffer, helmChart.Name())
	if err != nil {
		return err
	}

	log.Printf("chart removed successfully: %s/%s-%s", helmChart.Name(), spec.ReleaseName, helmChart.AppVersion())

	return nil
}

// uninstallRelease
//
// Uninstall the provided release
func (c *Client) uninstallRelease(spec *ChartSpec) error {
	client := action.NewUninstall(c.ActionConfig)

	c.mergeUninstallReleaseOptions(spec, client)

	resp, err := client.Run(spec.ReleaseName)
	if err != nil {
		return err
	}

	log.Printf("release removed, response: %v", resp)

	return nil
}

// lint
//
// Lint a chart's values
func (c *Client) lint(chartPath string, values map[string]interface{}) error {
	client := action.NewLint()

	result := client.Run([]string{chartPath}, values)

	for _, err := range result.Errors {
		log.Printf("Error %s", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("linting for chartpath %q failed", chartPath)
	}

	return nil
}

// upgradeCRDs
//
// Upgrade the CRDs of the provided chart
func (c *Client) upgradeCRDs(chartInstance *chart.Chart) error {
	cfg, err := c.Settings.RESTClientGetter().ToRESTConfig()
	if err != nil {
		return err
	}

	k8sClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	for _, crd := range chartInstance.CRDObjects() {
		// use this ugly detour to parse the crdYaml to a CustomResourceDefinitions-Object because direct
		// yaml-unmarshalling does not find the correct keys
		jsonCRD, err := yaml.ToJSON(crd.File.Data)
		if err != nil {
			return err
		}

		var meta metaV1.TypeMeta
		err = json.Unmarshal(jsonCRD, &meta)
		if err != nil {
			return err
		}

		switch meta.APIVersion {

		case "apiextensions.k8s.io/apiextensionsV1":
			var crdObj apiextensionsV1.CustomResourceDefinition
			err = json.Unmarshal(jsonCRD, &crdObj)
			if err != nil {
				return err
			}
			existingCRDObj, err := k8sClient.ApiextensionsV1().CustomResourceDefinitions().Get(crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1().CustomResourceDefinitions().Update(&crdObj)
			if err != nil {
				return err
			}

		case "apiextensions.k8s.io/v1beta1":
			var crdObj v1beta1.CustomResourceDefinition
			err = json.Unmarshal(jsonCRD, &crdObj)
			if err != nil {
				return err
			}
			existingCRDObj, err := k8sClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(&crdObj)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("failed to update crd %q: unsupported api-version %q", crd.Name, meta.APIVersion)
		}
	}

	return nil
}

// getChart
//
// Return a chart matching the provided chart name and options
func (c *Client) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
	chartPath, err := chartPathOptions.LocateChart(chartName, c.Settings)
	if err != nil {
		return nil, "", err
	}

	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, "", err
	}

	if helmChart.Metadata.Deprecated {
		log.Printf("WARNING: This chart (%q) is deprecated", helmChart.Metadata.Name)
	}

	return helmChart, chartPath, err
}

// chartIsInstalled
//
// Check whether a chart is already installed or not by the provided release name
func (c *Client) chartIsInstalled(release string) (bool, error) {
	histClient := action.NewHistory(c.ActionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(release); err == driver.ErrReleaseNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// mergeInstallOptions
//
// Merge values of the provided chart to helm install options used by the client
func (c *Client) mergeInstallOptions(chartSpec *ChartSpec, installOptions *action.Install) {
	installOptions.DisableHooks = chartSpec.DisableHooks
	installOptions.Replace = chartSpec.Replace
	installOptions.Wait = chartSpec.Wait
	installOptions.DependencyUpdate = chartSpec.DependencyUpdate
	installOptions.Timeout = chartSpec.Timeout
	installOptions.Namespace = chartSpec.Namespace
	installOptions.ReleaseName = chartSpec.ReleaseName
	installOptions.Version = chartSpec.Version
	installOptions.GenerateName = chartSpec.GenerateName
	installOptions.NameTemplate = chartSpec.NameTemplate
	installOptions.Atomic = chartSpec.Atomic
	installOptions.SkipCRDs = chartSpec.SkipCRDs
	installOptions.SubNotes = chartSpec.SubNotes
}

// mergeUpgradeOptions
//
// Merge values of the provided chart to helm upgrade options used by the client
func (c *Client) mergeUpgradeOptions(chartSpec *ChartSpec, upgradeOptions *action.Upgrade) {
	upgradeOptions.Version = chartSpec.Version
	upgradeOptions.Namespace = chartSpec.Namespace
	upgradeOptions.Timeout = chartSpec.Timeout
	upgradeOptions.Wait = chartSpec.Wait
	upgradeOptions.DisableHooks = chartSpec.DisableHooks
	upgradeOptions.Force = chartSpec.Force
	upgradeOptions.ResetValues = chartSpec.ResetValues
	upgradeOptions.ReuseValues = chartSpec.ReuseValues
	upgradeOptions.Recreate = chartSpec.Recreate
	upgradeOptions.MaxHistory = chartSpec.MaxHistory
	upgradeOptions.Atomic = chartSpec.Atomic
	upgradeOptions.CleanupOnFail = chartSpec.CleanupOnFail
	upgradeOptions.SubNotes = chartSpec.SubNotes
}

// mergeUninstallReleaseOptions
//
// Merge values of the provided chart to helm uninstall options used by the client
func (c *Client) mergeUninstallReleaseOptions(chartSpec *ChartSpec, uninstallReleaseOptions *action.Uninstall) {
	uninstallReleaseOptions.DisableHooks = chartSpec.DisableHooks
	uninstallReleaseOptions.Timeout = chartSpec.Timeout
}
