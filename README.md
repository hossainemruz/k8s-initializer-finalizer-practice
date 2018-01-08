Custom Kubernetes controller for CRD. It is a simplified deployment controller.

### Commands:
Define CRD:
```
kubectl create -f ./yaml/crd-customdeployment.yaml
```
Create custom deployment:
```
kubectl create -f ./yaml/customdeployment.yaml
```
Run controller:
```
go run main.go
```

### Yaml for CRD:
```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: customdeployments.crd.emruz.com
spec:
  group: crd.emruz.com
  version: v1alpha1
  scope: Namespaced
  names:
    plural: customdeployments
    singular: customdeployments
    kind: CustomDeployment
    shortNames:
      - csd
```

### Yaml for custom-deployment:
```yaml
apiVersion: "crd.emruz.com/v1alpha1"
kind: CustomDeployment
metadata:
  name:  my-customdeployment
spec:
  replicas: 7
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
     containers:
      - name: webcalculator
        image: emruzhossain/webcalculator:latest
        ports:
        - name:  container-port
          containerPort:  9000
          protocol: TCP
```
