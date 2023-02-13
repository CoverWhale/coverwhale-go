package main

import (
	"fmt"
	"log"

	"github.com/CoverWhale/coverwhale-go/k8s"
	"sigs.k8s.io/yaml"
)

func printYaml(i interface{}) {
	o, err := yaml.Marshal(i)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("---\n%s\n", o)
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
}
