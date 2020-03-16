/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package v1beta1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

const (
	HAPlural   string = "hostattributes"
	HASingular string = "hostattribute"
	HAKind     string = "HostAttributesCrd"
	HAGroup    string = "crd.isecl.intel.com"
	HAVersion  string = "v1beta1"
)

//HAClient returns CRD clientset required to apply watch on the CRD
func HAClient(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *haclient {
	return &haclient{cl: cl, ns: namespace, plural: HAPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type haclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HostAttributesCrd struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               Spec `json:"spec"`
}

type Host struct {
	Hostname     string            `json:"hostName"`
	Trusted      string            `json:"trusted"`
	Expiry       string            `json:"validTo"`
	SignedReport string            `json:"signedTrustReport"`
	Assettag     map[string]string `json:"assetTags"`
}

type Spec struct {
	HostList []Host `json:"hostList"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HostAttributesCrdList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []HostAttributesCrd `json:"items"`
}
