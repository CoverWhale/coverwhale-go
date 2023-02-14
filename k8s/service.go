package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Service struct {
	corev1.Service
}

type ServiceOpt func(*Service)

func NewService(name string, opts ...ServiceOpt) Service {
	service := Service{
		Service: corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.ServiceSpec{
				Selector: make(map[string]string),
			},
		},
	}

	for _, v := range opts {
		v(&service)
	}

	return service
}

func ServiceNamespace(n string) ServiceOpt {
	return func(s *Service) {
		s.ObjectMeta.Namespace = n
	}
}

func ServicePort(port, targetPort int) ServiceOpt {
	return func(s *Service) {
		s.Spec.Ports = append(s.Spec.Ports, corev1.ServicePort{
			Port:       int32(port),
			TargetPort: intstr.FromInt(targetPort),
		})
	}
}

func ServiceSelector(key, value string) ServiceOpt {
	return func(s *Service) {
		s.Spec.Selector[key] = value
	}
}
