package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/api/core/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//Specification for CustomeDeployment
type CustomDeployment struct{
	metav1.TypeMeta					`json:",inline"`
	metav1.ObjectMeta 				`json:"metadata,omitempty"`

	Spec CustomDeploymentSpec 		`json:"spec"`
	Status CustomDeploymentStatus  `json:"status"`
}

//Specification for CustomDeployment
type CustomDeploymentSpec struct {
	Replicas int32 `json:"replicas"`
	Template CustomPodTemplate `json:"template"`
} 


//Status for CustomDeployment
type CustomDeploymentStatus struct{
	AvailableReplicas	int32 `json:"available_replicas"`
	CreatingReplicas	int32 `json:"creating_replicas"`
	TerminatingReplicas int32 `json:"terminating_replicas"`
}

//Specification for DemoPod
type CustomPodTemplate struct{
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec apiv1.PodSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//List of CustomDeployment
type CustomDeploymentList struct{
	metav1.TypeMeta	`json:",inline"`
	metav1.ListMeta	`json:"metadata"`

	Items []CustomDeployment `json:"items"`
}