package helmclient

import (
	"helm.sh/helm/v3/pkg/repo"
)

//go:generate mockgen -source=interface.go -destination=./mock/interface.go -package=mockhelmclient -self_package=.

type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(spec *ChartSpec) error
	DeleteChartFromCache(spec *ChartSpec) error
	UninstallRelease(spec *ChartSpec) error
}
