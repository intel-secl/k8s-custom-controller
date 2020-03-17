/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"flag"
	"github.com/intel-secl/k8s-custom-controller/crdController"
	"sync"

	"github.com/golang/glog"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

const TrustedPrefixConf = "/opt/isecl-k8s-extensions/tag_prefix.conf"

func main() {

	glog.V(4).Infof("Starting ISecL Custom Controller")

	kubeConf := flag.String("kubeconf", "", "Path to a kube config. ")
	skipCrdCreate := flag.Bool("skip-crd-create", false, "skip crd creation in code")
	flag.Parse()

	config, err := GetClientConfig(*kubeConf)
	if err != nil {
		glog.Errorf("Error in config %v", err)
		return
	}

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		glog.Errorf("Error in config %v", err)
		return
	}

	//Create mutex to sync operation between the two CRD threads
	var crdmutex = &sync.Mutex{}

        if !*skipCrdCreate {
                CrdDef := crdController.GetHACrdDef()

                //crdController.NewIseclCustomResourceDefinition to create CRD
                err = crdController.NewIseclCustomResourceDefinition(cs, &CrdDef)
                if err != nil {
                        glog.Errorf("Error in creating platform CRD %v", err)
                        return
                }
        }

	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "iseclcontroller")

	indexer, informer := crdController.NewIseclHAIndexerInformer(config, queue, crdmutex, TrustedPrefixConf)

	controller := crdController.NewIseclHAController(queue, indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	glog.V(4).Infof("Waiting for updates on  ISecl Custom Resource Definitions")

	// Wait forever
	select {}
}
