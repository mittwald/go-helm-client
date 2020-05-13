package helmclient

import (
	"helm.sh/helm/v3/pkg/repo"
)

//go:generate mockgen -source=interface.go -package=mockhelmclient -destination=./mock/interface.go -self_package=. Client

type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(spec *ChartSpec) error
	DeleteChartFromCache(spec *ChartSpec) error
	UninstallRelease(spec *ChartSpec) error
}
