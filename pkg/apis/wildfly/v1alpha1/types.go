package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WildflyAppServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WildflyAppServer `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WildflyAppServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WildflyAppServerSpec   `json:"spec"`
	Status            WildflyAppServerStatus `json:"status,omitempty"`
}

type WildflyAppServerSpec struct {
	NodeCount   int32             `json:"nodeCount"`
	Image       string            `json:"image"`
	DataSources []DatasourceSpec  `json:"dataSources,omitempty"`
}

type DatasourceSpec struct {
	Type        string `json:"type"`
	JndiName    string `json:"jndiName"`
	ServiceName string `json:"serviceName,omitempty"`
	PortName    string `json:"portName,omitempty"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type WildflyAppServerStatus struct {
	Nodes []string `json:"nodes"`
}
