package k8s

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

func MarshalYaml(i interface{}) (string, error) {
	o, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("---\n%s\n", o), nil
}
