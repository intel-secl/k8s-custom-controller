package crd_controller

import (
	trust_schema "cit_custom_controller/crd_schema/cit_trust_schema"
	api "k8s.io/client-go/pkg/api/v1"
	"testing"
)

func TestGetTLCrdDef(t *testing.T) {
	expecTlCrd := CrdDefinition{
		Plural:   "trustcrds",
		Singular: "trustcrd",
		Group:    "cit.intel.com",
		Kind:     "TrustCrd",
	}
	recvTlCrd := GetTLCrdDef()
	if expecTlCrd != recvTlCrd {
		t.Fatalf("Changes found in TL CRD Definition ")
		t.Fatalf("Expected :%v however Received: %v ", expecTlCrd, recvTlCrd)
	}
	t.Logf("Test GetTLCrd Def success")
}

func TestGetTlObjLabel(t *testing.T) {
	trust_obj := trust_schema.HostList{
		Hostname:             "Node123",
		Trusted:              "true",
		TrustTagExpiry:       "12-23-45T123.91.12",
		TrustTagSignedReport: "495270d6242e2c67e24e22bad49dgdah",
	}
	node := &api.Node{}
	recvlabel, recannotate := GetTlObjLabel(trust_obj, node)
	if _, ok := recvlabel["trusted"]; ok {
		t.Logf("Found in TL label Trusted field")
	} else {
		t.Fatalf("Could not get label trusted from TL Report")
	}
	if _, ok := recvlabel["TrustTagExpiry"]; ok {
		t.Logf("Found in TL label TrustTagExpiry field")
	} else {
		t.Fatalf("Could not get label TrustTagExpiry from TL Report")
	}
	if _, ok := recannotate["TrustTagSignedReport"]; ok {
		t.Logf("Found in TL annotation TrustTagSignedReport ")
	} else {
		t.Fatalf("Could not get annotation TrustTagSignedReport from TL Report")
	}
	t.Logf("Test getTlObjLabel success")
}
