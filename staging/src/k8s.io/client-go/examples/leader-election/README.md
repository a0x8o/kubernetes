# Leader Election Example

This example demonstrates how to use the leader election package.

## Running

Run the following three commands in separate terminals. Each terminal needs a unique `id`.

```bash
<<<<<<< HEAD
# first terminal 
go run *.go -kubeconfig=/my/config -logtostderr=true -id=1

# second terminal 
=======
# first terminal
go run *.go -kubeconfig=/my/config -logtostderr=true -id=1

# second terminal
>>>>>>> Merge branch 'master' of https://github.com/kubernetes/kubernetes
go run *.go -kubeconfig=/my/config -logtostderr=true -id=2

# third terminal
go run *.go -kubeconfig=/my/config -logtostderr=true -id=3
```
> You can ignore the `-kubeconfig` flag if you are running these commands in the Kubernetes cluster.

Now kill the existing leader. You will see from the terminal outputs that one of the remaining two processes will be elected as the new leader.
