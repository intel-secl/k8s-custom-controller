/*
Copyright © 2018 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/
package crdController

import (
	geo_schema "cit_custom_controller/crdSchema/citGeolocationSchema"
	"testing"
)

func TestGetGLCrdDef(t *testing.T) {
	expecGlCrd := CrdDefinition{
		Plural:   "geolocationcrds",
		Singular: "geolocationcrd",
		Group:    "cit.intel.com",
		Kind:     "GeolocationCrd",
	}
	recvGlCrd := GetGLCrdDef()
	if expecGlCrd != recvGlCrd {
		t.Fatalf("Changes found in GL CRD Definition ")
		t.Fatalf("Expected :%v however Received: %v ", expecGlCrd, recvGlCrd)
	}
	t.Logf("Test GetGLCrd Def success")
}

func TestGetGlObjLabel(t *testing.T) {
	geoObj := geo_schema.HostList{
		Hostname:             "Node123",
		AssetTagExpiry:       "12-23-45T123.91.12",
		AssetTagSignedReport: "295270d6242e2c67e24e22bad49dtera",
		Assettag: map[string]string{
			"country.us":  "true",
			"country.uk":  "true",
			"state.ca":    "true",
			"city.seatle": "true",
		},
	}
	recvlabel, recannotate := GetGlObjLabel(geoObj)
	if _, ok := recvlabel["country.us"]; ok {
		t.Logf("Found GL label in AssetTag report")
	} else {
		t.Fatalf("Could not get required label from GL Report")
	}
	if _, ok := recvlabel["AssetTagExpiry"]; ok {
		t.Logf("Found in GL label AssetTagExpiry field")
	} else {
		t.Fatalf("Could not get label AssetTagExpiry from GL Report")
	}
	if _, ok := recannotate["AssetTagSignedReport"]; ok {
		t.Logf("Found in GL annotation : AssetTagSignedReport ")
	} else {
		t.Fatalf("Could not get annotation AssetTagSignedReport from TL Report")
	}
	t.Logf("Test GetGlObjLabel success")
}
