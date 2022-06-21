package helmclient

import (
	"sigs.k8s.io/yaml"
)

// RawValuesKey represents the key for save raw values
const RawValuesKey = "raw_values"

// GetValuesMap returns the mapped out values of a chart
func (spec *ChartSpec) GetValuesMap() (map[string]interface{}, error) {
	var values map[string]interface{}

	err := yaml.Unmarshal([]byte(spec.ValuesYaml), &values)
	if err != nil {
		return nil, err
	}

	if spec.SaveRawValues {
		values[RawValuesKey] = spec.ValuesYaml
	}

	return values, nil
}
