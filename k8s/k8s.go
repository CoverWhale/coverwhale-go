package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func MarshalYaml(i interface{}) (string, error) {
	o, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("---\n%s\n", o), nil
}

func addAnnotation(key, value string, m *metav1.ObjectMeta) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	m.Annotations[key] = value
}

func setNamespace(n string, m *metav1.ObjectMeta) {
	m.Namespace = n
}

func newObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
}

func addLabel(key, value string, m *metav1.ObjectMeta) {
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}

	m.Labels[key] = value
}
