/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdLabelAnnotate

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type APIHelpers interface {

	// GetNode returns the Kubernetes node on which this container is running.
	GetNode(*k8sclient.Clientset, string) (*corev1.Node, error)

	// AddLabelsAnnotations modifies the supplied node's labels and annotations collection.
	// In order to publish the labels, the node must be subsequently updated via the
	// API server using the client library.
	AddLabelsAnnotations(*corev1.Node, Labels, Annotations)

	// UpdateNode updates the node via the API server using a client.
	UpdateNode(*k8sclient.Clientset, *corev1.Node) error

	// AddTaint modifies the supplied node's taints to add an additional taint
	// effect should be one of: NoSchedule, PreferNoSchedule, NoExecute
	AddTaint(n *corev1.Node, key string, value string, effect string) error
}

// Implements main.APIHelpers
type K8sHelpers struct{}
type Labels map[string]string
type Annotations map[string]string

//Getk8sClientHelper returns helper object and clientset to fetch node
func Getk8sClientHelper(config *rest.Config) (APIHelpers, *k8sclient.Clientset) {
	helper := APIHelpers(K8sHelpers{})

	cli, err := k8sclient.NewForConfig(config)
	if err != nil {
		glog.Errorf("Error while creating k8s client %v", err)
	}
	return helper, cli
}

//GetNode returns node API based on nodename
func (h K8sHelpers) GetNode(cli *k8sclient.Clientset, NodeName string) (*corev1.Node, error) {
	// Get the node object using the node name
	node, err := cli.CoreV1().Nodes().Get(NodeName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Can't get node: %s", err.Error())
		return nil, err
	}

	return node, nil
}

//AddLabelsAnnotations applys labels and annotations to the node
func (h K8sHelpers) AddLabelsAnnotations(n *corev1.Node, labels Labels, annotations Annotations) {
	for k, v := range labels {
		n.Labels[k] = v
	}
	for k, v := range annotations {
		n.Annotations[k] = v
	}
}

//AddTaint applys labels and annotations to the node
//effect should be one of: NoSchedule, PreferNoSchedule, NoExecute
func (h K8sHelpers) AddTaint(n *corev1.Node, key string, value string, effect string) error {
	taintEffect, ok := map[string]corev1.TaintEffect{
		"NoSchedule":       corev1.TaintEffectNoSchedule,
		"PreferNoSchedule": corev1.TaintEffectPreferNoSchedule,
		"NoExecute":        corev1.TaintEffectNoExecute,
	}[effect]

	if !ok {
		return errors.Errorf("Taint effect %v not valid", effect)
	}

	n.Spec.Taints = append(n.Spec.Taints, corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: taintEffect,
	})

	return nil
}

//UpdateNode updates the node API
func (h K8sHelpers) UpdateNode(c *k8sclient.Clientset, n *corev1.Node) error {
	// Send the updated node to the apiserver.
	_, err := c.CoreV1().Nodes().Update(n)
	if err != nil {
		glog.Errorf("Error while updating node label:", err.Error())
		return err
	}
	return nil
}
