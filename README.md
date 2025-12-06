# pequod
// TODO: 

## Getting Started

### Prerequisites
- go (version v1.24.6+).
- docker (version 17.03+).
- kubectl (version v1.11.3+).

### To Deploy on cluster
1. Build & push your image to location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/pequod:tag
```

2. Install the CRDs into the cluster:

```sh
make install
```

3. Deploy the Manager to the cluster with image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/pequod:tag
```

> If you encounter RBAC errors, you may need to grant yourself cluster-admin privileges or be logged in as admin.

4. Create instances of your solution.
You can apply samples from config/sample:

```sh
kubectl apply -k config/samples/
```

> Ensure that samples has default values to test it out.

### To Uninstall
1. Delete the instances (CRs) from cluster:

```sh
kubectl delete -k config/samples/
```

2. Delete the APIs(CRDs) from cluster:

```sh
make uninstall
```

3. UnDeploy the controller from cluster:

```sh
make undeploy
```

## How to test in local

1. **Create a local cluster**
```sh
kind create cluster --name pequod
```

2. **Build and load image**
```sh
make docker-build IMG=pequod:v1
kind load docker-image pequod:v1 --name pequod
```

3. **Deploy controller**
```sh
make deploy IMG=pequod:v1
```

4. **Verify deployment**
```sh
kubectl get pods -n pequod-system
```

5. **Run a test DAG**
```sh
kubectl apply -f config/samples/workflow_v1_dag_test.yaml
kubectl get dags
kubectl logs -n pequod-system -l control-plane=controller-manager
```

6. **Cleanup**
```sh
kind delete cluster --name pequod
```
