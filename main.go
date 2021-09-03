package main

import (
	"fmt"
	"os"
	"path/filepath"

	"angelabad.me/node-safe-drain/utils"

	"angelabad.me/node-safe-drain/client"
	flag "github.com/spf13/pflag"
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
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
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

func usage() {
	msg := `usage: node-safe-drainer [OPTIONS] <COMMA_SEPPARATED_NODE_NAMES>
       Simple tool for safe draining nodes, rolling out deployments without donwtime.
	   `
	fmt.Println(msg)
	flag.PrintDefaults()
}
