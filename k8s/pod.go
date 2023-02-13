package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodOpt func(*PodSpec)

type PodSpec struct {
	Name      string
	Namespace string
	Image     string
	Spec      corev1.PodTemplateSpec
}

func NewPodSpec(name string, opts ...PodOpt) corev1.PodTemplateSpec {
	pod := &PodSpec{
		Spec: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{},
			},
		},
	}

	for _, v := range opts {
		v(pod)
	}

	return pod.Spec
}

func PodLabel(key, value string) PodOpt {
	return func(p *PodSpec) {
		p.Spec.ObjectMeta.Labels = map[string]string{
			key: value,
		}
	}
}

func PodContainer(c corev1.Container) PodOpt {
	return func(p *PodSpec) {
		p.Spec.Spec.Containers = append(p.Spec.Spec.Containers, c)
	}
}
