# oomeventer

## ⚠️ Important Notice

Starting from Kubernetes 1.28, when an OOM (Out Of Memory) event occurs inside a container, the default behavior is to
restart the entire Pod. In Kubernetes 1.32 and later, a new configuration option `singleProcessOOMKill` was introduced,
allowing you to control this behavior. With `singleProcessOOMKill` enabled, only the container that experienced the OOM
will be restarted, instead of the whole Pod. Please refer to the official Kubernetes documentation for details on
configuring `singleProcessOOMKill` to avoid unnecessary Pod restarts.

singleProcessOOMKill: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/#kubelet-config-k8s-io-v1beta1-KubeletConfiguration

## Project Overview

`oomeventer` is a Kubernetes OOM (Out Of Memory) event monitoring and reporting tool based on eBPF. It captures OOM
process information at the node level in real-time and reports events as Kubernetes Events, helping users quickly
identify and analyze memory overflow issues.

## Features

- Real-time monitoring of Linux system OOM events using eBPF technology
- Automatic parsing of OOM process details such as command line and environment variables
- Association with Kubernetes metadata like containers, Pods, and Namespaces
- Reporting OOM events as Kubernetes Events for unified cluster observation
- Support for automatic deployment on each node in the cluster via DaemonSet

## Architecture

1. Monitor kernel OOM events using eBPF programs to capture killed process information
2. Parse process cgroup information to extract container IDs
3. Retrieve container metadata through the CRI (e.g., containerd) interface
4. Combine with Kubernetes API to obtain Pod and Namespace information
5. Report events to Kubernetes for unified observation and alerting

## Quick Start

### 1. Build the Image

```bash
docker build -t oomeventer .
```

### 2. Deploy to Kubernetes

The project provides a `daemonset.yaml` file for deployment:

```bash
kubectl apply -f daemonset.yaml
```

### 3. View OOM Events

When an OOM event occurs on a node, you can view the events using:

```bash
kubectl get events --all-namespaces | grep OOM
```

## Local Development and Testing

### Dependencies

- Go 1.20+
- Docker
- Kubernetes cluster (local or remote)

### Build

```bash
./build.sh
```

### Unit Tests

```bash
go test ./...
```

### Simulating OOM Events

The project includes a simple OOM test script:

```bash
cd test
bash test.sh
```

This script continuously allocates memory in a container to trigger an OOM event.

## Directory Structure

- `cmd/oomeventer/`: Main program entry
- `bpf/`: eBPF-related code
- `kube/`: Kubernetes event reporting and utility methods
- `test/`: Test scripts and cases
- `daemonset.yaml`: Kubernetes DaemonSet deployment file
- `Dockerfile`: Image build file

## Contributing

Issues and PRs are welcome!

## License

This project is open-sourced under the MIT License.