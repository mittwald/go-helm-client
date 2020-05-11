package helmclient

import (
	"helm.sh/helm/v3/pkg/repo"
)

type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(spec *ChartSpec) error
	DeleteChartFromCache(spec *ChartSpec) error
	UninstallRelease(spec *ChartSpec) error
}
