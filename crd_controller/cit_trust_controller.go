package crd_controller

import (
         "fmt"
         "time"
         "strings"
	"cit_custom_controller/crd_label_annotate"
        "github.com/golang/glog"
	trust_schema "cit_custom_controller/crd_schema/cit_trust_schema"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
        "k8s.io/apimachinery/pkg/util/wait"
        "k8s.io/apimachinery/pkg/util/runtime"
        "k8s.io/client-go/util/workqueue"
        "k8s.io/client-go/tools/cache"
)

type citTLController struct {
        indexer  cache.Indexer
        informer cache.Controller
        queue workqueue.RateLimitingInterface
}

func NewCitTLController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *citTLController {
        return &citTLController{
                informer: informer,
                indexer:  indexer,
                queue:    queue,
                //tlObj:    make(map[string]trust_schema.Trusttabspec),
        }
}

func GetTLCrdDef() CrdDefinition {
        return CrdDefinition{
                Plural:   trust_schema.CITTLPlural,
                Singular: trust_schema.CITTLSingular,
                Group:    trust_schema.CITTLGroup,
                Kind:     trust_schema.CITTLKind,
        }
}

func (c *citTLController) processNextItem() bool {
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
	if err == nil {
                c.queue.Forget(key)
                return true
        }
        // Handle the error if something went wrong during the execution of the business logic
        c.handleErr(err, key)
        return true
}

//processTLQueue : can be extended to validate the crd objects are been acted upon 
func (c *citTLController) processTLQueue(key string) error {
	glog.Infof("processTLQueue for Key %#v ", key)
	return nil
}

// syncFromQueue is the business logic of the controller. In this controller it simply prints
// information about the CRD to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *citTLController) syncFromQueue(key string) error {
        obj, exists, err := c.indexer.GetByKey(key)
        if err != nil {
                glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
                return err
        }

        if !exists {
                // Below we will warm up our cache with a CDR, so that we will see a delete for one CRD
		glog.Infof("TL CRD object %s does not exist anymore\n", key)
        } else {
                // Note that you also have to check the uid if you have a local controlled resource, which
                // is dependent on the actual instance, to detect that a CRD object was recreated with the same name
		glog.Infof("Sync/Add/Update for TL CRD Object %#v ", obj)
		c.processTLQueue(key)
        }
        return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *citTLController) handleErr(err error, key interface{}) {
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

func (c *citTLController) Run(threadiness int, stopCh chan struct{}) {
        defer runtime.HandleCrash()

        // Let the workers stop when we are done
        defer c.queue.ShutDown()
        glog.Info("Starting Trust Tab controller")

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
        glog.Info("Stopping Trust Tab controller")
}

func (c *citTLController) runWorker() {
        for c.processNextItem() {
        }
}

//GetTlObjLabel creates lables and annotations map based on TL CRD
func GetTlObjLabel(obj trust_schema.HostList) (crd_label_annotate.Labels, crd_label_annotate.Annotations) {
        var lbl = make(crd_label_annotate.Labels, 2)
        var annotation = make(crd_label_annotate.Annotations, 1)
        expiry := strings.Replace(obj.TrustTagExpiry, ":", ".", -1)
        lbl[trustexpiry] = expiry
        lbl[trustlabel] = obj.Trusted
        annotation[trustsignreport] = obj.TrustTagSignedReport

        return lbl, annotation
}

//AddTrustTabObj Handler for addition event of the TL CRD
func AddTrustTabObj(trustobj *trust_schema.Trustcrd, helper crd_label_annotate.APIHelpers, cli *k8sclient.Clientset) {
	//fmt.Println("cast event name ", trustobj.Name)
	for index, ele := range trustobj.Spec.HostList {
		nodeName := trustobj.Spec.HostList[index].Hostname
		node, err := helper.GetNode(cli, nodeName)
		if err != nil {
        		glog.Info("Failed to get node within cluster: %s", err.Error())
			return
		}
		lbl, ann := GetTlObjLabel(ele)
		helper.AddLabelsAnnotations(node, lbl, ann)
		err = helper.UpdateNode(cli, node)
		if err != nil {
        		glog.Info("can't update node: %s", err.Error())
			return
		}
	}
}

//NewTLIndexerInformer returns informer for TL CRD object
func NewTLIndexerInformer(config *rest.Config, queue workqueue.RateLimitingInterface) ( cache.Indexer, cache.Controller ) {
	// Create a new clientset which include our CRD schema
        crdcs, scheme, err := trust_schema.NewTLClient(config)
        if err != nil {
                panic(err)
        }

        // Create a CRD client interface
        tlcrdclient := trust_schema.CitTLClient(crdcs, scheme, "default")

	//Create a TL CRD Helper object
	h_inf, cli := crd_label_annotate.Getk8sClientHelper(config)

	return cache.NewIndexerInformer(tlcrdclient.NewTLListWatch(), &trust_schema.Trustcrd{}, 0, cache.ResourceEventHandlerFuncs{
                AddFunc: func(obj interface{}) {
                        key, err := cache.MetaNamespaceKeyFunc(obj)
        		glog.Info("Received Add event for ", key)
                        trustobj := obj.(*trust_schema.Trustcrd)
                        if err == nil {
                                queue.Add(key)
                        }
                	AddTrustTabObj(trustobj, h_inf, cli) 
                },
                UpdateFunc: func(old interface{}, new interface{}) {
                        key, err := cache.MetaNamespaceKeyFunc(new)
        		glog.Info("Received Update event for ", key)
                        trustobj := new.(*trust_schema.Trustcrd)
                        if err == nil {
                                queue.Add(key)
                        }
                	AddTrustTabObj(trustobj, h_inf, cli) 
                },
                DeleteFunc: func(obj interface{}) {
                        // IndexerInformer uses a delta queue, therefore for deletes we have to use this
                        // key function.
                        key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
        		glog.Info("Received delete event for ", key)
                        if err == nil {
                                queue.Add(key)
                        }
                },
        }, cache.Indexers{})
}
