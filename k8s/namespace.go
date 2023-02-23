package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespace struct {
	corev1.Namespace
}

type NamespaceOpt func(*Namespace)

func NewNamespace(name string, opts ...NamespaceOpt) Namespace {
	ns := Namespace{
		corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: newObjectMeta(name),
		},
	}

	for _, v := range opts {
		v(&ns)
	}

	return ns
}

func NamespaceAnnotation(key, value string) NamespaceOpt {
	return func(n *Namespace) {
		addAnnotation(key, value, &n.ObjectMeta)
	}
}

func NamespaceAnnotations(annotations map[string]string) NamespaceOpt {
	return func(n *Namespace) {
		for k, v := range annotations {
			addAnnotation(k, v, &n.ObjectMeta)
		}
	}
}
