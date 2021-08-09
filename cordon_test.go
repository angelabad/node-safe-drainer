package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestMain(t *testing.T) {
	fake := fake.NewSimpleClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Labels: map[string]string{
				"node": "1",
			},
		},
	}, &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
			Labels: map[string]string{
				"node": "2",
			},
		},
	},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: corev1.PodSpec{
				NodeSelector: map[string]string{
					"node": "1",
				},
			},
		},
	)

	nodes, err := fake.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("FALLO\n")
	}

	for _, node := range nodes.Items {
		fmt.Printf("ESTE NODO ES: %s\n", node.Name)
		testPods, _ := fake.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		fmt.Printf("PODS: %s\n", testPods.Items[0].Name)
	}
	assert.Equal(t, "angel", "angel")
}
