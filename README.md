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


