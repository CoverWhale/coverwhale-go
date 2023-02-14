package main

import (
	"fmt"
	"log"

	"github.com/CoverWhale/coverwhale-go/k8s"
	corev1 "k8s.io/api/core/v1"
)

func printYaml(i interface{}) {
	data, err := k8s.MarshalYaml(i)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(data)
}

func main() {

	pv := k8s.NewPersistentVolume("myvolume",
		k8s.PersistentVolumeHostPath("/", corev1.HostPathDirectory),
	)

	printYaml(pv)

	n := k8s.NewNamespace("test",
		k8s.NamespaceAnnotation("test", "test2"),
		k8s.NamespaceAnnotations(map[string]string{
			"hey": "there",
			"yo":  "what's up",
		}),
	)

	printYaml(n)

	multiLine := `this is
a multiline
config
`

	conf := k8s.NewConfigMap("myconfigmap",
		k8s.ConfigMapNamespace("testing"),
		k8s.ConfigMapData("multiline", multiLine),
		k8s.ConfigMapDataMap(map[string]string{
			"testing": "123",
			"hey":     "this is a test",
		}),
		k8s.ConfigMapBinaryData("test", []byte("gimme some bytes")),
	)

	printYaml(conf)

	c := k8s.NewContainer("test",
		k8s.ContainerImage("myrepo/ratings:latest"),
		k8s.ContainerEnvVar("hey", "there"),
		k8s.ContainerEnvFromSecret("testsecret", "thing", "apiKey"),
		k8s.ContainerImagePullPolicy("Always"),
		k8s.ContainerArgs([]string{"server", "start"}),
		k8s.ContainerPort("http", 8080),
		k8s.ContainerPort("https", 443),
	)

	// can also call the options later for conditionals
	f := k8s.ContainerEnvVar("added", "later")
	f(&c)

	p := k8s.NewPodSpec("test",
		k8s.PodLabel("testing", "again"),
		k8s.PodContainer(c),
	)

	d := k8s.NewDeployment("testing",
		k8s.DeploymentNamespace("testing"),
		k8s.DeploymentSelector("app", "testing"),
		k8s.DeploymentPodSpec(p),
		k8s.DeploymentReplicas(3),
	)

	printYaml(d)

	s := k8s.NewService("test",
		k8s.ServiceNamespace("testing"),
		k8s.ServicePort(80, 8080),
		k8s.ServiceSelector("app", "mytest"),
	)
	printYaml(s)

	r := k8s.Rule{
		Host: "test.test.com",
		TLS:  true,
		Paths: []k8s.Path{
			{
				Name:    "/test",
				Service: "test",
				Port:    8080,
				Type:    "PathPrefix",
			},
		},
	}
	i := k8s.NewIngress("test",
		k8s.IngressClass("nginx"),
		k8s.IngressNamespace("testing"),
		k8s.IngressRule(r),
	)

	printYaml(i)

	sec := k8s.NewSecret("test",
		k8s.SecretNamespace("testing"),
		k8s.SecretData("apiKey", []byte("thekey")),
	)

	printYaml(sec)

}
