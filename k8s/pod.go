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

func NewPodSpec(name string, opts ...PodOpt) PodSpec {
	pod := PodSpec{
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
		v(&pod)
	}

	return pod
}

func PodLabel(key, value string) PodOpt {
	return func(p *PodSpec) {
		p.Spec.ObjectMeta.Labels = map[string]string{
			key: value,
		}
	}
}

func PodContainer(c Container) PodOpt {
	return func(p *PodSpec) {
		p.Spec.Spec.Containers = append(p.Spec.Spec.Containers, c.Container)
	}
}

func PodInitContainer(c Container) PodOpt {
	return func(p *PodSpec) {
		p.Spec.Spec.InitContainers = append(p.Spec.Spec.InitContainers, c.Container)
	}
}

// func PodVolume(name string, pv PersistentVolume) PodOpt {
// 	return func(p *PodSpec) {
// 		p.Spec.Spec.Volumes = append(p.Spec.Spec.Volumes, corev1.Volume{
// 			Name: name,
// 			VolumeSource: corev1.VolumeSource{
// 				pv.PersistentVolume,
// 			},
// 		})
// 	}
// }
