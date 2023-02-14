package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMap struct {
	corev1.ConfigMap
}

type ConfigMapOpt func(*ConfigMap)

func NewConfigMap(name string, opts ...ConfigMapOpt) ConfigMap {
	c := ConfigMap{
		ConfigMap: corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: newObjectMeta(name),
			Data:       make(map[string]string),
			BinaryData: make(map[string][]byte),
		},
	}

	for _, v := range opts {
		v(&c)
	}

	return c
}

func ConfigMapNamespace(n string) ConfigMapOpt {
	return func(c *ConfigMap) {
		setNamespace(n, &c.ObjectMeta)
	}
}

func ConfigMapImmutable(b bool) ConfigMapOpt {
	return func(c *ConfigMap) {
		c.ConfigMap.Immutable = &b
	}
}

func ConfigMapData(key, value string) ConfigMapOpt {
	return func(c *ConfigMap) {
		c.ConfigMap.Data[key] = value
	}
}

func ConfigMapDataMap(data map[string]string) ConfigMapOpt {
	return func(c *ConfigMap) {
		for k, v := range data {
			c.ConfigMap.Data[k] = v
		}
	}
}

func ConfigMapBinaryData(key string, value []byte) ConfigMapOpt {
	return func(c *ConfigMap) {
		c.ConfigMap.BinaryData[key] = value
	}
}

func ConfigMapBinaryDataMap(data map[string][]byte) ConfigMapOpt {
	return func(c *ConfigMap) {
		for k, v := range data {
			c.ConfigMap.BinaryData[k] = v
		}
	}
}
