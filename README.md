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
