# Walk-through

If RBAC enabled, first make sure Spartakus has the appropriate permissions:

```bash
$ kubectl create role nodelister \
        --verb=get --verb=list \
        --resource=nodes

$ kubectl create rolebinding nodelisterbinding \
        --role=nodelister \
        --serviceaccount=default:default
```

Now you can launch it:

```bash
$ kubectl run spartakus \
    --image=gcr.io/google_containers/spartakus-amd64:v1.0.0 \
    -- volunteer --cluster-id=$(uuidgen)
```

Check if it works properly by finding the Spartakus pod:

```bash
$ kubectl get po | grep sparta
spartakus-76bddd7bbf-4hllm   1/1       Running   0          5m
```

… and then exec into it to call it manually:

```bash
$ kubectl exec -it spartakus-76bddd7bbf-4hllm -- sh
~ $ /spartakus volunteer --cluster-id=1E679954-3666-4275-ACB2 --database=stdout
I0310 05:32:03.906610      28 database.go:42] using "stdout" database
I0310 05:32:03.907117      28 volunteer.go:65] started volunteer
{
    "version": "v1.0.0",
    "timestamp": "",
    "clusterID": "1E679954-3666-4275-ACB2",
    "masterVersion": "v1.9.2",
    "nodes": [
        {
            "id": "8158ffac04c90df934c1e47193f0d834",
            "operatingSystem": "linux",
            "osImage": "Docker for Mac",
            "kernelVersion": "4.9.75-linuxkit-aufs",
            "architecture": "amd64",
            "containerRuntimeVersion": "docker://18.3.0",
            "kubeletVersion": "v1.9.2",
            "capacity": [
                {
                    "resource": "cpu",
                    "value": "4"
                },
                {
                    "resource": "memory",
                    "value": "4042012Ki"
                },
                {
                    "resource": "pods",
                    "value": "110"
                }
            ]
        }
    ]
}
I0310 05:32:03.931955      28 volunteer.go:70] report successfully sent
I0310 05:32:03.931975      28 volunteer.go:75] next attempt in 24h0m0s
```