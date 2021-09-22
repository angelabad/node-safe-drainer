package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"angelabad.me/node-safe-drainer/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	pollInterval = 5 * time.Second
)

type Client struct {
	Clientset kubernetes.Interface
	MaxJobs   int
	Timeout   time.Duration
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}

func (c *Client) Rollout(d Deployment) error {
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
		updatedDeployment, err := c.Clientset.AppsV1().Deployments(d.Namespace).Update(context.TODO(), result, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		err = c.waitForDeploymentComplete(updatedDeployment)
		if err != nil {
			return err
		}
		fmt.Printf("Finished: %s/%s\n", d.Namespace, d.Name)

		return nil
	})

	return err
}

func (c *Client) CordonAndEmpty(nodes []string) error {
	if err := c.cordonNodes(nodes); err != nil {
		return err
	}

	if err := c.updateDeployments(nodes); err != nil {
		return err
	}

	return nil

}

func (c *Client) GetAllNodes() ([]string, error) {
	nodeList := []string{}

	nodes, err := c.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		nodeList = append(nodeList, node.Name)
	}

	return nodeList, nil
}

func (c *Client) cordonNodes(nodes []string) error {
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/unschedulable",
		Value: true,
	}}

	for _, node := range nodes {
		err := c.checkNodeName(node)
		if err != nil {
			return err
		}

		payloadBytes, _ := json.Marshal(payload)
		_, err = c.Clientset.CoreV1().Nodes().Patch(context.TODO(), node, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) checkNodeName(name string) error {
	_, err := c.Clientset.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return nil
}

// TODO: Manage and return errors on this function
func (c *Client) getPodDeploymentOwner(pod corev1.Pod) *Deployment {
	namespace := pod.Namespace

	replicaOwner := metav1.GetControllerOf(&pod)
	if replicaOwner == nil {
		return nil
	}

	if replicaOwner.Kind == "ReplicaSet" {
		replica, err := c.Clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaOwner.Name, metav1.GetOptions{})
		if err != nil {
			return nil
		}
		deploymentOwner := metav1.GetControllerOf(replica)
		if deploymentOwner.Kind == "Deployment" {
			deployment, err := c.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentOwner.Name, metav1.GetOptions{})
			if err != nil {
				return nil
			}
			return &Deployment{
				Namespace: deployment.Namespace,
				Name:      deployment.Name,
			}
		}
	}

	return nil
}

func (c *Client) getNodeDeployments(nodes []string) (Deployments, error) {
	var deployments Deployments

	for _, node := range nodes {
		pods, err := c.Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node,
		})
		if err != nil {
			return Deployments{}, err
		}

		for _, pod := range pods.Items {
			deploy := c.getPodDeploymentOwner(pod)

			if deploy != nil {
				deployments = append(deployments, *deploy)
			}
		}
	}

	deployments.deduplicate()

	return deployments, nil
}

func (c *Client) updateDeployments(nodes []string) error {
	deployments, err := c.getNodeDeployments(nodes)
	if err != nil {
		return err
	}

	q := utils.NewQueue(c.MaxJobs)
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

func (c *Client) waitForDeploymentComplete(d *appsv1.Deployment) error {
	var reason string

	err := wait.PollImmediate(pollInterval, c.Timeout, func() (bool, error) {
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
