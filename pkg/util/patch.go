package util

import (
	"encoding/json"

	clientversioned "k8s-initializer-finalizer-practice/pkg/client/clientset/versioned"
	"github.com/appscode/go/log"
	crdv1alpha1 "k8s-initializer-finalizer-practice/pkg/apis/crd.emruz.com/v1alpha1"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
)
func PatchCustomDeployment(c clientversioned.Clientset, cur *crdv1alpha1.CustomDeployment, transform func(*crdv1alpha1.CustomDeployment) *crdv1alpha1.CustomDeployment) (*crdv1alpha1.CustomDeployment,error)  {

	curJson,err:=json.Marshal(cur)
	if err!=nil{
		return nil,err
	}

	modJson,err:=json.Marshal(transform(cur.DeepCopy()))
	if err!=nil{
		return nil,err
	}

	patch, err:=jsonmergepatch.CreateThreeWayJSONMergePatch(curJson,modJson,curJson)
	if err!=nil{
		return nil,err
	}

	if len(patch)==0||string(patch)=="{}"{
		return cur,nil
	}

	log.Infoln("Patching ",cur.Name)
	return c.CrdV1alpha1().CustomDeployments(api_v1.NamespaceDefault).Patch(cur.Name,types.MergePatchType,patch)
}