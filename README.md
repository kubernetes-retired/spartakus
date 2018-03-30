# Spartakus
[![Build Status](https://travis-ci.org/kubernetes-incubator/spartakus.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/spartakus)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-incubator/spartakus)](https://goreportcard.com/report/github.com/kubernetes-incubator/spartakus)

> Project Spartakus aims at collecting usage information about Kubernetes clusters. This information will help the Kubernetes development team to understand what people are doing with it, and to make it a better product.

- [What is in a report?](#what-is-in-a-report)
- [How do I run it?](#how-do-i-run-it)
- [What will we do with this information?](#what-will-we-do-with-this-information)
- [User documentation](docs/)
- [Development](#development)
- [Future](#future)

Note: **Spartakus does not report any personal identifiable information (PII)**.  Anything that might be identifying, including IP addresses, container images, and object names are anonymized. We take this very seriously. If you think something we are collecting might be considered PII, please do let us know by raising an issue here.

Running Spartakus is a voluntary effort, that is, it is not baked into Kubernetes in any way, shape, or form. In other words: it operates on an opt-in basis—if you don't want to run it, you don't have to. If you want to run your own server and collect data yourself, you can do that, too—see also the [user docs](docs/) for more info on how to customize the reports.  

## What is in a report?

Reports include a user-provided cluster identifier, the version strings of your Kubernetes master, and some information about each node in the cluster, including the operating system version, `kubelet` version, container runtime version, as well as CPU and memory capacity.

An example report payload looks as follows:

```json
{
    "version": "v1.0.0",
    "timestamp": "867530909031976",
    "clusterID": "2f9c93d3-156c-47aa-8802-578ffca9b50e",
    "masterVersion": "v1.3.5",
    "nodes": [
        {
            "id": "c8863d09ecc5be8d9791f72acd275fc2",
            "operatingSystem": "linux",
            "osImage": "Debian GNU/Linux 7 (wheezy)",
            "kernelVersion": "3.16.0-4-amd64",
            "architecture": "amd64",
            "containerRuntimeVersion": "docker://1.11.2",
            "kubeletVersion": "v1.3.2",
            "cloudProvider": "aws",
            "capacity": [
                {
                    "resource": "alpha.kubernetes.io/nvidia-gpu",
                    "value": "0"
                },
                {
                    "resource": "cpu",
                    "value": "4"
                },
                {
                    "resource": "memory",
                    "value": "15437428Ki"
                },
                {
                    "resource": "pods",
                    "value": "110"
                }
            ]
        },
        {
            "id": "5b919a15947b0680277acddf68d4b7aa",
            "operatingSystem": "linux",
            "osImage": "Debian GNU/Linux 7 (wheezy)",
            "kernelVersion": "3.16.0-4-amd64",
            "architecture": "amd64",
            "containerRuntimeVersion": "docker://1.11.2",
            "kubeletVersion": "v1.3.2",
            "cloudProvider": "aws",
            "capacity": [
                {
                    "resource": "alpha.kubernetes.io/nvidia-gpu",
                    "value": "0"
                },
                {
                    "resource": "cpu",
                    "value": "4"
                },
                {
                    "resource": "memory",
                    "value": "15437428Ki"
                },
                {
                    "resource": "pods",
                    "value": "110"
                }
            ]
        }
    ]
}
```

## How do I run it?

To start using Spartakus in your cluster, use the following:

```bash
$ kubectl run spartakus \
    --image=gcr.io/google_containers/spartakus-amd64:v1.0.0 \
    -- volunteer --cluster-id=$(uuidgen)
```

This will generate a deployment called `spartakus` in your `default`
namespace which sends a report once every 24 hours. It will report a random
UUID as the cluster ID. Note that if you stop this deployment and re-run
it, the UUID will be different. Managing the UUIDs is outside of the scope of Spartakus.

If you want to save the YAML manifest that the command above produces, you can simply execute the following (note: the `--export` flag strips cluster-specific information):

```bash
$ kubectl get deployment spartakus --export -o yaml
```

You needn't worry about CPU and memory usage of Spartakus, its resource usage footprint is minimal. If you're still concerned, you can edit the deployment to request a small share of CPU and memory; for example, Spartakus will work fine with `1m` CPU and `10Mi` mem on a five-nodes cluster.

## What will we do with this information?

The information reported by Spartakus, in aggregate, will help the Kubernetes development teams prioritize our efforts. The better we understand what people are doing, the better we can focus on the important issues.

## Development

Note that you only will need this section if you're contributing to Spartakus.

To build artefacts, we're using make. Simply run `make` or `make test` to build the binary in `bin/`; the container image build is carried out through Docker and hence requires you've got Docker running on your machine.

### Future

Anything we add will follow the same strict privacy rules as outlined above.  Some examples of things we would consider adding:

- How many namespaces exist.
- A histogram of how many pods, services, deployments, etc. exist on a per-namespace basis.
- Average lifetime of namespaces, pods, services, etc.