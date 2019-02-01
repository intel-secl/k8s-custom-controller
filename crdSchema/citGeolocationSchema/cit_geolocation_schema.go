/*
Copyright © 2018 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/
package crdGeolocationSchema

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	CITGLPlural   string = "geolocationcrds"
	CITGLSingular string = "geolocationcrd"
	CITGLGroup    string = "isecl.intel.com"
	CITGLKind     string = "GeolocationCrd"
	CITGLVersion  string = "v1beta1"
)

//CitGLClient returns rest client for the GL CRD
func CitGLClient(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *citglclient {
	return &citglclient{cl: cl, ns: namespace, plural: CITGLPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type citglclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

// Definition of our CRD Example class
type Geolocationcrd struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               geolocationspec `json:"spec"`
}

type HostList struct {
	Hostname             string            `json:"hostName"`
	AssetTagExpiry       string            `json:"validTo"`
	Assettag             map[string]string `json:"assetTags"`
	AssetTagSignedReport string            `json:"signedTrustReport"`
}

type geolocationspec struct {
	HostList []HostList `json:"hostList"`
}

type geolocationtabstatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type GeolocationCrdList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Geolocationcrd `json:"items"`
}

// Create a  Rest client with the new CRD Schema
var SchemeGroupVersion = schema.GroupVersion{Group: CITGLGroup, Version: CITGLVersion}

//addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Geolocationcrd{},
		&GeolocationCrdList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

//NewTLClient registers CRD schema and returns rest client for the CRD
func NewGLClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}
	return client, scheme, nil
}

// Create a new List watch for our GL CRD
func (f *citglclient) NewGLListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}
