package helmclient

import (
	"sigs.k8s.io/yaml"
)

const fmtAnnoationKey = "g0-he1m-c1ient__%s"

// GetValuesMap returns the mapped out values of a chart
func (spec *ChartSpec) GetValuesMap() (map[string]interface{}, error) {
	var values map[string]interface{}

	err := yaml.Unmarshal([]byte(spec.ValuesYaml), &values)
	if err != nil {
		return nil, err
	}

	if values == nil {
		values = map[string]interface{}{}
	}

	for k, v := range spec.Annotations {
		values[toAnnotationKey(k)] = v
	}

	return values, nil
}
