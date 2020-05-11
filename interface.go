package helmclient

import (
	"helm.sh/helm/v3/pkg/repo"
)

type Client interface {
	New(options *Options) (*Client, error)
	NewClientFromKubeConf(options *KubeConfClientOptions) (*Client, error)
	NewClientFromRestConf(options *RestConfClientOptions) (*Client, error)
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(spec *ChartSpec) error
	DeleteChartFromCache(spec *ChartSpec) error
	UninstallRelease(spec *ChartSpec) error
}
