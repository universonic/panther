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
	"sort"
)

// Sortor is a generic object sortor
type Sortor struct {
	scope byte
	list  []SortableObject
}

func (in *Sortor) Len() int {
	return len(in.list)
}

func (in *Sortor) Swap(i, j int) {
	in.list[i], in.list[j] = in.list[j], in.list[i]
}

func (in *Sortor) Less(i, j int) bool {
	switch in.scope {
	case 1:
		return in.list[i].GetGUID() < in.list[j].GetGUID()
	case 2:
		return in.list[i].GetName() < in.list[j].GetName()
	case 3:
		return in.list[i].GetNamespace() < in.list[j].GetNamespace()
	case 4:
		return in.list[j].GetCreationTimestamp().After(in.list[i].GetCreationTimestamp())
	case 5:
		a := in.list[i].GetUpdatingTimestamp()
		b := in.list[j].GetUpdatingTimestamp()
		if b == nil {
			return false
		} else if a == nil {
			return true
		}
		return b.After(*a)
	default:
		return false
	}
}

// AppendMember appends a new object into sortor.
func (in *Sortor) AppendMember(obj SortableObject) {
	in.list = append(in.list, obj)
}

// SetMembers sets a list of object to sort.
func (in *Sortor) SetMembers(cv []SortableObject) {
	in.list = cv
}

// OrderByGUID issue a result that is ordered by GUID.
func (in *Sortor) OrderByGUID() []SortableObject {
	in.scope = 1
	sort.Sort(in)
	return in.list
}

// OrderByName issue a result that is ordered by Name.
func (in *Sortor) OrderByName() []SortableObject {
	in.scope = 2
	sort.Sort(in)
	return in.list
}

// OrderByNamespace issue a result that is ordered by Namespace.
func (in *Sortor) OrderByNamespace() []SortableObject {
	in.scope = 3
	sort.Sort(in)
	return in.list
}

// OrderByCreationTimestamp issue a result that is ordered by CreatedAt.
func (in *Sortor) OrderByCreationTimestamp() []SortableObject {
	in.scope = 4
	sort.Sort(in)
	return in.list
}

// OrderByUpdatingTimestamp issue a result that is ordered by UpdatedAt.
func (in *Sortor) OrderByUpdatingTimestamp() []SortableObject {
	in.scope = 5
	sort.Sort(in)
	return in.list
}

// Reset clean up the inner storage for reuse.
func (in *Sortor) Reset() {
	in.list = in.list[:0]
	in.scope = 0
}

// NewSortor initializes a new sorter.
func NewSortor() *Sortor {
	return new(Sortor)
}
