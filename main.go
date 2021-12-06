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

package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"angelabad.me/node-safe-drainer/client"
	"angelabad.me/node-safe-drainer/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	defaultMaxJobs = 10
	defaultTimeout = 20 * time.Minute
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "abslute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	maxJobs := flag.Int("max-jobs", defaultMaxJobs, "max concurrent rollouts.")
	timeout := flag.Duration("timeout", defaultTimeout, "deployment rollouts timeout.")
	allNodes := flag.Bool("all-nodes", false, "cordon and empty all nodes (use with caution)")
	flag.Usage = utils.Usage
	flag.Parse()

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
		Timeout:   *timeout,
	}

	var nodes []string
	if *allNodes {
		nodes, err = client.GetAllNodes()
		if err != nil {
			panic(err.Error())
		}
	} else if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	} else {
		nodes = utils.ParseArgs(flag.Arg(0))
	}

	if err := client.CordonAndEmpty(nodes); err != nil {
		panic(err.Error())
	}
}
