# kube-secrets

Kube-secrets solves the problem of storing your secrets in your Kube configs by accessing secrets in AWS Secrets Manager and applying them as Kubernetes Secrets. 

Once referenced from your deployment, each secret key/value will be injected as Environment variables into your pods. (Note: this will require you restart any existing pods to get the updated environment variables).

If AWS Secrets Manager is not your thing or you are not using AWS, you could easily swap aws-secrets.go for Hashicorp Vault.


How it works:
- kube-secrets runs as a pod in cluster and loops over all available namespaces (Excluding defaults like kube-system, kube-public etc.)

- For each namepsace it attempts to find a cooresponding secret in AWS Secrets Manager and creates/updates a Kubernetes secret. (Note: this automatically base64 encodes your secret values. The Kubernetes API does not check if a secret is already encoded) 

- Example: secret name "kubernetes/testing-stage" matches namespace "testing-stage" and it will create/update a secret in testing-stage namespace named testing-stage-secret with those key/values.

- Only need to reference the Kubernetes Secret in your Deployment and kube-secrets will take care of creating and updating the Secret Dynamically.



### Example Kubernetes  (Note the "envFrom" key)


```
---

apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: debian-slim
  namespace: testing-stage
  labels:
    app: debian-slim
spec:
  replicas: 1
  selector:
    matchLabels:
      app: debian-slim
  template:
    metadata:
      labels:
        app: debian-slim
    spec:
      containers:
        - name: app
          image: debian:stretch-slim
          command: ["sleep", "3600"]
          imagePullPolicy: Always
          envFrom:
          - secretRef:
              name: testing-stage-secret
```

### Secret that kube-secrets creates and manages looks something like this

```
apiVersion: v1
data:
  CLUSTER: RG9ja2VyRGVza3RvcA==
  USER: Y29mZmVl
kind: Secret
metadata:
  creationTimestamp: 2018-06-13T16:01:41Z
  name: testing-stage
  namespace: testing-stage
  resourceVersion: "141850"
  selfLink: /api/v1/namespaces/supersecrets/secrets/supersecrets-secret
  uid: 0fa1c24e-6f23-11e8-9b57-025000000001
type: Opaque
```
