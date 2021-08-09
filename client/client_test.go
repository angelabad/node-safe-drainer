package client

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func fakeClient() Client {
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

	client := Client{
		Clientset: fake,
	}

	return client
}

func TestCordonNode(t *testing.T) {
	client := fakeClient()
	if err := client.cordonNode("node1"); err != nil {
		panic(err.Error())
	}

}

func TestCheckNodeName(t *testing.T) {
	client := fakeClient()
	if err := client.checkNodeName("node2"); err != nil {
		panic(err.Error())
	}
}

func TestUpdateDeployments(t *testing.T) {
	client := fakeClient()
	if err := client.updateDeployments("node1"); err != nil {
		panic(err.Error())
	}
}

func TestCordonAndEmpty(t *testing.T) {
	client := fakeClient()
	if err := client.CordonAndEmpty("node1"); err != nil {
		panic(err.Error())
	}
}
