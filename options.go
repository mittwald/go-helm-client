package helmclient

import (
	"helm.sh/helm/v3/pkg/action"
)

// ListOptions represents the options for list releases
type ListOptions struct {
	Namespace string
	States    action.ListStates

	//label.Selector
	Selectors map[string]string

	// name filter  case-insensitive
	Filter string
}
