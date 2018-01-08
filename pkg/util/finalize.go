package util

import (
	"k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_v1 "k8s.io/api/core/v1"
)

func DeletePods(kubeclient kubernetes.Clientset,podLabel string) error {
	podClient:= kubeclient.CoreV1().Pods(api_v1.NamespaceDefault)

	podList,err:=kubeclient.CoreV1().Pods(api_v1.NamespaceDefault).List(meta_v1.ListOptions{LabelSelector: podLabel})
	if err!=nil{
		return err
	}


	for _,pod:=range podList.Items{

		delErr:=podClient.Delete(pod.GetName(),&meta_v1.DeleteOptions{})

		if delErr!=nil{
			err=delErr
		}
	}

	return err
}

func RemoveFinalizer(in meta_v1.ObjectMeta) meta_v1.ObjectMeta {
	if in.GetFinalizers()!=nil{
		if len(in.Finalizers)==1{
			in.Finalizers=nil
		}else {
			in.Finalizers=append(in.Finalizers[:0],in.Finalizers[1:]...)
		}
	}
	return in
}