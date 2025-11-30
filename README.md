# gostration
// TODO: 

## Getting Started

### Prerequisites
- go (version v1.24.6+).
- docker (version 17.03+).
- kubectl (version v1.11.3+).

### To Deploy on the cluster
1. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/gostration:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

2. Install the CRDs into the cluster:

```sh
make install
```

3. Deploy the Manager to the cluster with image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/gostration:tag
```

> If you encounter RBAC errors, you may need to grant yourself cluster-admin privileges or be logged in as admin.

4. Create instances of your solution.
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> Ensure that the samples has default values to test it out.

### To Uninstall
1. Delete the instances (CRs) from the cluster:

```sh
kubectl delete -k config/samples/
```

2. Delete the APIs(CRDs) from the cluster:

```sh
make uninstall
```

3. UnDeploy the controller from the cluster:

```sh
make undeploy
```

## How to test in local

1. **Create a local cluster**
```sh
kind create cluster --name gostration
```

2. **Build and load the image**
```sh
make docker-build IMG=gostration:v1
kind load docker-image gostration:v1 --name gostration
```

3. **Deploy the controller**
```sh
make deploy IMG=gostration:v1
```

4. **Verify the deployment**
```sh
kubectl get pods -n gostration-system
```

5. **Run a test DAG**
```sh
kubectl apply -f config/samples/workflow_v1_dag_test.yaml
kubectl get dags
kubectl logs -n gostration-system -l control-plane=controller-manager
```

6. **Cleanup**
```sh
kind delete cluster --name gostration
```

## (WIP) Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/gostration:tag
```

> The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/gostration/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

