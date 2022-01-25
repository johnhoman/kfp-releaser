# kfp-releaser
WIP

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
