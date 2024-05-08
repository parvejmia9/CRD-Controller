package main

import (
	"flag"
	"github.com/parvejmia9/CRD-Controller/controller"
	clientset "github.com/parvejmia9/CRD-Controller/pkg/generated/clientset/versioned"
	informers "github.com/parvejmia9/CRD-Controller/pkg/generated/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main() {
	log.Println("Configure KubeConfig...")

	var kubeConfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeConfig file")
	} else {
		kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	exampleClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// we will use these factories to create informers to watch on specific resource

	kubeInformationFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformationFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	// new controller is being created
	c := controller.NewController(kubeClient, exampleClient,
		kubeInformationFactory.Apps().V1().Deployments(),
		exampleInformationFactory.BookStore().V1().BookStores(),
	)

	stopCh := make(chan struct{})

	kubeInformationFactory.Start(stopCh)
	exampleInformationFactory.Start(stopCh)

	if err := c.Run(1, stopCh); err != nil {
		log.Println("error running controller:", err)
	}

}
