/*
Copyright © 2018 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/
package crdController

import (
        "k8s_custom_cit_controllers-k8s_custom_controllers/crdLabelAnnotate"
        trust_schema "k8s_custom_cit_controllers-k8s_custom_controllers/crdSchema/citTrustSchema"
        "encoding/json"
        "fmt"
        "github.com/golang/glog"
        "io/ioutil"
        "k8s.io/apimachinery/pkg/util/runtime"
        "k8s.io/apimachinery/pkg/util/wait"
        k8sclient "k8s.io/client-go/kubernetes"
        api "k8s.io/client-go/pkg/api/v1"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/tools/cache"
        "k8s.io/client-go/util/workqueue"
        "strings"
        "sync"
        "time"
)

type citPLController struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface
}

type Config struct {
        Trusted string `"json":"trustedPrefix"`
}


func NewCitPLController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *citPLController {
	return &citPLController{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func GetPLCrdDef() CrdDefinition {
	return CrdDefinition{
		Plural:   trust_schema.CITPLPlural,
		Singular: trust_schema.CITPLSingular,
		Group:    trust_schema.CITPLGroup,
		Kind:     trust_schema.CITPLKind,
	}
}

func (c *citPLController) processNextItem() bool {
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

//processPLQueue : can be extended to validate the crd objects are been acted upon
func (c *citPLController) processPLQueue(key string) error {
	glog.Infof("processPLQueue for Key %#v ", key)
	return nil
}

// syncFromQueue is the business logic of the controller. In this controller it simply prints
// information about the CRD to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *citPLController) syncFromQueue(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a CDR, so that we will see a delete for one CRD
		glog.Infof("PL CRD object %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a CRD object was recreated with the same name
		glog.Infof("Sync/Add/Update for PL CRD Object %#v ", obj)
		c.processPLQueue(key)
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *citPLController) handleErr(err error, key interface{}) {
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

func (c *citPLController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	glog.Info("Starting Platformcrd controller")

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
	glog.Info("Stopping Platform controller")
}

func (c *citPLController) runWorker() {
	for c.processNextItem() {
	}
}

//GetPlObjLabel creates lables and annotations map based on PL CRD
func GetPlObjLabel(obj trust_schema.HostList, node *api.Node, trustedPrefixConf string) (crdLabelAnnotate.Labels, crdLabelAnnotate.Annotations,error) {
	var lbl = make(crdLabelAnnotate.Labels, 2)
	var annotation = make(crdLabelAnnotate.Annotations, 1)
	trustPresent := false
	trustLabelWithPrefix,err := getPrefixFromConf(trustedPrefixConf) 
	if err != nil {
		return nil,nil,err
	}
	trustLabelWithPrefix = trustLabelWithPrefix + trustlabel

	//Comparing with existing node labels
	for key, value := range node.Labels {
		if key == trustLabelWithPrefix {
			trustPresent = true
			if value == obj.Trusted {
				glog.Info("No change in Trustlabel, updating Trustexpiry time only")
			} else {
				glog.Info("Updating Complete Trustlabel for the node")
				lbl[trustLabelWithPrefix] = obj.Trusted
			}
		}
	}
	if !trustPresent {
		glog.Info("Trust value was not present on node adding for first time")
		lbl[trustLabelWithPrefix] = obj.Trusted
	}
	expiry := strings.Replace(obj.TrustTagExpiry, ":", ".", -1)
	lbl[trustexpiry] = expiry
	annotation[trustsignreport] = obj.TrustTagSignedReport

	return lbl, annotation,nil
}

func getPrefixFromConf(path string) (string, error) {
	out, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("Error: %s %v", path, err)
		return "",err
	}
	s := Config{}
	err = json.Unmarshal(out, &s)
	if err != nil {
		glog.Errorf("Error:  %v", err)
		return "",err
	}
	return s.Trusted, nil
}

//AddTrustTabObj Handler for addition event of the TL CRD
func AddTrustTabObj(trustobj *trust_schema.Platformcrd, helper crdLabelAnnotate.APIHelpers, cli *k8sclient.Clientset, mutex *sync.Mutex, trustedPrefixConf string) {
	for index, ele := range trustobj.Spec.HostList {
		nodeName := trustobj.Spec.HostList[index].Hostname
		node, err := helper.GetNode(cli, nodeName)
		if err != nil {
			glog.Info("Failed to get node within cluster: %s", err.Error())
			return
		}
		lbl, ann ,err := GetPlObjLabel(ele, node, trustedPrefixConf)
		if err != nil {
			glog.Fatalf("Error: %v", err)
		}
		mutex.Lock()
		helper.AddLabelsAnnotations(node, lbl, ann)
		err = helper.UpdateNode(cli, node)
		mutex.Unlock()
		if err != nil {
			glog.Info("can't update node: %s", err.Error())
			return
		}
	}
}

//NewPLIndexerInformer returns informer for PL CRD object
func NewPLIndexerInformer(config *rest.Config, queue workqueue.RateLimitingInterface, crdMutex *sync.Mutex, trustedPrefixConf string) (cache.Indexer, cache.Controller) {
	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := trust_schema.NewPLClient(config)
	if err != nil {
		glog.Fatalf("Failed to create new clientset for Platform CRD", err)
	}

	// Create a CRD client interface
	plcrdclient := trust_schema.CitPLClient(crdcs, scheme, "default")

	//Create a PL CRD Helper object
	hInf, cli := crdLabelAnnotate.Getk8sClientHelper(config)

	return cache.NewIndexerInformer(plcrdclient.NewPLListWatch(), &trust_schema.Platformcrd{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			glog.Info("Received Add event for ", key)
			trustobj := obj.(*trust_schema.Platformcrd)
			if err == nil {
				queue.Add(key)
			}
			AddTrustTabObj(trustobj, hInf, cli, crdMutex, trustedPrefixConf)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			glog.Info("Received Update event for ", key)
			trustobj := new.(*trust_schema.Platformcrd)
			if err == nil {
				queue.Add(key)
			}
			AddTrustTabObj(trustobj, hInf, cli, crdMutex, trustedPrefixConf)
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
