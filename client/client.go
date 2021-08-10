package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	poll             = 2 * time.Second
	pollShortTimeout = 1 * time.Minute
)

type Client struct {
	Clientset kubernetes.Interface
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}

func (c Client) CordonAndEmpty(name string) error {
	if err := c.cordonNode(name); err != nil {
		return err
	}

	if err := c.updateDeployments(name); err != nil {
		return err
	}

	return nil
}

func (c Client) cordonNode(name string) error {
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/unschedulable",
		Value: true,
	}}

	err := c.checkNodeName(name)
	if err != nil {
		return err
	}

	payloadBytes, _ := json.Marshal(payload)
	_, err = c.Clientset.CoreV1().Nodes().Patch(context.TODO(), name, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})

	return err
}

func (c Client) checkNodeName(name string) error {
	_, err := c.Clientset.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c Client) updateDeployments(node string) error {
	pods, err := c.Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		if pod.OwnerReferences[0].Kind == "ReplicaSet" {
			replica, err := c.Clientset.AppsV1().ReplicaSets(pod.Namespace).Get(context.TODO(), pod.OwnerReferences[0].Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Namespace: %s, Deployment: %s\n", pod.Namespace, replica.OwnerReferences[0].Name)

			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				result, getErr := c.Clientset.AppsV1().Deployments(replica.Namespace).Get(context.TODO(), replica.OwnerReferences[0].Name, metav1.GetOptions{})
				if getErr != nil {
					panic(getErr.Error())
				}
				annotations := result.Spec.Template.GetAnnotations()
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations["roller.angelabad.me/restartedAt"] = time.Now().String()
				result.Spec.Template.Annotations = annotations
				deployment, updateErr := c.Clientset.AppsV1().Deployments(replica.Namespace).Update(context.TODO(), result, metav1.UpdateOptions{})

				err = c.waitForDeploymentComplete(deployment)

				return updateErr
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c Client) waitForDeploymentComplete(d *appsv1.Deployment) error {
	var reason string

	err := wait.PollImmediate(poll, pollShortTimeout, func() (bool, error) {
		deployment, err := c.Clientset.AppsV1().Deployments(d.Namespace).Get(context.TODO(), d.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if deploymentComplete(d, &deployment.Status) {
			return true, nil
		}

		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)

		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("error waiting timeout: %s", reason)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", d.Name, err)
	}

	return nil
}

func deploymentComplete(deployment *appsv1.Deployment, newStatus *appsv1.DeploymentStatus) bool {
	return newStatus.UpdatedReplicas == *(deployment.Spec.Replicas) &&
		newStatus.Replicas == *(deployment.Spec.Replicas) &&
		newStatus.AvailableReplicas == *(deployment.Spec.Replicas) &&
		newStatus.ObservedGeneration >= deployment.Generation
}