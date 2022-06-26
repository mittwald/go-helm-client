package helmclient

import (
	"sigs.k8s.io/yaml"
)

// transparentKey represents the key for save transparent values
const transparentKey = "transparent__go-helm-client"

// walkAroundCustomLabelKey walkaround & wait for https://github.com/helm/helm/issues/11049
const walkAroundCustomLabelKey = "walk-around-custom-label__go-helm-client"
const walkAroundCustomLabelValue = "ok"

// GetValuesMap returns the mapped out values of a chart
func (spec *ChartSpec) GetValuesMap() (map[string]interface{}, error) {
	var values map[string]interface{}

	err := yaml.Unmarshal([]byte(spec.ValuesYaml), &values)
	if err != nil {
		return nil, err
	}

	if spec.Transparent != "" {
		values[transparentKey] = spec.Transparent
	}
	values[walkAroundCustomLabelKey] = walkAroundCustomLabelValue

	return values, nil
}
