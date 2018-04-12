/*
Copyright © 2018 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/
package main

import (
	"k8s_custom_cit_controllers-k8s_custom_controllers/crdController"
	"flag"
	"github.com/golang/glog"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"os"
	"sync"
)

const MAXFILESIZE int64 = (256 * 1024 * 1024)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {

	glog.V(4).Infof("Starting Cit Custom Controller")

	kubeConf := flag.String("kubeconf", "", "Path to a kube config. Only required if out-of-cluster.")
	trustedPrefixConf := flag.String("trustedprefixconf", "", "Path to a Trusted Prefix config. Only required if out-of-cluster.")
	flag.Parse()


	config, err := GetClientConfig(*kubeConf)
	if err != nil {
		panic(err.Error())
	}

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//Create mutex to sync operation between the two CRD threads
	var crdmutex = &sync.Mutex{}

	plCrdDef := crdController.GetPLCrdDef()

	//crdController.NewCitCustomResourceDefinition to create PL CRD
	err = crdController.NewCitCustomResourceDefinition(cs, &plCrdDef)
	if err != nil {
		panic(err)
	}

	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "citPLcontroller")

	plindexer, plinformer := crdController.NewPLIndexerInformer(config, queue, crdmutex,*trustedPrefixConf)

	controller := crdController.NewCitPLController(queue, plindexer, plinformer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	glCrdDef := crdController.GetGLCrdDef()

	// note: if the CRD exist our CreateCRD function is set to exit without an error
	err = crdController.NewCitCustomResourceDefinition(cs, &glCrdDef)
	if err != nil {
		panic(err)
	}

	// Create a queue
	glQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "citGLcontroller")

	glindexer, glinformer := crdController.NewGLIndexerInformer(config, glQueue, crdmutex)

	geolocationController := crdController.NewCitGLController(glQueue, glindexer, glinformer)
	stopGl := make(chan struct{})
	defer close(stopGl)
	go geolocationController.Run(1, stopGl)

	glog.V(4).Infof("Waiting for updates on  Cit Custom Resource Definitions")

	// Wait forever
	select {}
}
