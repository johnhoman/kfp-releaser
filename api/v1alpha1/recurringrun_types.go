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
)

type RecurringRunSchedule struct {
	Cron string `json:"cron,omitempty"`
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RecurringRunSpec defines the desired state of RecurringRun
type RecurringRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of RecurringRun. Edit recurringrun_types.go to remove/update
	Schedule   RecurringRunSchedule `json:"schedule,omitempty"`
	VersionRef string               `json:"versionRef,omitempty"`
	Parameters []Parameter          `json:"parameters,omitempty"`
}

// RecurringRunStatus defines the observed state of RecurringRun
type RecurringRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID         string `json:"id,omitempty"`
	PipelineID string `json:"pipelineId,omitempty"`
	VersionID  string `json:"versionId,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule.cron`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.versionRef`

// RecurringRun is the Schema for the recurringruns API
type RecurringRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecurringRunSpec   `json:"spec,omitempty"`
	Status RecurringRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RecurringRunList contains a list of RecurringRun
type RecurringRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RecurringRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RecurringRun{}, &RecurringRunList{})
}
