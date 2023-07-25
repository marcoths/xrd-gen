// +groupName=example.io
package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:resource:shortName=rbt,scope=Namespaced
// +kubebuilder:object:root=true
// Robot is a sample API type
type Robot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RobotSpec   `json:"spec,omitempty"`
	Status            RobotStatus `json:"status,omitempty"`
}

// +kubebuilder:subresource:spec
// SpaceshipSpec defines the desired state of Robot
type RobotSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Optional
	Age int `json:"age,omitempty"`
}

// +kubebuilder:subresource:status
// SpaceshipStatus defines the observed state of Robot
type RobotStatus struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:
	Active bool `json:"active,omitempty"`
}
