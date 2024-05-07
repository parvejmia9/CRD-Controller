package controller

import (
	"fmt"
	clientset "github.com/parvejmia9/CRD-Controller/pkg/generated/clientset/versioned"
	informer "github.com/parvejmia9/CRD-Controller/pkg/generated/informers/externalversions/reader.com/v1"
	lister "github.com/parvejmia9/CRD-Controller/pkg/generated/listers/reader.com/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsInformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appsListers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

type Controller struct {
	// standard kubernetes client-set
	KubeClientSet kubernetes.Clientset
	// Client-set for our own api group
	SampleClientSet clientset.Interface

	DeploymentsLister appsListers.DeploymentLister
	DeploymentsSynced cache.InformerSynced
	BookStoreLister   lister.BookStoreLister
	BookStoreSynced   cache.InformerSynced
	// workQueue is a rate limited work queue. This is used to queue work to be processed instead of
	// performing it as soon as a change happens. This means we can ensure we only process a fixed
	// amount of resources at a time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers
	WorkQueue workqueue.RateLimitingInterface
}

// NewController returns a sample controller
func NewController(
	kubectlclientset kubernetes.Clientset,
	sampleclientset clientset.Interface,
	deploymentInformer appsInformer.DeploymentInformer,
	bookstoreInformer informer.BookStoreInformer) *Controller {
	cntrlr := &Controller{
		KubeClientSet:     kubectlclientset,
		SampleClientSet:   sampleclientset,
		DeploymentsLister: deploymentInformer.Lister(),
		DeploymentsSynced: deploymentInformer.Informer().HasSynced,
		BookStoreLister:   bookstoreInformer.Lister(),
		BookStoreSynced:   bookstoreInformer.Informer().HasSynced,
		WorkQueue:         workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}
	log.Println("Setting up event handlers")

	// Setting up an event handler for the changes of BookStore type resources
	bookstoreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cntrlr.enqueueBookStore,
		UpdateFunc: func(oldObj, newObj interface{}) {
			cntrlr.enqueueBookStore(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			cntrlr.enqueueBookStore(obj)
		},
	})
	return cntrlr
}

// this function is used to take kubernetes resources and
// converting it into a unique string and then adding it into the queue
func (c *Controller) enqueueBookStore(obj interface{}) {
	log.Println("Enqueuing book store")

	// key is a string that uniquely identifies obj
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
	}
	c.WorkQueue.AddRateLimited(key)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.WorkQueue.ShutDown()
	log.Println("Starting BookStore controller...")
	log.Println("Waiting for informer caches to sync...")
	if ok := cache.WaitForCacheSync(stopCh, c.DeploymentsSynced, c.BookStoreSynced); !ok {
		return fmt.Errorf("failed to wait for cache to sync")
	}
	log.Println("Starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	log.Println("Worker started successfully")
	<-stopCh
	log.Println("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.ProcessNextItem() {

	}
}
func (c *Controller) ProcessNextItem() bool {
	obj, shutdown := c.WorkQueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer c.WorkQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.WorkQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.WorkQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.WorkQueue.Forget(obj)
		log.Printf("Successfully synced '%s'", key)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	bookStore, err := c.BookStoreLister.BookStores(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("BookStore '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	deploymentName := bookStore.Spec.Name
	if deploymentName == "" {
		utilruntime.HandleError(fmt.Errorf("BookStore '%s' in work queue no longer exists", key))
	}

}
