package main

import (
	"flag"
	"os"
	"path/filepath"

	"angelabad.me/node-safe-drainer/client"
	"angelabad.me/node-safe-drainer/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "abslute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	maxJobs := flag.Int("max-jobs", 10, "max concurrent rollouts.")
	flag.Usage = utils.Usage
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	nodes := utils.ParseArgs(flag.Arg(0))
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
		MaxJobs:   *maxJobs,
	}

	if err := client.CordonAndEmpty(nodes); err != nil {
		panic(err.Error())
	}
}
