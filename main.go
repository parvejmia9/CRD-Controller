package main

import (
	"context"
	"flag"
	myv1 "github.com/parvejmia9/CRD/pkg/apis/reader.com/v1"
	myclient "github.com/parvejmia9/CRD/pkg/generated/clientset/versioned"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main() {
	log.Println("Configuring KubeConfig")
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	crdClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	customCRD := v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bookstores.reader.com",
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: "reader.com",
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1.JSONSchemaProps{
										"name": {
											Type: "string",
										},
										"replicas": {
											Type: "integer",
										},
										"container": {
											Type: "object",
											Properties: map[string]v1.JSONSchemaProps{
												"image": {
													Type: "string",
												},
												"port": {
													Type: "string",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Scope: "Namespaced",
			Names: v1.CustomResourceDefinitionNames{
				Kind:       "BookStore",
				Plural:     "bookstores",
				Singular:   "bookstore",
				ShortNames: []string{"bk"},
				Categories: []string{"all"},
			},
		},
	}
	ctx := context.TODO()
	// deleting existing CR
	_ = crdClient.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, customCRD.Name, metav1.DeleteOptions{})
	time.Sleep(1 * time.Second)
	// creating new CR
	_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, &customCRD, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
	log.Println("CRD is Created!")
	time.Sleep(100 * time.Millisecond)

	log.Println("Creating BookStore...")

	client, err := myclient.NewForConfig(config)
	bookStoreOBJ := myv1.BookStore{
		ObjectMeta: metav1.ObjectMeta{
			Name: "book",
		},
		Spec: myv1.BookStoreSpec{
			Name:     "my-book-store",
			Replicas: intPtr(3),
			Container: myv1.ContainerSpec{
				Image: "parvejmia9/api-server:0.0.4",
				Port:  "9090",
			},
		},
	}
	_, err = client.BookStoreV1().BookStores("default").Create(ctx, &bookStoreOBJ, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	time.Sleep(2 * time.Second)
	log.Println("BookStore Created!!")

}

func intPtr(i int32) *int32 {
	return &i
}
