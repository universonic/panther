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

type overlayError struct {
	s string
}

func (e overlayError) Error() string {
	return e.s
}

func newOverlayError(s string) *overlayError {
	return &overlayError{s}
}

var (
	// ErrResourceNotFound is the error returned by storages if a resource cannot be found.
	ErrResourceNotFound = newOverlayError("Not found")

	// ErrResourceAlreadyExists is the error returned by storages if a resource ID has been taken during a creation.
	ErrResourceAlreadyExists = newOverlayError("ID or name already exists")
)

// IsInternalError checks if a given error is an internal error which is generally a storage error.
func IsInternalError(e error) bool {
	if _, ok := e.(*overlayError); !ok {
		return true
	}
	return false
}

// IsNotFound return true if a given error is ErrResourceNotFound
func IsNotFound(e error) bool {
	if e == ErrResourceNotFound {
		return true
	}
	return false
}

// IsConflict returns true if a given error is ErrResourceAlreadyExists
func IsConflict(e error) bool {
	if e == ErrResourceAlreadyExists {
		return true
	}
	return false
}
