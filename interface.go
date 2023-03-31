package helmclient

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

//go:generate mockgen -source=interface.go -package mockhelmclient -destination=./mock/interface.go -self_package=. Client

// Client holds the method signatures for a Helm client.
// NOTE: This is an interface to allow for mocking in tests.
type Client interface {
	AddOrUpdateChartRepo(entry repo.Entry) error
	UpdateChartRepos() error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec, opts *GenericHelmOptions) (*release.Release, error)
	InstallChart(ctx context.Context, spec *ChartSpec, opts *GenericHelmOptions) (*release.Release, error)
	UpgradeChart(ctx context.Context, spec *ChartSpec, opts *GenericHelmOptions) (*release.Release, error)
	ListDeployedReleases() ([]*release.Release, error)
	ListReleasesByStateMask(action.ListStates) ([]*release.Release, error)
	ListReleases(opts ListOptions) ([]*release.Release, error)
	GetRelease(name string) (*release.Release, error)
	// RollBack is an interface to abstract a rollback action.
	RollBack
	GetReleaseValues(name string, allValues bool) (map[string]interface{}, error)
	UninstallRelease(spec *ChartSpec) error
	UninstallReleaseByName(name string) error
	TemplateChart(spec *ChartSpec, options *HelmTemplateOptions) ([]byte, error)
	LintChart(spec *ChartSpec) error
	SetDebugLog(debugLog action.DebugLog)
	ListReleaseHistory(name string, max int) ([]*release.Release, error)
	GetChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error)

	// AnnotationWithRelease returns the annotation value bind to the key in the release if not existed, returns ""
	AnnotationWithRelease(rel *release.Release, key string) interface{}
}

type RollBack interface {
	RollbackRelease(spec *ChartSpec) error
}
