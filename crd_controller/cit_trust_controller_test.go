package crd_controller

import (
	//apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	//"k8s.io/client-go/kubernetes/fake"
	"apiextensions-apiserver/test/integration/testserver"
	//"meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	 trust_schema "citcrd/crd_schema/cit_trust_schema"
	 "testing"
	)

func TestTLCRDCreation(t *testing.T) {
	masterConfig, err := rest.InClusterConfig()
	fakeClient := &fake.Clientset{}
	err := NewcitTLCustomResourceDefinition(fakeClient)
        if err != nil {
                t.Fatalf("error creating Trust Label CRD: %v", err)
        }

        t.Logf("Testing cit TL controller success")
}
