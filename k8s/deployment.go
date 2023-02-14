package k8s

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrNameRequired = fmt.Errorf("name is required")

type Deployment struct {
	appsv1.Deployment
}

type DeploymentOpt func(*Deployment)

func NewDeployment(name string, depOpts ...DeploymentOpt) *Deployment {
	dep := &Deployment{
		appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: make(map[string]string),
				},
				Template: corev1.PodTemplateSpec{},
			},
		},
	}

	for _, v := range depOpts {
		v(dep)
	}

	return dep

}

func DeploymentNamespace(n string) DeploymentOpt {
	return func(d *Deployment) {
		d.ObjectMeta.Namespace = n
	}
}

func DeploymentSelector(key, value string) DeploymentOpt {
	return func(d *Deployment) {
		d.Spec.Selector.MatchLabels[key] = value
	}
}

func DeploymentPodSpec(p PodSpec) DeploymentOpt {
	return func(d *Deployment) {
		d.Spec.Template = p.Spec
	}
}

func DeploymentReplicas(r int) DeploymentOpt {
	replicas := int32(r)
	return func(d *Deployment) {
		d.Spec.Replicas = &replicas
	}
}
