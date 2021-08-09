package main

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func cordonNode(clientset *kubernetes.Clientset, nodeName string) error {
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/unschedulable",
		Value: true,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, err := clientset.CoreV1().Nodes().Patch(context.TODO(), nodeName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})

	return err
}
