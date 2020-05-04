package go_helm_client

import (
	"time"

	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// KubeConfClientOptions defines the options used for constructing a client via kubeconfig
type KubeConfClientOptions struct {
	*ClientOptions
	KubeContext string
	KubeConfig  []byte
}

// RestConfClientOptions defines the options used for constructing a client via REST config
type RestConfClientOptions struct {
	*ClientOptions
	RestConfig *rest.Config
}

// ClientOptions defines the options of a client
type ClientOptions struct {
	Namespace        string
	RepositoryConfig string
	RepositoryCache  string
	Debug            bool
	Linting          bool
}

// RESTClientGetter defines the values of a helm REST client
type RESTClientGetter struct {
	namespace  string
	kubeConfig []byte
	restConfig *rest.Config
}

// Client defines the values of a helm client
type Client struct {
	Settings     *cli.EnvSettings
	Providers    getter.Providers
	storage      *repo.File
	ActionConfig *action.Configuration
	linting      bool
}

// ChartSpec defines the values of a helm chart
type ChartSpec struct {
	ReleaseName string `json:"release"`
	ChartName   string `json:"chart"`
	Namespace   string `json:"namespace"`

	// use string instead of map[string]interface{}
	// https://github.com/kubernetes-sigs/kubebuilder/issues/528#issuecomment-466449483
	// and https://github.com/kubernetes-sigs/controller-tools/pull/317
	// +optional
	ValuesYaml string `json:"valuesYaml,omitempty"`

	// +optional
	Version string `json:"version,omitempty"`

	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`

	// +optional
	Replace bool `json:"replace,omitempty"`

	// +optional
	Wait bool `json:"wait,omitempty"`

	// +optional
	DependencyUpdate bool `json:"dependencyUpdate,omitempty"`

	// +optional
	Timeout time.Duration `json:"timeout,omitempty"`

	// +optional
	GenerateName bool `json:"generateName,omitempty"`

	// +optional
	NameTemplate string `json:"NameTemplate,omitempty"`

	// +optional
	Atomic bool `json:"atomic,omitempty"`

	// +optional
	SkipCRDs bool `json:"skipCRDs,omitempty"`

	// +optional
	UpgradeCRDs bool `json:"upgradeCRDs,omitempty"`

	// +optional
	SubNotes bool `json:"subNotes,omitempty"`

	// +optional
	Force bool `json:"force,omitempty"`

	// +optional
	ResetValues bool `json:"resetValues,omitempty"`

	// +optional
	ReuseValues bool `json:"reuseValues,omitempty"`

	// +optional
	Recreate bool `json:"recreate,omitempty"`

	// +optional
	MaxHistory int `json:"maxHistory,omitempty"`

	// +optional
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`
}
