package helmclient

import (
	"sigs.k8s.io/yaml"
)

// transparentKey represents the key for save transparent values
const transparentKey = "transparent__go-helm-client"

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

	return values, nil
}
