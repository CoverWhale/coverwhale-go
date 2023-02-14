package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Secret struct {
	corev1.Secret
}

type SecretOpt func(*Secret)

func NewSecret(name string, opts ...SecretOpt) Secret {
	s := Secret{
		Secret: corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	for _, v := range opts {
		v(&s)
	}

	return s
}

func SecretNamespace(n string) SecretOpt {
	return func(s *Secret) {
		s.Namespace = n
	}
}

func SecretData(key string, value []byte) SecretOpt {
	return func(s *Secret) {
		s.Data = map[string][]byte{
			key: value,
		}
	}
}
