package util

import (
	"github.com/appscode/go/encoding/yaml"
	crdv1alpha1 "k8s-initializer-finalizer-practice/pkg/apis/crd.emruz.com/v1alpha1"
	"k8s.io/client-go/kubernetes"

	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_v1 "k8s.io/api/core/v1"
)

type busyboxconfig struct{
	Containers []api_v1.Container
}
func AddSidecarBusyBox(clientset  kubernetes.Clientset  ,customdeployment *crdv1alpha1.CustomDeployment) (*crdv1alpha1.CustomDeployment,error){
	fmt.Println("SideCar Added in ",customdeployment.Name)

	// get busyboxconfig from configMap
	cm,err:=clientset.CoreV1().ConfigMaps(metav1.NamespaceDefault).Get("busybox-sidecar-configmap",metav1.GetOptions{})
	if err!=nil{
		fmt.Println("Can't get configmap. Reason: ",err)
		return nil,err
	}

	// Now get the sidecar configuration from configmap
	var sidecarconfig busyboxconfig
	err=yaml.Unmarshal([]byte(cm.Data["config"]),&sidecarconfig)
	if err!=nil{
		fmt.Println("Can't get sidecarconfig from busybox config. Reason: ",err)
		return nil,err
	}

	//now add the sidecar to deployment
	initalizedDeployment:=customdeployment.DeepCopy()
	initalizedDeployment.Spec.Template.Spec.Containers=append(customdeployment.Spec.Template.Spec.Containers,sidecarconfig.Containers...)

	return initalizedDeployment,nil
}

func RemoveInitializer(in metav1.ObjectMeta) metav1.ObjectMeta {
	if in.GetInitializers()!=nil{
		if len(in.GetInitializers().Pending)==1{
			in.Initializers=nil
		}else {
			in.Initializers.Pending=append(in.Initializers.Pending[:0],in.Initializers.Pending[1:]...)
		}
	}
	return in
}