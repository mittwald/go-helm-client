package helmclient

import (
	"time"

	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// Type Guard asserting that HelmClient satisfies the HelmClient interface.
var _ Client = &HelmClient{}

// KubeConfClientOptions defines the options used for constructing a client via kubeconfig.
type KubeConfClientOptions struct {
	*Options
	KubeContext string
	KubeConfig  []byte
}

// RestConfClientOptions defines the options used for constructing a client via REST config.
type RestConfClientOptions struct {
	*Options
	RestConfig *rest.Config
}

// Options defines the options of a client.
type Options struct {
	Namespace        string
	RepositoryConfig string
	RepositoryCache  string
	Debug            bool
	Linting          bool
	DebugLog         action.DebugLog
}

// RESTClientGetter defines the values of a helm REST client.
type RESTClientGetter struct {
	namespace  string
	kubeConfig []byte
	restConfig *rest.Config
}

// HelmClient Client defines the values of a helm client.
type HelmClient struct {
	// Settings defines the environment settings of a client.
	Settings  *cli.EnvSettings
	Providers getter.Providers
	storage   *repo.File
	// ActionConfig is the helm action configuration.
	ActionConfig *action.Configuration
	linting      bool
	DebugLog     action.DebugLog
}

// ChartSpec defines the values of a helm chart
type ChartSpec struct {
	ReleaseName string `json:"release"`
	ChartName   string `json:"chart"`
	// Namespace where the chart release is deployed.
	// Note that helmclient.Options.Namespace should ideally match the namespace configured here.
	Namespace string `json:"namespace"`
	// ValuesYaml is the values.yaml content.
	// use string instead of map[string]interface{}
	// https://github.com/kubernetes-sigs/kubebuilder/issues/528#issuecomment-466449483
	// and https://github.com/kubernetes-sigs/controller-tools/pull/317
	// +optional
	ValuesYaml string `json:"valuesYaml,omitempty"`
	// Version of the chart release.
	// +optional
	Version string `json:"version,omitempty"`
	// CreateNamespace indicates whether to create the namespace if it does not exist.
	// +optional
	CreateNamespace bool `json:"createNamespace,omitempty"`
	// DisableHooks indicates whether to disable hooks.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`
	// Replace indicates whether to replace the chart release if it already exists.
	// +optional
	Replace bool `json:"replace,omitempty"`
	// Wait indicates whether to wait for the release to be deployed or not.
	// +optional
	Wait bool `json:"wait,omitempty"`
	// DependencyUpdate indicates whether to update the chart release if the dependencies have changed.
	// +optional
	DependencyUpdate bool `json:"dependencyUpdate,omitempty"`
	// Timeout configures the time to wait for any individual Kubernetes operation (like Jobs for hooks).
	// +optional
	Timeout time.Duration `json:"timeout,omitempty"`
	// GenerateName indicates that the release name should be generated.
	// +optional
	GenerateName bool `json:"generateName,omitempty"`
	// NameTemplate is the template used to generate the release name if GenerateName is configured.
	// +optional
	NameTemplate string `json:"NameTemplate,omitempty"`
	// Atomic indicates whether to install resources atomically.
	// 'Wait' will automatically be set to true when using Atomic.
	// +optional
	Atomic bool `json:"atomic,omitempty"`
	// SkipCRDs indicates whether to skip CRDs during installation.
	// +optional
	SkipCRDs bool `json:"skipCRDs,omitempty"`
	// Upgrade indicates whether to perform a CRD upgrade during installation.
	// +optional
	UpgradeCRDs bool `json:"upgradeCRDs,omitempty"`
	// SubNotes indicates whether to print sub-notes.
	// +optional
	SubNotes bool `json:"subNotes,omitempty"`
	// Force indicates whether to force the operation.
	// +optional
	Force bool `json:"force,omitempty"`
	// ResetValues indicates whether to reset the values.yaml file during installation.
	// +optional
	ResetValues bool `json:"resetValues,omitempty"`
	// ReuseValues indicates whether to reuse the values.yaml file during installation.
	// +optional
	ReuseValues bool `json:"reuseValues,omitempty"`
	// Recreate indicates whether to recreate the release if it already exists.
	// +optional
	Recreate bool `json:"recreate,omitempty"`
	// MaxHistory limits the maximum number of revisions saved per release.
	// +optional
	MaxHistory int `json:"maxHistory,omitempty"`
	// CleanupOnFail indicates whether to cleanup the release on failure.
	// +optional
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`
	// DryRun indicates whether to perform a dry run.
	// +optional
	DryRun bool `json:"dryRun,omitempty"`
}
