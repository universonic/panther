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

import (
	"encoding/json"
	"fmt"
)

const (
	// RESOURCE_HOST indicates the kind of a Host
	RESOURCE_HOST = "host"
	// RESOURCE_SYSTEM_SCAN indicates the kind of a SystemScan
	RESOURCE_SYSTEM_SCAN = "system_scan"
	// RESOURCE_HOST_OPERATION indicates the kind of a HostOperation
	RESOURCE_HOST_OPERATION = "host_operation"
)

// Host indicates host data object
type Host struct {
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	SSHAddress    string          `json:"ssh_addr,omitempty" protobuf:"bytes,2,opt,name=ssh_addr"`
	SSHPort       uint16          `json:"ssh_port,omitempty" protobuf:"varint,3,opt,name=ssh_port"`
	SSHCredential LoginCredential `json:"ssh_cred,omitempty" protobuf:"bytes,4,opt,name=ssh_cred"`
	OpCredential  LoginCredential `json:"op_cred,omitempty" protobuf:"bytes,5,opt,name=op_cred"`
	Comment       string          `json:"comment,omitempty" protobuf:"bytes,6,opt,name=comment"`
}

// Header returns a set of headers that will be used for generating ASCII table.
func (in *Host) Header() []string {
	return []string{"GUID", "Name", "SSH Address", "SSH Port", "SSH User", "Op User", "Comment", "Created At", "Updated At"}
}

// Row returns the value of object as a row of ASCII table.
func (in *Host) Row() (row []string) {
	row = []string{
		in.GetGUID(),
		in.GetName(),
		in.SSHAddress,
		fmt.Sprintf("%d", in.SSHPort),
		in.SSHCredential.User,
		in.OpCredential.User,
		string(in.Comment),
		in.GetCreationTimestamp().String(),
	}
	if in.GetUpdatingTimestamp() != nil && !in.GetUpdatingTimestamp().IsZero() {
		return append(row, in.UpdatedAt.String())
	}
	return append(row, "")
}

// LoginCredential indicates a pair of login user and password.
type LoginCredential struct {
	User     string `json:"user,omitempty" protobuf:"bytes,1,opt,name=user"`
	Password []byte `json:"pass,omitempty" protobuf:"bytes,2,opt,name=pass"`
}

// NewHost generates a new empty Host instance
func NewHost() *Host {
	return &Host{
		ObjectMeta: ObjectMeta{Kind: RESOURCE_HOST},
	}
}

// HostList indicates list of Host
type HostList struct {
	ObjectListMeta `json:",inline"`
	Members        []Host `json:"members,omitempty"`
}

// AppendRaw appends raw format data to object list, and returns any encountered error.
func (in *HostList) AppendRaw(dAtA []byte) error {
	cv := NewHost()
	if err := json.Unmarshal(dAtA, cv); err != nil {
		return err
	}
	in.Members = append(in.Members, *cv)
	return nil
}

// NewHostList generates a new empty HostList instance
func NewHostList() *HostList {
	return &HostList{
		ObjectListMeta: ObjectListMeta{
			Kind: RESOURCE_HOST,
		},
	}
}

// SystemScan indicates a set of available system update on a single host
type SystemScan struct {
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	State    State            `json:"state,omitempty" protobuf:"bytes,2,opt,name=state"`
	Security []SecurityUpdate `json:"security,omitempty" protobuf:"bytes,3,rep,name=security"`
}

// Header returns a set of headers that will be used for generating ASCII table.
func (in *SystemScan) Header() []string {
	return []string{"GUID", "Name", "State", "Security (Critical)", "Security (Important)", "Security (Moderate)", "Updated At"}
}

// Row returns the value of object as a row of ASCII table.
func (in *SystemScan) Row() (row []string) {
	var critical, important, moderate int
	for _, each := range in.Security {
		switch each.Severity {
		case CriticalSec:
			critical++
		case ImportantSec:
			important++
		case ModerateSec:
			moderate++
		}
	}
	row = []string{
		in.GetGUID(),
		in.GetName(),
		in.State.String(),
		fmt.Sprintf("%d", critical),
		fmt.Sprintf("%d", important),
		fmt.Sprintf("%d", moderate),
	}
	if in.GetUpdatingTimestamp() != nil && !in.GetUpdatingTimestamp().IsZero() {
		return append(row, in.UpdatedAt.String())
	}
	return append(row, "")
}

// SecurityUpdate is a single security update entity.
type SecurityUpdate struct {
	CVEID    string           `json:"cve_id,omitempty" protobuf:"bytes,1,opt,name=cve_id"`
	Severity SecuritySeverity `json:"severity,omitempty" protobuf:"bytes,2,opt,name=severity"`
	Package  string           `json:"package,omitempty" protobuf:"bytes,3,opt,name=package"`
}

// SecuritySeverity represents the severity of a security update.
type SecuritySeverity int

func (of SecuritySeverity) String() string {
	switch of {
	case UnknownSec:
		return "<null>"
	case CriticalSec:
		return "Critical"
	case ImportantSec:
		return "Important"
	case ModerateSec:
		return "Moderate"
	}
	return "<invalid>"
}

const (
	// UnknownSec indicates an unset security severity. This should not be presented in
	// common cases.
	UnknownSec SecuritySeverity = iota
	// CriticalSec indicates critical security severity.
	CriticalSec
	// ImportantSec indicates important security severity.
	ImportantSec
	// ModerateSec indicates moderate security severity.
	ModerateSec
)

// NewSystemScan generates a new empty SystemScan instance
func NewSystemScan() *SystemScan {
	return &SystemScan{
		ObjectMeta: ObjectMeta{Kind: RESOURCE_SYSTEM_SCAN},
	}
}

// SystemScanList indicates list of SystemScan
type SystemScanList struct {
	ObjectListMeta `json:",inline"`

	Members []SystemScan `json:"members,omitempty"`
}

// AppendRaw appends raw format data to object list, and returns any encountered error.
func (in *SystemScanList) AppendRaw(dAtA []byte) error {
	cv := NewSystemScan()
	if err := json.Unmarshal(dAtA, cv); err != nil {
		return err
	}
	in.Members = append(in.Members, *cv)
	return nil
}

// NewSystemScanList generates a new empty SystemScanList instance
func NewSystemScanList() *SystemScanList {
	return &SystemScanList{
		ObjectListMeta: ObjectListMeta{
			Kind: RESOURCE_SYSTEM_SCAN,
		},
	}
}

// State is the generic execution state
type State int

func (of State) String() string {
	switch of {
	case UnknownState:
		return "<null>"
	case StartedState:
		return "STARTED"
	case AbortState:
		return "ABORT"
	case InProgressState:
		return "IN-PROGRESS"
	case SuccessState:
		return "COMPLETED"
	case FailureState:
		return "FAILED"
	}
	return "<invalid>"
}

const (
	// UnknownState indicates an unset state. This should not be presented in common cases.
	UnknownState State = iota
	// StartedState indicates an initiated state.
	StartedState
	// AbortState indicates an abort state.
	AbortState
	// InProgressState indicates that a job is executing in progress.
	InProgressState
	// SuccessState indicates that a job has finished and succeeded.
	SuccessState
	// FailureState indicates that a job has finished and failed.
	FailureState
)

// OperationType is the original issuer of an operation.
type OperationType int

func (of OperationType) String() string {
	switch of {
	case UnknownOperation:
		return "<null>"
	case InternalOperation:
		return "internal"
	case UserOperation:
		return "user"
	}
	return "<invalid>"
}

const (
	// UnknownOperation indicates an unset operation type. This should not be presented in
	// common cases.
	UnknownOperation OperationType = iota
	// InternalOperation indicates the internal operation type.
	InternalOperation
	// UserOperation indicates the user-driven operation type.
	UserOperation
)

// OperationMethod is the method to perform the operation.
type OperationMethod int

func (of OperationMethod) String() string {
	switch of {
	case UnknownMethod:
		return "<null>"
	case RunMethod:
		return "run"
	case OutputMethod:
		return "output"
	case CombinedOutputMethod:
		return "combined_output"
	}
	return "<invalid>"
}

const (
	// UnknownMethod indicates an unset operation method. This should not be presented in
	// common cases.
	UnknownMethod OperationMethod = iota
	// RunMethod is a flag that drives executor to run a command with Run(). This method will
	// not return any details, but only the exit code and very few messages if failed.
	RunMethod
	// OutputMethod is a flag that drives executor to run a command with Output(). This method
	// will return stdout as data, or hints if failed.
	OutputMethod
	// CombinedOutputMethod is a flag that drives executor to run a command with CombinedOutput().
	// This method is similar to OutputMethod, but also includes stderr in data.
	CombinedOutputMethod
)

// HostOperation is a operation that is performed on a single host. It is namespace-sensitive,
// and its namespace's value is restricted to be the name of the host that the command to be
// exactly performed on. The data is either the returned result, or the reason of failure.
type HostOperation struct {
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Command string          `json:"command,omitempty" protobuf:"bytes,2,opt,name=command"`
	Type    OperationType   `json:"type,omitempty" protobuf:"bytes,3,opt,name=type"`
	Method  OperationMethod `json:"method,omitempty" protobuf:"bytes,4,opt,name=method"`
	Timeout uint            `json:"timeout,omitempty" protobuf:"varint,5,opt,name=timeout"`
	State   State           `json:"state,omitempty" protobuf:"bytes,6,opt,name=state"`
	Data    []byte          `json:"data,omitempty" protobuf:"bytes,7,opt,name=data"`
}

// Header returns a set of headers that will be used for generating ASCII table.
func (in *HostOperation) Header() []string {
	return []string{"GUID", "Host", "Command", "Type", "Method", "State", "Data", "Created At", "Updated At"}
}

// Row returns the value of object as a row of ASCII table.
func (in *HostOperation) Row() (row []string) {
	row = []string{
		in.GetGUID(),
		in.GetNamespace(),
		in.Command,
		in.Type.String(),
		in.Method.String(),
		in.State.String(),
		string(in.Data),
		in.GetCreationTimestamp().String(),
	}
	if in.GetUpdatingTimestamp() != nil && !in.GetUpdatingTimestamp().IsZero() {
		return append(row, in.UpdatedAt.String())
	}
	return append(row, "")
}

// HasNamespace returns true if object is namespace-sensitive
func (in *HostOperation) HasNamespace() bool { return true }

// NewHostOperation generates a new empty HostOperation instance
func NewHostOperation() *HostOperation {
	return &HostOperation{
		ObjectMeta: ObjectMeta{Kind: RESOURCE_HOST_OPERATION},
	}
}

// HostOperationList indicates list of HostOperation
type HostOperationList struct {
	ObjectListMeta `json:",inline"`

	Members []HostOperation `json:"members,omitempty"`
}

// AppendRaw appends raw format data to object list, and returns any encountered error.
func (in *HostOperationList) AppendRaw(dAtA []byte) error {
	cv := NewHostOperation()
	if err := json.Unmarshal(dAtA, cv); err != nil {
		return err
	}
	in.Members = append(in.Members, *cv)
	return nil
}

// NewHostOperationList generates a new empty HostOperationList instance
func NewHostOperationList() *HostOperationList {
	return &HostOperationList{
		ObjectListMeta: ObjectListMeta{
			Kind:     RESOURCE_HOST_OPERATION,
			Isolated: true,
		},
	}
}
