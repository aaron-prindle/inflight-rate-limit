/*
Copyright 2019 The Kubernetes Authors.

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

package inflight

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityLevelConfiguration represents the configuration of a priority level.
type PriorityLevelConfiguration struct {
	metav1.TypeMeta
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta
	// Specification of the desired behavior of a request-priority.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PriorityLevelConfigurationSpec
	// Current status of a request-priority.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status PriorityLevelConfigurationStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityLevelConfigurationList is a list of PriorityLevelConfiguration objects.
type PriorityLevelConfigurationList struct {
	metav1.TypeMeta
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta
	// Items is a list of request-priorities.
	Items []PriorityLevelConfiguration
}

// PriorityLevelConfigurationSpec is specification of a priority level
type PriorityLevelConfigurationSpec struct {
	// GlobalDefault specifies whether this priority level should be considered as the default priority for requests
	// that do not have any priority level. Only one PriorityClass should be marked as `globalDefault`.
	// +optional
	GlobalDefault bool
	// Exempt defines whether the priority level is exempted or not.  There should be at most one exempt priority level.
	// Being exempt means that requests of that priority are not subject to concurrency limits (and thus are never queued)
	// and do not detract from the concurrency available for non-exempt requests. In a more sophisticated system, the
	// exempt priority level would be the highest priority level. The field is default to false and only those system
	// preset priority level can be exempt.
	// +optional
	Exempt bool
	// AssuredConcurrencyShares is a positive number for a non-exempt priority level, representing the weight by which
	// the priority level shares the concurrency from the global limit. The concurrency limit of an apiserver is divided
	// among the non-exempt priority levels in proportion to their assured concurrency shares. Basically this produces
	// the assured concurrency value (ACV) for each priority level:
	//
	//             ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )
	//
	AssuredConcurrencyShares int32
	// Queues is a number of queues that belong to a non-exempt PriorityLevelConfiguration object. The queues exist independently
	// at each apiserver.
	Queues int32
	// HandSize is a small positive number for applying shuffle sharding. When a request arrives at an apiserver the
	// request flow identifier’s string pair is hashed and the hash value is used to shuffle the queue indices and deal
	// a hand of the size specified here. If empty, the hand size will the be set to 1 at which  the priority drains the
	// queues on a FIFO basis.
	// +optional
	HandSize int32
	// QueueLengthLimit is a length limit applied to each queue belongs to the priority.
	// +optional
	QueueLengthLimit int32
}

// PriorityLevelConfigurationConditionType is a valid value for PriorityLevelConfigurationStatusCondition.Type
type PriorityLevelConfigurationConditionType string

// PriorityLevelConfigurationStatus represents the current state of a request-priority.
type PriorityLevelConfigurationStatus struct {
	// Current state of request-priority.
	Conditions []PriorityLevelConfigurationStatusCondition
}

// PriorityLevelConfigurationStatusCondition defines the condition of priority level.
type PriorityLevelConfigurationStatusCondition struct {
	// Type is the type of the condition.
	Type PriorityLevelConfigurationConditionType
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status api.ConditionStatus
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string
	// Human-readable message indicating details about last transition.
	Message string
}
