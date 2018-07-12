package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PhilsThingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PhilsThing `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PhilsThing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PhilsThingSpec   `json:"spec"`
	Status            PhilsThingStatus `json:"status,omitempty"`
}

type PhilsThingSpec struct {
	ServiceClassName    string                 `json:"service_class_name"`
	Params              map[string]interface{} `json:"params"`
	ServiceInstanceName string                 `json:"service_instance_name"`
}

type PhilsThingStatus struct {
	Phase string `json:"phase"`
}
