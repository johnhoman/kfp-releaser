/*
Copyright 2022 John Homan.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PipelineVersionSpec defines the desired state of PipelineVersion
type PipelineVersionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PipelineVersion. Edit pipelineversion_types.go to remove/update
	Description string `json:"description,omitempty"`
	Pipeline    string `json:"pipeline"`
	//+kubebuilder:pruning:PreserveUnknownFields
	Workflow runtime.RawExtension `json:"workflow"`
}

// PipelineVersionStatus defines the observed state of PipelineVersion
type PipelineVersionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PipelineID string `json:"pipelineId,omitempty"`
	Name       string `json:"name,omitempty"`
	ID         string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.name`
//+kubebuilder:printcolumn:name="PipelineName",type=string,JSONPath=`.spec.pipeline`
//+kubebuilder:printcolumn:name="PipelineId",type=string,JSONPath=`.status.pipelineId`

// PipelineVersion is the Schema for the pipelineversions API
type PipelineVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineVersionSpec   `json:"spec,omitempty"`
	Status PipelineVersionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineVersionList contains a list of PipelineVersion
type PipelineVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineVersion{}, &PipelineVersionList{})
}
