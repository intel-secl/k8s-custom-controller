package crdController

import (
	"cit_custom_controller/crdLabelAnnotate"
	geolocation_schema "cit_custom_controller/crdSchema/citGeolocationSchema"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strings"
	"sync"
	"time"
)

type citGLController struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface
}

func NewCitGLController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *citGLController {
	return &citGLController{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func GetGLCrdDef() CrdDefinition {
	return CrdDefinition{
		Plural:   geolocation_schema.CITGLPlural,
		Singular: geolocation_schema.CITGLSingular,
		Group:    geolocation_schema.CITGLGroup,
		Kind:     geolocation_schema.CITGLKind,
	}
}

func (c *citGLController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two CRD with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncFromQueue(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the CRD to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *citGLController) syncFromQueue(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a CDR, so that we will see a delete for one CRD
		glog.Infof("CRD object %#v does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a CRD object was recreated with the same name
		glog.Infof("Sync/Add/Update for CRD %#v", obj)
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *citGLController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		glog.Infof("Error syncing CRD %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping CRD %q out of the queue: %v", key, err)
}

func (c *citGLController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	glog.Info("Starting Geo Location controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping Geo Location controller")
}

func (c *citGLController) runWorker() {
	for c.processNextItem() {
	}
}

//AddGeoTabObj Handler for addition event of the GL CRD
func AddGeoTabObj(geoobj *geolocation_schema.Geolocationcrd, helper crdLabelAnnotate.APIHelpers, cli *k8sclient.Clientset, mutex *sync.Mutex) {
	for index, ele := range geoobj.Spec.HostList {
		nodeName := geoobj.Spec.HostList[index].Hostname
		node, err := helper.GetNode(cli, nodeName)
		if err != nil {
			glog.Infof("failed to get node: %s", err.Error())
			return
		}
		lbl, ann := GetGlObjLabel(ele)
		mutex.Lock()
		helper.AddLabelsAnnotations(node, lbl, ann)
		err = helper.UpdateNode(cli, node)
		mutex.Unlock()
		if err != nil {
			glog.Infof("can't update node: %s", err.Error())
			//return
		}
	}
}

//getGlAssettag creates lables map based on asset tag field of GL CRD
func getGlAssettag(obj geolocation_schema.HostList) crdLabelAnnotate.Labels {
	size := len(obj.Assettag)
	//fmt.Printf("Number of keys in AssetTag: %d \n", size)
	var lbl = make(crdLabelAnnotate.Labels, size+1)
	for key, val := range obj.Assettag {
		labelkey := strings.Replace(key, " ", ".", -1)
		labelkey = strings.Replace(labelkey, ":", ".", -1)
		lbl[labelkey] = val
	}
	return lbl
}

//GetGlObjLabel creates lables and annotations map based on GL CRD
func GetGlObjLabel(obj geolocation_schema.HostList) (crdLabelAnnotate.Labels, crdLabelAnnotate.Annotations) {
	var annotation = make(crdLabelAnnotate.Annotations, 1)
	lbl := getGlAssettag(obj)
	assetexpiryval := strings.Replace(obj.AssetTagExpiry, ":", ".", -1)
	lbl[assetexpiry] = assetexpiryval
	annotation[assetsignreport] = obj.AssetTagSignedReport
	return lbl, annotation
}

func NewGLIndexerInformer(config *rest.Config, queue workqueue.RateLimitingInterface, crdMutex *sync.Mutex) (cache.Indexer, cache.Controller) {
	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := geolocation_schema.NewGLClient(config)
	if err != nil {
		panic(err)
	}

	// Create a CRD client interface
	glcrdclient := geolocation_schema.CitGLClient(crdcs, scheme, "default")

	//Create a GL CRD Helper object
	hInf, cli := crdLabelAnnotate.Getk8sClientHelper(config)

	return cache.NewIndexerInformer(glcrdclient.NewGLListWatch(), &geolocation_schema.Geolocationcrd{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			geoObject := obj.(*geolocation_schema.Geolocationcrd)
			key, err := cache.MetaNamespaceKeyFunc(geoObject)
			glog.Infof("Received Add event for %#v", key)
			if err == nil {
				queue.Add(key)
			}
			AddGeoTabObj(geoObject, hInf, cli, crdMutex)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			geoObject := new.(*geolocation_schema.Geolocationcrd)
			key, err := cache.MetaNamespaceKeyFunc(geoObject)
			glog.Infof("Received Update event for %#v", key)
			if err == nil {
				queue.Add(key)
			}
			AddGeoTabObj(geoObject, hInf, cli, crdMutex)
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			glog.Infof("Received delete event f0r %#v", key)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})
}
