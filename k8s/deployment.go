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
			ObjectMeta: newObjectMeta(name),
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
		setNamespace(n, &d.ObjectMeta)
	}
}

func DeploymentSelector(key, value string) DeploymentOpt {
	return func(d *Deployment) {
		metav1.AddLabelToSelector(d.Spec.Selector, key, value)
	}
}

func DeploymentLabel(key, value string) DeploymentOpt {
	return func(d *Deployment) {
		addLabel(key, value, &d.ObjectMeta)
	}
}

func DeploymentLabels(labels map[string]string) DeploymentOpt {
	return func(d *Deployment) {
		for k, v := range labels {
			addLabel(k, v, &d.ObjectMeta)
		}
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
