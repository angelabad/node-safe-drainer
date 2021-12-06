# Warning

This repository contains experimental code. Use it at your own risk!

# Node Safe Drainer

Tool to safely drain Kubernetes nodes.

## The problem

When you initiate a drain of the full Kubernetes node, this node
doesn't start new pods before delete old ones, so if you only have one
pod your application will experience a downtime, for applications with
more than one pod you can prevent this behavior using pod disruption
budget, but this is impossible with only one pod.

There are times when we have development or integration environments
where web only have one pod per application but we don't want
downtime.

## The solution

Node safe drainer checks the nodes (which are indicated or all of
them), it takes note of the deployments in each of them, it cordons
the nodes to prevent new deployments from entering and then make a
rollout restart of each deployment, this way we ensure each pod
doesn't have downtime.

### Notes

If don't have an autoscaler for the nodes you must ensure that you
have enough capacity in your nodes to store new deployments,
otherwise, the nodes will be blocked until there is enough capacity.

## Compilation

```
$ git clone git@github.com:angelabad/node-safe-drainer.git
$ cd node-safe-dainer
$ go build
```

## Usage

```
usage: ./node-safe-drainer [OPTIONS] <COMMA_SEPPARATED_NODE_NAMES>

Simple tool for safe draining nodes, rolling out deployments without downtime.

Options:
  -all-nodes
    	cordon and empty all nodes (use with caution)
  -kubeconfig string
    	abslute path to the kubeconfig file (default "/home/angel/.kube/config")
  -max-jobs int
    	max concurrent rollouts. (default 10)
  -timeout duration
    	deployment rollouts timeout. (default 20m0s)
```

### For all nodes on cluster

```
$ ./node-safe-drainer -all-nodes
```

### For custom nodes

```
$ ./node-safe-drainer k3d-k3s-default-server-0,k3d-k3s-default-agent-0
```
