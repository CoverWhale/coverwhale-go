package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PersistentVolume struct {
	corev1.PersistentVolume
}

type PersistentVolumeOpt func(*PersistentVolume)

func NewPersistentVolume(name string, opts ...PersistentVolumeOpt) PersistentVolume {
	pv := PersistentVolume{
		PersistentVolume: corev1.PersistentVolume{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PersistentVolume",
				APIVersion: "v1",
			},
			ObjectMeta: newObjectMeta(name),
			Spec:       corev1.PersistentVolumeSpec{},
		},
	}

	for _, v := range opts {
		v(&pv)
	}

	return pv
}

func PersistentvolumeCapacity(capacity resource.Quantity) PersistentVolumeOpt {
	return func(pv *PersistentVolume) {
		if pv.Spec.Capacity == nil {
			pv.Spec.Capacity = corev1.ResourceList{
				"capacity": capacity,
			}
		}
	}
}

func PersistentVolumeHostPath(path string, pathType corev1.HostPathType) PersistentVolumeOpt {
	return func(pv *PersistentVolume) {
		pv.Spec.HostPath = &corev1.HostPathVolumeSource{
			Path: path,
			Type: &pathType,
		}
	}
}

func PersistentVolumeLocal(path string, fsType string) PersistentVolumeOpt {
	return func(pv *PersistentVolume) {
		pv.Spec.Local = &corev1.LocalVolumeSource{
			Path:   path,
			FSType: &fsType,
		}
	}
}
