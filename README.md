This is a sample implementation of a k8sgpt [custom analyzer](https://docs.k8sgpt.ai/tutorials/custom-analyzers/) for argocd and image-updater.

## Configure argocd

See [argocd docs](https://argo-cd.readthedocs.io/en/stable/getting_started/) for installing and configuring argocd, and
[User Management](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/) for adding and managing
local users. The key steps are as follow:

### install argocd
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

### expose argocd-server service
Expose `argocd-server` service by changing its type from `ClusterIP` to `LoadBalancer`, or using port-forward.
Set the environment variable `ARGOCD_SERVER` to the argocd server address.
```bash
# port-forward 8080:443 in a separate terminal panel
kubectl port-forward -n argocd svc/argocd-server 8080:443 &
export ARGOCD_SERVER=localhost:8080
```

### login as argocd admin
Login from argocd cli as user `admin`. If admin account password has been updated, use the updated password when prompted.
Otherwise, get the initial admin password first, and copy it for login.
```bash
argocd admin -n argocd initial-password
argocd login localhost:8080 --insecure --skip-test-tls --username admin
```

### create and configure local user `test`
Create a new local user `test` for login and apiKey access, by editing `argocd-cm` ConfigMap, and adding the following
data section to `argocd-cm` ConfigMap.
```bash
kubectl edit -n argocd cm argocd-cm
```
```yaml
data:
  # add an additional local user with apiKey and login capabilities
  #   apiKey - allows generating API keys
  #   login - allows to login using UI
  accounts.test: apiKey, login
  accounts.test.enabled: "true"
```
Set the password for the new local user `test` while logged in as `admin`:
```bash
argocd account update-password --account test
```

### generate argocd auth token for local user `test` while logged in as `admin`, and set it to the environment variable
`ARGOCD_TOKEN`
```bash
argocd account generate-token --account test
export ARGOCD_TOKEN='<auth-token>'
```

## Configure and start ollama as a local AI provider
You can use various AI providers with k8sgpt, including hosted and local providers. Run `k8sgpt auth list` to view 
available providers. Start ollama in a separate terminal panel so it can be used the local provider in later steps.
```bash
ollama start

time=2025-06-23T13:55:51.340-04:00 level=INFO source=images.go:463 msg="total blobs: 12"
time=2025-06-23T13:55:51.340-04:00 level=INFO source=images.go:470 msg="total unused blobs removed: 0"
time=2025-06-23T13:55:51.340-04:00 level=INFO source=routes.go:1300 msg="Listening on 127.0.0.1:11434 (version 0.6.8)"
time=2025-06-23T13:55:51.398-04:00 level=INFO source=types.go:130 msg="inference compute" id=0 library=metal variant="" compute="" driver=0.0 name="" total="27.0 GiB" available="27.0 GiB"
```
Take note of the provider url: `127.0.0.1:11434`, which is used when adding this provider to k8sgpt.

## Configure and run k8sgpt
```bash
# add the above ollama local provider with `deepseek-r1` model to k8sgpt
# If `deepseek-r1` model is not available, you will need to install it to ollama first.
k8sgpt auth add --backend ollama --model deepseek-r1 --baseurl http://localhost:11434

# to analyze all resources in the cluster
k8sgpt analyse --explain --backend ollama

# to analyze resources in `argocd` namespace only
k8sgpt analyse --explain --backend ollama --namespace argocd
```

## Resources
* [k8sgpt](https://k8sgpt.ai/)
* [k8sgpt docs](https://k8sgpt.ai/docs)
* [argocd docs](https://argo-cd.readthedocs.io/en/stable/)
* [argocd image updater](https://github.com/argoproj-labs/argocd-image-updater)

