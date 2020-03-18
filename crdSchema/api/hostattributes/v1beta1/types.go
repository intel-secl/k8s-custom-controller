/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package v1beta1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HAPlural   string = "hostattributeses"
	HASingular string = "hostattributes"
	HAKind     string = "HostAttributes"
	HAGroup    string = "crd.isecl.intel.com"
	HAVersion  string = "v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type HostAttributes struct {
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
type HostAttributesList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []HostAttributes `json:"items"`
}