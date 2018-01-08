Custom Kubernetes controller for CRD with Initializer and Finalizer. It is simple custom deployment type
controller.The initializer add ```busybox```pod as sidecar and add finalizer to delete all pods when customdeployment is deleted.

### Note:
Initializers are alpha feature hence we need to enable it manually. For minikube use the flowing command to
to start minikube with initializer enabled.
```
minikube start --extra-config=apiserver.Admission.PluginNames="Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota"
```

## Commands:
#### Step 1: Define CRD:
```
kubectl create -f ./yaml/crd-customdeployment.yaml
```

##### Yaml for CRD:
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

### Step 2: Create initializer configuration:
```
kubectl create -f ./yaml/initializerConfiguration.yaml
```
##### Yaml for Initizer configuration:
```yaml
apiVersion: admissionregistration.k8s.io/v1alpha1
kind: InitializerConfiguration
metadata:
  name: custom-deployment-initializer
initializers:
  - name: addbusybox.crd.emruz.com
    rules:
      - apiGroups:
          - crd.emruz.com
        apiVersions:
          - v1alpha1
        resources:
          - customdeployments
  - name: addfinalizer.crd.emruz.com
    rules:
      - apiGroups:
          - crd.emruz.com
        apiVersions:
          - v1alpha1
        resources:
          - customdeployments

```

### Step 3: Create Busybox configmap:
```
kubectl create -f ./yaml/busybox-configmap.yaml
```

##### Yaml for Busybox configmap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: busybox-sidecar-configmap
data:
  config: |
    containers:
      - name: busybox-sidecar
        image: busybox:glibc
        imagePullPolicy: IfNotPresent
        command: ["sh","-c","while true; do date; sleep 5; done"]
```
### Step 4: Run controller:
```
go run main.go
```
### Step 5: Create custom deployment:
```
kubectl create -f ./yaml/customdeployment.yaml
```


#### Yaml for custom-deployment:
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
