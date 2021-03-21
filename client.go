package helmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var storage = repo.File{}

const (
	defaultCachePath            = "/tmp/.helmcache"
	defaultRepositoryConfigPath = "/tmp/.helmrepo"
)

// New returns a new Helm client with the provided options
func New(options *Options) (Client, error) {
	settings := cli.New()

	err := setEnvSettings(options, settings)
	if err != nil {
		return nil, err
	}

	return newClient(options, settings.RESTClientGetter(), settings)
}

// NewClientFromKubeConf returns a new Helm client constructed with the provided kubeconfig options
func NewClientFromKubeConf(options *KubeConfClientOptions) (Client, error) {
	settings := cli.New()
	if options.KubeConfig == nil {
		return nil, fmt.Errorf("kubeconfig missing")
	}

	clientGetter := NewRESTClientGetter(options.Namespace, options.KubeConfig, nil)
	err := setEnvSettings(options.Options, settings)
	if err != nil {
		return nil, err
	}

	if options.KubeContext != "" {
		settings.KubeContext = options.KubeContext
	}

	return newClient(options.Options, clientGetter, settings)
}

// NewClientFromRestConf returns a new Helm client constructed with the provided REST config options
func NewClientFromRestConf(options *RestConfClientOptions) (Client, error) {
	settings := cli.New()

	clientGetter := NewRESTClientGetter(options.Namespace, nil, options.RestConfig)

	err := setEnvSettings(options.Options, settings)
	if err != nil {
		return nil, err
	}

	return newClient(options.Options, clientGetter, settings)
}

// newClient returns a new Helm client via the provided options and REST config
func newClient(options *Options, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (Client, error) {
	err := setEnvSettings(options, settings)
	if err != nil {
		return nil, err
	}

	debugLog := options.DebugLog
	if debugLog == nil {
		debugLog = func(format string, v ...interface{}) {
			log.Printf(format, v...)
		}
	}

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		clientGetter,
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)
	if err != nil {
		return nil, err
	}

	return &HelmClient{
		Settings:     settings,
		Providers:    getter.All(settings),
		storage:      &storage,
		ActionConfig: actionConfig,
		linting:      options.Linting,
		DebugLog:     debugLog,
	}, nil
}

// setEnvSettings sets the client's environment settings based on the provided client configuration
func setEnvSettings(options *Options, settings *cli.EnvSettings) error {
	if options == nil {
		options = &Options{
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

// AddOrUpdateChartRepo adds or updates the provided helm chart repository
func (c *HelmClient) AddOrUpdateChartRepo(entry repo.Entry) error {
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
		c.DebugLog("WARNING: repository name %q already exists", entry.Name)
		return nil
	}

	c.storage.Update(&entry)
	err = c.storage.WriteFile(c.Settings.RepositoryConfig, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// UpdateChartRepos updates the list of chart repositories stored in the client's cache
func (c *HelmClient) UpdateChartRepos() error {
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

	return c.storage.WriteFile(c.Settings.RepositoryConfig, 0o644)
}

// InstallOrUpgradeChart triggers the installation of the provided chart.
// If the chart is already installed, trigger an upgrade instead
func (c *HelmClient) InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) error {
	installed, err := c.chartIsInstalled(spec.ReleaseName)
	if err != nil {
		return err
	}

	if installed {
		return c.upgrade(ctx, spec)
	}
	return c.install(spec)
}

// DeleteChartFromCache deletes the provided chart from the client's cache
func (c *HelmClient) DeleteChartFromCache(spec *ChartSpec) error {
	return c.deleteChartFromCache(spec)
}

// UninstallRelease uninstalls the provided release
func (c *HelmClient) UninstallRelease(spec *ChartSpec) error {
	return c.uninstallRelease(spec)
}

// install lints and installs the provided chart
func (c *HelmClient) install(spec *ChartSpec) error {
	client := action.NewInstall(c.ActionConfig)
	mergeInstallOptions(spec, client)

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

	c.DebugLog("release installed successfully: %s/%s-%s", rel.Name, rel.Name, rel.Chart.Metadata.Version)

	return nil
}

// upgrade upgrades a chart and CRDs
func (c *HelmClient) upgrade(ctx context.Context, spec *ChartSpec) error {
	client := action.NewUpgrade(c.ActionConfig)
	mergeUpgradeOptions(spec, client)

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
		c.DebugLog("updating crds")
		err = c.upgradeCRDs(ctx, helmChart)
		if err != nil {
			return err
		}
	}

	rel, err := client.Run(spec.ReleaseName, helmChart, values)
	if err != nil {
		return err
	}

	c.DebugLog("release upgrade successfully: %s/%s-%s", rel.Name, rel.Name, rel.Chart.Metadata.Version)

	return nil
}

// deleteChartFromCache deletes the provided chart from the client's cache
func (c *HelmClient) deleteChartFromCache(spec *ChartSpec) error {
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

	c.DebugLog("chart removed successfully: %s/%s-%s", helmChart.Name(), spec.ReleaseName, helmChart.AppVersion())

	return nil
}

// uninstallRelease uninstalls the provided release
func (c *HelmClient) uninstallRelease(spec *ChartSpec) error {
	client := action.NewUninstall(c.ActionConfig)

	mergeUninstallReleaseOptions(spec, client)

	resp, err := client.Run(spec.ReleaseName)
	if err != nil {
		return err
	}

	c.DebugLog("release removed, response: %v", resp)

	return nil
}

// lint lints a chart's values
func (c *HelmClient) lint(chartPath string, values map[string]interface{}) error {
	client := action.NewLint()

	result := client.Run([]string{chartPath}, values)

	for _, err := range result.Errors {
		c.DebugLog("Error %s", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("linting for chartpath %q failed", chartPath)
	}

	return nil
}

func (c *HelmClient) TemplateChart(spec *ChartSpec) ([]byte, error) {
	client := action.NewInstall(c.ActionConfig)
	mergeInstallOptions(spec, client)

	client.DryRun = true
	client.ReleaseName = "RELEASE-NAME"
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = []string{}
	client.IncludeCRDs = true

	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return nil, err
	}

	if helmChart.Metadata.Type != "" && helmChart.Metadata.Type != "application" {
		return nil, fmt.Errorf(
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
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	values, err := spec.GetValuesMap()
	if err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	rel, err := client.Run(helmChart, values)

	// We ignore a potential error here because, when the --debug flag was specified,
	// we always want to print the YAML, even if it is not valid. The error is still returned afterwards.
	if rel != nil {
		var manifests bytes.Buffer
		fmt.Fprintln(&manifests, strings.TrimSpace(rel.Manifest))
		if !client.DisableHooks {
			for _, m := range rel.Hooks {
				fmt.Fprintf(&manifests, "---\n# Source: %s\n%s\n", m.Path, m.Manifest)
			}
		}

		// if we have a list of files to render, then check that each of the
		// provided files exists in the chart.
		fmt.Fprintf(out, "%s", manifests.String())
	}

	return out.Bytes(), err
}

func (c *HelmClient) LintChart(spec *ChartSpec) error {
	_, chartPath, err := c.getChart(spec.ChartName, &action.ChartPathOptions{})
	if err != nil {
		return err
	}

	values, err := spec.GetValuesMap()
	if err != nil {
		return err
	}

	return c.lint(chartPath, values)
}

func (c *HelmClient) SetDebugLog(debugLog action.DebugLog) {
	c.DebugLog = debugLog
}

// upgradeCRDs upgrades the CRDs of the provided chart
func (c *HelmClient) upgradeCRDs(ctx context.Context, chartInstance *chart.Chart) error {
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
			existingCRDObj, err := k8sClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, &crdObj, metaV1.UpdateOptions{})
			if err != nil {
				return err
			}

		case "apiextensions.k8s.io/v1beta1":
			var crdObj v1beta1.CustomResourceDefinition
			err = json.Unmarshal(jsonCRD, &crdObj)
			if err != nil {
				return err
			}
			existingCRDObj, err := k8sClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(ctx, crdObj.Name, metaV1.GetOptions{})
			if err != nil {
				return err
			}
			crdObj.ResourceVersion = existingCRDObj.ResourceVersion
			_, err = k8sClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(ctx, &crdObj, metaV1.UpdateOptions{})
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("failed to update crd %q: unsupported api-version %q", crd.Name, meta.APIVersion)
		}
	}

	return nil
}

// getChart returns a chart matching the provided chart name and options
func (c *HelmClient) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
	chartPath, err := chartPathOptions.LocateChart(chartName, c.Settings)
	if err != nil {
		return nil, "", err
	}

	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, "", err
	}

	if helmChart.Metadata.Deprecated {
		c.DebugLog("WARNING: This chart (%q) is deprecated", helmChart.Metadata.Name)
	}

	return helmChart, chartPath, err
}

// chartIsInstalled checks whether a chart is already installed or not by the provided release name
func (c *HelmClient) chartIsInstalled(release string) (bool, error) {
	histClient := action.NewHistory(c.ActionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(release); err == driver.ErrReleaseNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// mergeInstallOptions merges values of the provided chart to helm install options used by the client
func mergeInstallOptions(chartSpec *ChartSpec, installOptions *action.Install) {
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
	installOptions.DryRun = chartSpec.DryRun
	installOptions.SubNotes = chartSpec.SubNotes
}

// mergeUpgradeOptions merges values of the provided chart to helm upgrade options used by the client
func mergeUpgradeOptions(chartSpec *ChartSpec, upgradeOptions *action.Upgrade) {
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
	upgradeOptions.DryRun = chartSpec.DryRun
	upgradeOptions.SubNotes = chartSpec.SubNotes
}

// mergeUninstallReleaseOptions merges values of the provided chart to helm uninstall options used by the client
func mergeUninstallReleaseOptions(chartSpec *ChartSpec, uninstallReleaseOptions *action.Uninstall) {
	uninstallReleaseOptions.DisableHooks = chartSpec.DisableHooks
	uninstallReleaseOptions.Timeout = chartSpec.Timeout
}
