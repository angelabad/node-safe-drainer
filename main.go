package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/client-go/util/retry"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) abslute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, node := range nodes.Items {
		fmt.Printf("NODE: %s\n", node.Name)

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		if err != nil {
			panic(err.Error())
		}

		for _, pod := range pods.Items {
			if (pod.OwnerReferences[0].Kind == "ReplicaSet") {
				replica, err := clientset.AppsV1().ReplicaSets(pod.Namespace).Get(context.TODO(), pod.OwnerReferences[0].Name, metav1.GetOptions{})
				if err != nil {
					panic(err.Error())
				}
				fmt.Printf("Namespace: %s, Deployment: %s\n", pod.Namespace, replica.OwnerReferences[0].Name)

				retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					result, getErr := clientset.AppsV1().Deployments(replica.Namespace).Get(context.TODO(), replica.OwnerReferences[0].Name, metav1.GetOptions{})
					if getErr != nil {
						panic(getErr.Error())
					}
					ann := result.Spec.Template.GetAnnotations()
					if ann == nil {
						ann = make(map[string]string)
					}
					ann["roller.angelabad.me/restartedAt"] = time.Now().String()
					result.Spec.Template.Annotations = ann
					_, updateErr := clientset.AppsV1().Deployments(replica.Namespace).Update(context.TODO(), result, metav1.UpdateOptions{})
					return updateErr
				})
				if retryErr != nil {
					panic(fmt.Errorf("update failed: %v", retryErr))
				}
			}
		}
	}
}
