package k8s

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Ingress struct {
	networkingv1.Ingress
}

type IngressOpt func(*Ingress)

func NewIngress(name string, opts ...IngressOpt) Ingress {
	i := Ingress{
		Ingress: networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: newObjectMeta(name),
		},
	}

	for _, v := range opts {
		v(&i)
	}

	return i
}

func IngressNamespace(n string) IngressOpt {
	return func(i *Ingress) {
		i.Ingress.ObjectMeta.Namespace = n
	}
}

func IngressClass(c string) IngressOpt {
	return func(i *Ingress) {
		i.Ingress.Spec.IngressClassName = &c
	}
}

type Rule struct {
	Host  string
	Paths []Path
	TLS   bool
}

type Path struct {
	Name    string
	Service string
	Port    int
	Type    networkingv1.PathType
}

func IngressRule(r Rule) IngressOpt {
	var paths []networkingv1.HTTPIngressPath
	for _, v := range r.Paths {
		paths = append(paths, networkingv1.HTTPIngressPath{
			Path:     v.Name,
			PathType: &v.Type,
			Backend: networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: v.Service,
					Port: networkingv1.ServiceBackendPort{
						Number: int32(v.Port),
					},
				},
			},
		})
	}

	return func(i *Ingress) {
		if r.TLS {
			addAnnotation("cert-manager.io/cluster-issuer", "letsencrypt-prod", &i.ObjectMeta)
			i.Spec.TLS = append(i.Spec.TLS, networkingv1.IngressTLS{
				Hosts:      []string{r.Host},
				SecretName: fmt.Sprintf("%s-tls", i.Ingress.ObjectMeta.Name),
			})
		}
		i.Ingress.Spec.Rules = append(i.Ingress.Spec.Rules, networkingv1.IngressRule{
			Host: r.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	}
}
