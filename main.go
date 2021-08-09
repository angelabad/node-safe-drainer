package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"angelabad.me/node-safe-drain/client"
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
	if len(os.Args) != 2 {
		fmt.Println("Empty node name, provide it as an argument.")
		os.Exit(0)
	}

	nodeName := os.Args[1]
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	client := client.Client{
		Clientset: clientset,
	}

	if err := client.CordonAndEmpty(nodeName); err != nil {
		panic(err.Error())
	}
}
