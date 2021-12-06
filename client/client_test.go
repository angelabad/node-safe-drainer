/*
 * Copyright (c) 2021 Angel Abad. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
		MaxJobs:   10,
	}

	return client
}

/*
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
*/

func TestCordonNode(t *testing.T) {
	client := fakeClient()
	if err := client.cordonNodes([]string{"node1"}); err != nil {
		panic(err.Error())
	}

}

func TestCheckNodeName(t *testing.T) {
	client := fakeClient()

	err := client.checkNodeName("node1")
	assert.Nil(t, err)
}

/*
func TestUpdateDeployments(t *testing.T) {
	client := fakeClient()
	if err := client.updateDeployments([]string{"node1"}); err != nil {
		panic(err.Error())
	}
}
*/

/*
func TestCordonAndEmpty(t *testing.T) {
	client := fakeClient()
	if err := client.CordonAndEmpty([]string{"node1"}); err != nil {
		panic(err.Error())
	}
}
*/

func TestGetAllNodes(t *testing.T) {
	expected := []string{"node1"}

	client := fakeClient()
	nodeList, err := client.GetAllNodes()

	assert.Nil(t, err)
	assert.Equal(t, expected, nodeList)
}

func TestGetPodDeploymentOwner(t *testing.T) {
	expected := &Deployment{
		Namespace: "default",
		Name:      "deploy1",
	}

	client := fakeClient()
	pod, err := client.Clientset.CoreV1().Pods("default").Get(context.TODO(), "deploy1-1", metav1.GetOptions{})

	assert.Nil(t, err)
	assert.Equal(t, expected, client.getPodDeploymentOwner(*pod))
}

func TestGetNodeDeployments(t *testing.T) {
	expected := Deployments{
		Deployment{
			Namespace: "default",
			Name:      "deploy1",
		},
	}

	client := fakeClient()
	deploys, err := client.getNodeDeployments([]string{"node1"})

	assert.Nil(t, err)
	assert.Equal(t, expected, deploys)
}
