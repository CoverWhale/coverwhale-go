package main

import (
	"fmt"
	"log"

	"github.com/CoverWhale/coverwhale-go/k8s"
)

func printYaml(i interface{}) {
	data, err := k8s.MarshalYaml(i)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(data)
}

func main() {

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
