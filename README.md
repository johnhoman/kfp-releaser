# kfp-releaser
`kfp-releaser` is a Kubernetes [controller] for [Kubeflow] pipeline deployments.
The release workflow needs a few components to effectively manage pipeline deployments
from git.

Since every deployment should be tracked, trigger and provisioned from version control
the controller was designed with GitOps in mind so a typical workflow would use the kfp [compiler] to create the
workflow (pipeline spec), [ytt](https://carvel.dev/ytt/) to render the `pipelineversions.kfp.jackhoman.com` manifest, git
to store the manifest and argo to release the manifest. Once the manifest reaches the cluster the controller
will handle syncing the pipelines with [Kubeflow].


## Setup
### Install Argo
```shell
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```
Argo is used to install manifest files from the upstream GitHub repo holding the pipeline manifests. Each
manifest should be tracked in separate manifests so that all the Pipeline versions are always available on Kubeflow
(if that's what's desired -- if not they can be removed/updated)
e.g. 
```shell
cat > pipelines/whalesay/manifest.yaml <<EOF
apiVersion: kfp.jackhoman.com/v1alpha1
kind: Pipeline
metadata:
  name: whalesay
spec:
  description: Shows a really exciting picture of a whale
EOF
cat > pipelines/whalesay/versions/v1.0.1/manifest.yaml <<EOF
apiVersion: kfp.jackhoman.com/v1alpha1
kind: PipelineVersion
metadata:
  name: whalesay-v0.1.0
spec:
  pipeline: whalesay
  workflow:
    apiVersion: argoproj.io/v1alpha1
    kind: Workflow
    metadata:
      name: whalesay
    spec:
      entrypoint: whalesay
      templates:
      - name: whalesay
        container:
          image: docker/whalesay
          command: [cowsay]
          args: ["hello world"]
EOF
```

```shell
git add pipelines/whalesay && git commit -m "new whalesay pipeline"
git push
```

Once those changes are in git you can create an argo application for the pipelines
folder

```shell
cat > kubeflow apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1                                                                                                                                                              
kind: Application                                                                                                                                                                             
metadata:                                                                                                                                                                                     
  name: whalesay                                                                                                                                                                              
  namespace: argocd                                                                                                                                                                           
spec:                                                                                                                                                                                         
  destination:                                                                                                                                                                                
    namespace: kubeflow-examples                                                                                                                                                              
    server: https://kubernetes.default.svc                                                                                                                                                    
  project: default                                                                                                                                                                            
  source:                                                                                                                                                                                     
    path: pipelines/whalesay                                                                                                                                                                            
    repoURL: https://github.com/example/sample-pipelines                                                                                                                                       
    targetRevision: HEAD                                                                                                                                                                      
EOF
```

## Custom Resource

```yaml
apiVersion: kfp.jackhoman.com/v1alpha1
kind: Pipeline
metadata:
  name: whalesay
  namespace: production
spec:
  description: Cowsay pipeline
---
apiVersion: kfp.jackhoman.com/v1alpha1
kind: PipelineVersion
metadata:
  generateName: whalesay-
  namespace: production
  annotations:
    kfp.jackhoman.com/pipelin-version: v1.0.0
spec:
  pipeline: whalesay
  template:
    apiVersion: argoproj.io/v1alpha1
    kind: Workflow
    metadata:
      name: whalesay
    spec:
      entrypoint: whalesay
      templates:
      - name: whalesay
        container:
          image: docker/whalesay
          command: [cowsay]
          args: ["hello world"]
```

[Controller]: https://kubernetes.io/docs/concepts/architecture/controller
[Kubeflow]: https://kubeflow.org
[compiler]: https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.compiler.html