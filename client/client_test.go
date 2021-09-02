package client

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

func fakeClient() Client {
	fake := fake.NewSimpleClientset(
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy1",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         "apps/v1",
						Kind:               "Deployment",
						Name:               "deploy1",
						Controller:         boolPtr(true),
						BlockOwnerDeletion: boolPtr(true),
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy1-1",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         "apps/v1",
						Kind:               "ReplicaSet",
						Name:               "deploy1",
						Controller:         boolPtr(true),
						BlockOwnerDeletion: boolPtr(true),
					},
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "web",
						Image: "nginx:1.12",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								Protocol:      corev1.ProtocolTCP,
								ContainerPort: 80,
							},
						},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy1",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
			},
		},
	)

	client := Client{
		Clientset: fake,
	}

	return client
}

func TestRollout(t *testing.T) {
	client := fakeClient()

	d := Deployment{
		Namespace: "default",
		Name:      "deploy1",
	}
	if err := client.Rollout(d); err != nil {
		panic(err.Error())
	}
}

func TestCordonNode(t *testing.T) {
	client := fakeClient()
	if err := client.cordonNodes([]string{"node1"}); err != nil {
		panic(err.Error())
	}

}

func TestCheckNodeName(t *testing.T) {
	client := fakeClient()
	if err := client.checkNodeName("node1"); err != nil {
		panic(err.Error())
	}
}

func TestUpdateDeployments(t *testing.T) {
	client := fakeClient()
	if err := client.updateDeployments([]string{"node1"}, 10); err != nil {
		panic(err.Error())
	}
}

func TestCordonAndEmpty(t *testing.T) {
	client := fakeClient()
	if err := client.CordonAndEmpty([]string{"node1"}, 10); err != nil {
		panic(err.Error())
	}
}
