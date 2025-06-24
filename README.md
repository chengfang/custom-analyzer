This is a sample implementation of a k8sgpt [custom analyzer](https://docs.k8sgpt.ai/tutorials/custom-analyzers/) for argocd and image-updater.

## Install and configure argocd

See [argocd docs](https://argo-cd.readthedocs.io/en/stable/getting_started/) for installing and configuring argocd, and
[User Management](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/) for adding and managing
local users. The key steps are as follows:

### install argocd
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```
The rest of the argocd configuration steps are needed if you're using `argocd` client, instead of `kubernetes` client,
to connect to the target cluster.

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

## Optionally, configure and start ollama as a local AI provider
You can use various AI providers with k8sgpt, including hosted and local providers. Run `k8sgpt auth list` to view 
available providers. Start ollama in a separate terminal panel so it can be used the local provider in later steps.
```bash
ollama start

time=2025-06-23T13:55:51.340-04:00 level=INFO source=images.go:463 msg="total blobs: 12"
time=2025-06-23T13:55:51.340-04:00 level=INFO source=images.go:470 msg="total unused blobs removed: 0"
time=2025-06-23T13:55:51.340-04:00 level=INFO source=routes.go:1300 msg="Listening on 127.0.0.1:11434 (version 0.6.8)"
time=2025-06-23T13:55:51.398-04:00 level=INFO source=types.go:130 msg="inference compute" id=0 library=metal variant="" compute="" driver=0.0 name="" total="27.0 GiB" available="27.0 GiB"

# add the above ollama local provider with `deepseek-r1` model to k8sgpt
# If `deepseek-r1` model is not available, you will need to install it to ollama first.
k8sgpt auth add --backend ollama --model deepseek-r1 --baseurl http://localhost:11434
```

## Optionally, configure a remote hosted AI provider, such as google gemini
```bash
k8sgpt auth add --backend google --model gemini-2.5-flash
```

## Configure and run k8sgpt to perform analysis, without custom analyzer for testing
```bash
# to analyze all resources in the cluster, using one of the configured AI providers (e.g., google | ollama)
k8sgpt analyse --explain --backend google

# to analyze resources in `argocd` namespace only
k8sgpt analyse --explain --backend google --namespace argocd

# to limit the analysis to certain resource types only, e.g., Pod
k8sgpt analyze --explain --backend google --filter Pod

# to list all available filters
k8sgpt filters list
```
## Configure and run image-updater-analyzer
```bash
# build this project
go build

# add image-updater-analyzer to k8sgpt
k8sgpt custom-analyzer add -n image-updater-analyzer
  image-updater-analyzer added to the custom analyzers config list

# list configured custom analyzers
k8sgpt custom-analyzer list
  Active:
    > image-updater-analyzer
# run image-updater-analyzer in a separate terminal panel
./custom-analyzer
  starting image-updater-analyzer at :8085
  
# run k8sgpt analysis with image-updater-analyzer
k8sgpt k8sgpt analyze --custom-analysis --explain --backend google --filter Pod --namespace argocd
```
For an image-updater application missing the required `image-list` annotation, the above analysis would produce the 
following errors and solutions:
```bash
 100% |████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████| (1/1, 12 it/min)
AI Provider: google

0: image-updater-analyzer image-updater-analyzer()
- Error: The required annotation not found.
Suggestion:
 Add annotation argocd-image-updater.argoproj.io/image-list to the application argocd/image-list-kustomize
Error: Argo CD Image Updater needs `argocd-image-updater.argoproj.io/image-list` annotation on your app to track images. It's missing.
Solution:
1. Edit your Argo CD Application manifest.
2. Add `metadata.annotations` block.
3. Add `argocd-image-updater.argoproj.io/image-list: |` with image details.
```
The output in `image-updater-analyzer` window could also be of help for troubleshooting:
```bash
Running image-updater-analyzer
Creating Kubernetes client with kubeConfigPath:  and argocdNamespace: argocd
Getting application list

Verifying application image-list-kustomize
-------------------------
✓ Verified annotation: argocd-image-updater.argoproj.io/nginx.platforms: linux/arm64,linux/amd64
✓ Verified annotation: argocd-image-updater.argoproj.io/nginx2.platforms: linux/arm64,linux/amd64
✓ Verified annotation: argocd-image-updater.argoproj.io/update-strategy: digest
✓ Verified annotation: argocd-image-updater.argoproj.io/write-back-method: git:secret:argocd/git-creds
✓ Verified annotation: argocd-image-updater.argoproj.io/force-update: false
✗ The required annotation not found.
Suggestion:
 Add annotation argocd-image-updater.argoproj.io/image-list to the application argocd/image-list-kustomize

Verifying application write-helmvalues
-------------------------
✓ Verified annotation: argocd-image-updater.argoproj.io/force-update: false
✓ Verified annotation: argocd-image-updater.argoproj.io/git-branch: main
✓ Verified annotation: argocd-image-updater.argoproj.io/git-repository: https://github.com/chengfang/image-updater-examples.git
✓ Verified annotation: argocd-image-updater.argoproj.io/image-list: nginx=docker.io/bitnami/nginx:1.27.x
✓ Verified annotation: argocd-image-updater.argoproj.io/nginx.helm.image-tag: image.tag
✓ Verified annotation: argocd-image-updater.argoproj.io/nginx.helm.image-name: image.repository
✓ Verified annotation: argocd-image-updater.argoproj.io/update-strategy: semver
✓ Verified annotation: argocd-image-updater.argoproj.io/write-back-method: git:secret:argocd/git-creds
✓ Verified annotation: argocd-image-updater.argoproj.io/write-back-target: helmvalues:/write-helmvalues/source2/values.yaml

=========================
Consolidated results: Encountered error while verifying argocd/image-list-kustomize: The required annotation not found.
Suggestion:
 Add annotation argocd-image-updater.argoproj.io/image-list to the application argocd/image-list-kustomize
```

## Resources
* [k8sgpt](https://k8sgpt.ai/)
* [k8sgpt docs](https://k8sgpt.ai/docs)
* [argocd docs](https://argo-cd.readthedocs.io/en/stable/)
* [argocd image updater](https://github.com/argoproj-labs/argocd-image-updater)

