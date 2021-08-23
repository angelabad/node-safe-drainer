package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"angelabad.me/node-safe-drain/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	poll             = 2 * time.Second
	pollShortTimeout = 1 * time.Minute
	maxInQueue       = 10
)

type Client struct {
	Clientset kubernetes.Interface
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}

func (c Client) Rollout(d Deployment) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := c.Clientset.AppsV1().Deployments(d.Namespace).Get(context.TODO(), d.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		annotations := result.Spec.Template.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["app.kubernetes.io/safeDrainRestarted"] = time.Now().Format(time.RFC3339)
		result.Spec.Template.Annotations = annotations
		fmt.Printf("Starting rolling out deployment: %s/%s\n", d.Namespace, d.Name)
		updatedDeployment, updateErr := c.Clientset.AppsV1().Deployments(d.Namespace).Update(context.TODO(), result, metav1.UpdateOptions{})

		err = c.waitForDeploymentComplete(updatedDeployment)
		fmt.Printf("Finished: %s/%s\n", d.Namespace, d.Name)

		return updateErr
	})

	return err
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

func (c Client) getPodDeploymentOwner(pod corev1.Pod) (Deployment, error) {
	var deploy Deployment
	namespace := pod.Namespace

	replicaOwner := metav1.GetControllerOf(&pod)
	if replicaOwner.Kind == "ReplicaSet" {
		replica, err := c.Clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaOwner.Name, metav1.GetOptions{})
		if err != nil {
			return Deployment{}, err
		}
		deploymentOwner := metav1.GetControllerOf(replica)
		if deploymentOwner.Kind == "Deployment" {
			deployment, err := c.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentOwner.Name, metav1.GetOptions{})
			if err != nil {
				return Deployment{}, err
			}
			deploy.Name = deployment.Name
			deploy.Namespace = deployment.Namespace
		}
	}
	return deploy, nil
}

func (c Client) getNodeDeployments(node string) (Deployments, error) {
	var deployments Deployments

	pods, err := c.Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		deploy, err := c.getPodDeploymentOwner(pod)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, deploy)
	}

	deployments.deduplicate()
	return deployments, nil
}

func (c Client) updateDeployments(node string) error {
	deployments, err := c.getNodeDeployments(node)
	if err != nil {
		return err
	}

	q := utils.NewQueue(maxInQueue)
	defer q.Close()

	for _, deployment := range deployments {
		// Wait for kubernetes server too much requests error
		time.Sleep(2 * time.Second)
		q.Add()

		go func(d Deployment) error {
			defer q.Done()
			if err := c.Rollout(d); err != nil {
				return err
			}
			return nil
		}(deployment)
	}
	q.Wait()

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
