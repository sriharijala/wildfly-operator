package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WildflyAppServerList struct {
	metav1.TypeMeta          `json:",inline"`
	metav1.ListMeta          `json:"metadata"`
	Items []WildflyAppServer `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WildflyAppServer struct {
	metav1.TypeMeta               `json:",inline"`
	metav1.ObjectMeta             `json:"metadata"`
	Spec   WildflyAppServerSpec   `json:"spec"`
	Status WildflyAppServerStatus `json:"status,omitempty"`
}

type WildflyAppServerSpec struct {
	NodeCount           int32                     `json:"nodeCount"`
	Image               string                    `json:"image"`
	ApplicationPath     string                    `json:"applicationPath"`
	ConfigMapName       string                    `json:"configMapName,omitempty"`
	StandaloneConfigKey string                    `json:"standaloneConfigKey,omitempty"`
	DataSourceConfig    map[string]DataSourceSpec `json:"dataSourceConfig"`
}

type DataSourceSpec struct {
	HostName     string `json:"hostName"`
	DatabaseName string `json:"databaseName"`
	JndiName     string `json:"jndiName"`
	User         string `json:"user,omitempty"`
	Password     string `json:"password,omitempty"`
}

type WildflyAppServerStatus struct {
	Nodes             []string          `json:"nodes"`
	ExternalAddresses map[string]string `json:"externalAddresses"`
}
