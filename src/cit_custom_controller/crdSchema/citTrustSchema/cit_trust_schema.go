package crd_trust_schema

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
	CITPLPlural   string = "platformcrds"
	CITPLSingular string = "platformcrd"
	CITPLKind     string = "PlatformCrd"
	CITPLGroup    string = "cit.intel.com"
	CITPLVersion  string = "v1beta1"
)

//CitPLClient returns CRD clientset required to apply watch on the CRD
func CitPLClient(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *citplclient {
	return &citplclient{cl: cl, ns: namespace, plural: CITPLPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type citplclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

type Platformcrd struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               Platformspec `json:"spec"`
}

type HostList struct {
	Hostname             string `json:"hostName"`
	Trusted              string `json:"trusted"`
	TrustTagExpiry       string `json:"validTo"`
	TrustTagSignedReport string `json:"signedTrustReport"`
}

type Platformspec struct {
	HostList []HostList `json:"hostList"`
}

type PlatformCrdList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Platformcrd `json:"items"`
}

// Create a  Rest client with the new CRD Schema
var SchemeGroupVersion = schema.GroupVersion{Group: CITPLGroup, Version: CITPLVersion}

//addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Platformcrd{},
		&PlatformCrdList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

//NewPLClient registers CRD schema and returns rest client for the CRD
func NewPLClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
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

// Create a new List watch for our PL CRD
func (f *citplclient) NewPLListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}
