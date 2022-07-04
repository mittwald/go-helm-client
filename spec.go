package helmclient

import (
	"sigs.k8s.io/yaml"
)

const fmtAnnoationKey = "g0-he1m-c1ient__%s"

// common annoation's key
const (
	// walkaround & wait for https://github.com/helm/helm/issues/11049
	ManagedBy = "manged-by"

	// the user defined transparent values, will retrieve in future
	Transparent = "transparent"
)

// GetValuesMap returns the mapped out values of a chart
func (spec *ChartSpec) GetValuesMap() (map[string]interface{}, error) {
	var values map[string]interface{}

	err := yaml.Unmarshal([]byte(spec.ValuesYaml), &values)
	if err != nil {
		return nil, err
	}

	for k, v := range spec.Annotations {
		values[toAnnotationKey(k)] = v
	}

	return values, nil
}
