// Copyright © 2018 Alfred Chou <unioverlord@gmail.com>
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

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	genericStorage "github.com/universonic/panther/pkg/storage/generic"
)

// Host handles requests from /api/v1/host.
// Usage:
//   - GET /api/v1/host?search=[HOST_LIST|*]
//   - POST /api/v1/host
//   - PUT /api/v1/host
//   - DELETE /api/v1/host?target=[HOST_NAME]
// TODO: make logger content qualified.
func (in *Handler) Host(w http.ResponseWriter, r *http.Request) {
	defer in.logger.Sync()
	defer func() {
		if rec := recover(); rec != nil {
			in.finalizeError(w, fmt.Errorf("Internal Server Error"), http.StatusInternalServerError)
			in.logger.Error(rec)
		}
	}()
	defer in.finalizeHeader(w)

	switch r.Method {
	case "GET":
		// GET implements host query process.
		targets := strings.Split(r.URL.Query().Get("search"), ",")
		if len(targets) == 1 && targets[0] == "" {
			in.finalizeError(w, fmt.Errorf("Target required"), http.StatusBadRequest)
			in.logger.Errorf("No target was specified during query")
			return
		}
		in.logger.Debugf("Search host: %s", strings.Join(targets, ", "))
		var all bool
		for _, each := range targets {
			if each == "*" {
				all = true
				break
			}
		}
		sortor := genericStorage.NewSortor()
		if all {
			cv := genericStorage.NewHostList()
			err := in.storage.List(cv)
			if err != nil {
				in.finalizeError(w, fmt.Errorf("Database Failure"), http.StatusInternalServerError)
				in.logger.Error(err)
				return
			}
			for i := range cv.Members {
				sortor.AppendMember(&cv.Members[i])
			}
		} else {
			for _, each := range targets {
				newObj := genericStorage.NewHost()
				newObj.Name = each
				err := in.storage.Get(newObj)
				if err != nil {
					in.logger.Error(err)
					if genericStorage.IsInternalError(err) {
						in.finalizeStorageError(w, err)
						return
					}
					in.finalizeError(w, fmt.Errorf("Database Failure"), http.StatusInternalServerError)
					return
				}
				sortor.AppendMember(newObj)
			}
		}
		dAtA, err := json.Marshal(sortor.OrderByName())
		if err != nil {
			panic(err)
		}
		in.finalizeJSON(w, bytes.NewReader(dAtA))
	case "POST":
		// POST implements host creation process.
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r.Body)
		if err != nil {
			panic(err)
		}
		cv := genericStorage.NewHost()
		err = json.Unmarshal(buf.Bytes(), cv)
		if err != nil {
			in.finalizeError(w, fmt.Errorf("Invalid Request Body"), http.StatusBadRequest)
			in.logger.Error(err)
			return
		}
		err = in.validateAndFulfillHost(cv)
		if err != nil {
			in.logger.Error(err)
			in.finalizeError(w, err, http.StatusBadRequest)
			return
		}
		err = in.storage.Create(cv)
		if err != nil {
			in.logger.Error(err)
			if genericStorage.IsInternalError(err) {
				in.finalizeStorageError(w, err)
				return
			}
			in.finalizeError(w, fmt.Errorf("Database Failure"), http.StatusInternalServerError)
			return
		}
		dAtA, err := json.Marshal(cv)
		if err != nil {
			panic(err)
		}
		in.finalizeJSON(w, bytes.NewReader(dAtA), http.StatusCreated)
	case "PUT":
		// PUT implements host updating process.
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r.Body)
		if err != nil {
			panic(err)
		}
		cv := genericStorage.NewHost()
		err = json.Unmarshal(buf.Bytes(), cv)
		if err != nil {
			in.finalizeError(w, fmt.Errorf("Invalid Request Body"), http.StatusBadRequest)
			in.logger.Error(err)
			return
		}
		err = in.validateAndFulfillHost(cv)
		if err != nil {
			in.logger.Error(err)
			in.finalizeError(w, err, http.StatusBadRequest)
			return
		}
		err = in.storage.Update(cv)
		if err != nil {
			in.logger.Error(err)
			if genericStorage.IsInternalError(err) {
				in.finalizeStorageError(w, err)
				return
			}
			in.finalizeError(w, fmt.Errorf("Database Failure"), http.StatusInternalServerError)
			return
		}
		dAtA, err := json.Marshal(cv)
		if err != nil {
			panic(err)
		}
		in.finalizeJSON(w, bytes.NewReader(dAtA))
	case "DELETE":
		// DELETE implements host deletion process.
		target := r.URL.Query().Get("target")
		if target == "" {
			in.finalizeError(w, fmt.Errorf("Target required"), http.StatusBadRequest)
			in.logger.Errorf("No machine target was specified")
			return
		}
		cv := genericStorage.NewHost()
		cv.Name = target
		err := in.storage.Delete(cv)
		if err != nil {
			in.logger.Error(err)
			if genericStorage.IsInternalError(err) {
				in.finalizeStorageError(w, err)
				return
			}
			in.finalizeError(w, fmt.Errorf("Database Failure"), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (in *Handler) validateAndFulfillHost(cv *genericStorage.Host) error {
	if ip := net.ParseIP(cv.SSHAddress); ip == nil {
		return fmt.Errorf("Invalid IP address: %s", cv.SSHAddress)
	}
	if cv.SSHPort == 0 {
		cv.SSHPort = 22
	}
	if cv.SSHCredential.User == "" || len(cv.SSHCredential.Password) == 0 {
		return fmt.Errorf("Invalid SSH authencation credential")
	}
	return nil
}
