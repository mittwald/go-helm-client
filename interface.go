package helmclient

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

//go:generate mockgen -source=interface.go -package mockhelmclient -destination=./mock/interface.go -self_package=. Client

type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error)
	ListDeployedReleases() ([]*release.Release, error)
	GetRelease(name string) (*release.Release, error)
	RollbackRelease(spec *ChartSpec, version int) error
	GetReleaseValues(name string, allValues bool) (map[string]interface{}, error)
	UninstallRelease(spec *ChartSpec) error
	UninstallReleaseByName(name string) error
	TemplateChart(spec *ChartSpec) ([]byte, error)
	LintChart(spec *ChartSpec) error
	SetDebugLog(debugLog action.DebugLog)
}
