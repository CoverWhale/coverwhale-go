package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Container struct {
	corev1.Container
}

type ContainerOpt func(*Container)

func NewContainer(name string, opts ...ContainerOpt) Container {
	c := Container{
		corev1.Container{
			Name: name,
		},
	}

	for _, v := range opts {
		v(&c)
	}

	return c
}

func ContainerImage(image string) ContainerOpt {
	return func(c *Container) {
		c.Image = image
	}
}

func ContainerEnvVar(key, value string) ContainerOpt {
	return func(c *Container) {
		c.Env = append(c.Env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
}

func ContainerEnvFromSecret(secret, name, key string) ContainerOpt {
	return func(c *Container) {
		c.Env = append(c.Env, corev1.EnvVar{
			Name: name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secret,
					},
					Key: key,
				},
			},
		})
	}
}

func ContainerEnvFromConfigMap(configmap, name, key string) ContainerOpt {
	return func(c *Container) {
		c.Env = append(c.Env, corev1.EnvVar{
			Name: name,
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configmap,
					},
					Key: key,
				},
			},
		})
	}
}

func ContainerCommands(commands []string) ContainerOpt {
	return func(c *Container) {
		c.Command = commands
	}
}

func ContainerImagePullPolicy(policy corev1.PullPolicy) ContainerOpt {
	return func(c *Container) {
		c.ImagePullPolicy = policy
	}
}

func ContainerArgs(args []string) ContainerOpt {
	return func(c *Container) {
		c.Args = args
	}
}

func ContainerPort(name string, port int) ContainerOpt {
	return func(c *Container) {
		c.Ports = append(c.Ports, corev1.ContainerPort{
			Name:          name,
			ContainerPort: int32(port),
		})
	}
}

func ContainerVolume(path string, pv PersistentVolume) ContainerOpt {
	return func(c *Container) {
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			MountPath: path,
			Name:      pv.ObjectMeta.Name,
		})
	}
}

type HTTPProbe struct {
	Path          string
	Port          int
	IntialDelay   int
	PeriodSeconds int
}

func ContainerLivenessProbeHTTP(h HTTPProbe) ContainerOpt {
	return func(c *Container) {
		c.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: h.Path,
					Port: intstr.FromInt(h.Port),
				},
			},
			InitialDelaySeconds: int32(h.IntialDelay),
			PeriodSeconds:       int32(h.PeriodSeconds),
		}
	}
}
