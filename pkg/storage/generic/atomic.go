// Copyright 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generic

import "time"

// ObjectMeta is the metadata of objects
type ObjectMeta struct {
	// GUID as a unified identifier for the object
	GUID string `json:"guid,omitempty" protobuf:"bytes,1,req,name=guid"`
	// Kind represents the resource name of the object
	Kind string `json:"kind,omitempty" protobuf:"bytes,2,req,name=kind"`
	// Name is the name of the object
	Name string `json:"name,omitempty" protobuf:"bytes,3,req,name=name"`
	// Namespace is the name of the object. Generally it is the name of its owner.
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,4,name=namespace"`
	// CreatedAt indicates the creation timestamp
	CreatedAt Time `json:"created_at,omitempty" protobuf:"bytes,5,req,name=created_at"`
	// UpdatedAt indicates the updating timestamp
	UpdatedAt *Time `json:"updated_at,omitempty" protobuf:"bytes,6,opt,name=updated_at"`
}

// SetGUID set the GUID for an object
func (in *ObjectMeta) SetGUID(id string) {
	in.GUID = id
}

// GetGUID returns the GUID of an object
func (in *ObjectMeta) GetGUID() string {
	return in.GUID
}

// SetKind set the Kind for an object
func (in *ObjectMeta) SetKind(kind string) {
	in.Kind = kind
}

// GetKind returns the Kind of an object
func (in *ObjectMeta) GetKind() string {
	return in.Kind
}

// SetName set the Name for an object
func (in *ObjectMeta) SetName(name string) {
	in.Name = name
}

// GetName returns the Name of an object
func (in *ObjectMeta) GetName() string {
	return in.Name
}

// SetNamespace set the Namespace for an object
func (in *ObjectMeta) SetNamespace(ns string) {
	in.Namespace = ns
}

// GetNamespace returns the Namespace of an object
func (in *ObjectMeta) GetNamespace() string {
	return in.Namespace
}

// SetCreationTimestamp set the CreatedAt for an object
func (in *ObjectMeta) SetCreationTimestamp(timestamp time.Time) {
	in.CreatedAt = Time{timestamp}
}

// GetCreationTimestamp returns the creation timestamp of an object
func (in *ObjectMeta) GetCreationTimestamp() time.Time {
	return in.CreatedAt.Time
}

// SetUpdatingTimestamp set the UpdatedAt for an object
func (in *ObjectMeta) SetUpdatingTimestamp(timestamp time.Time) {
	in.UpdatedAt = &Time{timestamp}
}

// GetUpdatingTimestamp returns the updating timestamp of an object
func (in *ObjectMeta) GetUpdatingTimestamp() *time.Time {
	if in.UpdatedAt == nil {
		return nil
	}
	return &in.UpdatedAt.Time
}

// HasNamespace returns true if object is namespace-sensitive. This could be overriden
// within specific object.
func (in *ObjectMeta) HasNamespace() bool { return false }

// ObjectListMeta is the metadata of object list
type ObjectListMeta struct {
	// Kind represents the resource name of the object
	Kind string `json:"kind,omitempty"`
	// Isolated represents it is a list of object with namespace
	Isolated bool `json:"isolated,omitempty"`
}

// GetKind returns the Kind of an object list
func (in *ObjectListMeta) GetKind() string {
	return in.Kind
}

// HasNamespace returns true if it has namespace
func (in *ObjectListMeta) HasNamespace() bool {
	return in.Isolated
}

// Time is a wrapper around time.Time which supports correct
// marshaling to YAML and JSON.  Wrappers are provided for many
// of the factory methods that the time package offers.
//
// +protobuf.options.marshal=false
// +protobuf.as=Timestamp
// +protobuf.options.(gogoproto.goproto_stringer)=false
type Time struct {
	time.Time `protobuf:"-"`
}
