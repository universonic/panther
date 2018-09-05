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
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	mux "github.com/gorilla/mux"
	websocket "github.com/gorilla/websocket"
	genericStorage "github.com/universonic/panther/pkg/storage/generic"
	zap "go.uber.org/zap"
)

// Handler is a bundled API handler
type Handler struct {
	*mux.Router
	wwwroot string
	storage genericStorage.Storage
	logger  *zap.SugaredLogger
	ws      websocket.Upgrader
}

// Frontend handles GET requests from /*
func (in *Handler) Frontend(w http.ResponseWriter, r *http.Request) {
	defer in.finalizeHeader(w)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if ext := filepath.Ext(r.URL.Path); ext == "" {
		http.ServeFile(w, r, in.wwwroot)
		return
	}
	http.FileServer(http.Dir(in.wwwroot)).ServeHTTP(w, r)
}

func (in *Handler) finalizeStorageError(w http.ResponseWriter, e error) {
	if genericStorage.IsNotFound(e) {
		in.finalizeError(w, e, http.StatusNotFound)
	}
	if genericStorage.IsConflict(e) {
		in.finalizeError(w, e, http.StatusConflict)
	}
}

func (in *Handler) finalizeBody(w http.ResponseWriter, r io.Reader, status int) {
	w.WriteHeader(status)
	length, err := io.Copy(w, r)
	if err != nil {
		// Panic here due to there is no possible way to recover.
		panic(err)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", length))
}

func (in *Handler) finalizeJSON(w http.ResponseWriter, r io.Reader, status ...int) {
	var stat int
	if len(status) > 0 {
		stat = status[0]
	} else {
		stat = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	in.finalizeBody(w, r, stat)
}

func (in *Handler) finalizeError(w http.ResponseWriter, e error, status int) {
	in.finalizeBody(w, bytes.NewReader([]byte(e.Error()+"\n")), status)
	w.Header().Set("Content-Type", "plain/text; charset=utf-8")
}

func (in *Handler) finalizeNull(w http.ResponseWriter) {
	defer in.finalizeHeader(w)
	w.WriteHeader(http.StatusNoContent)
}

func (in *Handler) finalizeHeader(w http.ResponseWriter) {
	w.Header().Set("Date", time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST"))
	w.Header().Set("Server", "CMDB/v0.1-alpha")
}

func (in *Handler) auditIntercepter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer in.logger.Sync()
		// The incoming request might be forwarded from a proxy server.
		ips := r.Header.Get("X-Forwarded-For")
		if ips == "" {
			ips = r.RemoteAddr
		}
		ip := strings.TrimSpace(strings.Split(ips, ",")[0])
		if ip == "" {
			ip = "local"
		}
		in.logger.Infow("Request =>",
			"user_agent", r.Header.Get("User-Agent"),
			"method", r.Method,
			"url", r.URL.Path,
			"from", ip,
		)
		next.ServeHTTP(w, r)
	})
}

// NewHandler returns a new initialized HTTP handler
func NewHandler(storage genericStorage.Storage, logger *zap.SugaredLogger, wwwroot string) *Handler {
	root := mux.NewRouter()
	h := &Handler{
		Router:  root,
		wwwroot: wwwroot,
		storage: storage,
		logger:  logger,
		ws: websocket.Upgrader{
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	root.Use(h.auditIntercepter)

	apiRoot := root.PathPrefix("/api/v1").Subrouter()
	apiRoot.HandleFunc("/host", h.Host)
	apiRoot.HandleFunc("/exec", h.Exec)

	root.PathPrefix("/").HandlerFunc(h.Frontend)
	return h
}
