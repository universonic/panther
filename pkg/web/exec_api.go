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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	websocket "github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	genericStorage "github.com/universonic/panther/pkg/storage/generic"
)

// Exec handles requests from /api/v1/exec, dispatches a new session for long-polling requests
// via websocket.
// Usage:
//   - GET /api/v1/exec?mode=scan&watch=[HOST_LIST|*]
//   - GET /api/v1/exec?mode=cmd
// Mode:
//   - scan: Retrieve and watch scanning data on given hosts. Users are able to send requests
//           to enforce a rescan operation on specified hosts.
//   - cmd:  Send qualified command on specified hosts, and retrieves execution result one by one.
//           If the previous request was not accomplished, it will not accept the next one until
//           current process has finished.
func (in *Handler) Exec(w http.ResponseWriter, r *http.Request) {
	defer in.logger.Sync()

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mode := r.URL.Query().Get("mode")
	watch := strings.Split(r.URL.Query().Get("watch"), ",")

	conn, err := in.ws.Upgrade(w, r, nil)
	if err != nil {
		in.logger.Errorf("Could not dispatch a new websocket session due to: %v", err)
		return
	}
	defer conn.Close()
	if mode == "" {
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Invalid mode."),
			time.Time{}, // Exit right now.
		)
		return
	}
	if mode != "scan" && mode != "cmd" {
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "No such channel."),
			time.Time{}, // Exit right now.
		)
		return
	}
	if mode == "scan" && len(watch) == 1 && watch[0] == "" {
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Target required."),
			time.Time{}, // Exit right now.
		)
		return
	}

	in.logger.Infow("Websocket session initiated with params: ", "mode", mode, "watch", watch)

	defer func() {
		defer in.logger.Sync()
		if r := recover(); r != nil {
			in.logger.Errorf("An unexpected error occurred: %v", r)
			// Still try to tell client we are about to exit.
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Internal Server Error"),
				time.Time{}, // Exit right now.
			)
		}
	}()

	cache := NewCache()

	switch mode {
	case "scan":
		clzChan := make(chan struct{})

		var (
			result    []*genericStorage.SystemScan
			observer  genericStorage.Watcher
			eventChan <-chan genericStorage.WatchEvent
		)

		observer, err = in.storage.Watch(genericStorage.NewSystemScan(), genericStorage.WatchOnKind)
		if err != nil {
			in.logger.Errorf("Could not watch on system updates due to: %v", err)
			panic(err)
		}
		defer observer.Close()
		eventChan = observer.Output()

		if watch[0] == "*" {
			cache.Loose = true
			list := genericStorage.NewSystemScanList()
			err = in.storage.List(list)
			if err != nil {
				in.logger.Errorf("Unexpected storage error: %v", err)
				panic(err)
			}
			for i := range list.Members {
				scan := &list.Members[i]
				cache.Set(scan.GetName(), scan)
				result = append(result, scan)
			}
		} else {
			for _, name := range watch {
				scan := genericStorage.NewSystemScan()
				scan.SetName(name)
				err = in.storage.Get(scan)
				if err != nil {
					if genericStorage.IsInternalError(err) {
						in.logger.Errorf("Unexpected storage error: %v", err)
						panic(err)
					}
					continue
				}
				cache.Set(scan.GetName(), scan)
				result = append(result, scan)
			}
		}
		err = conn.WriteJSON(result)
		if err != nil {
			in.logger.Errorf("Failed to send result: %v", err)
			panic(err)
		}
		result = result[:0]

		go func() {
			defer in.logger.Sync()
			for event := range eventChan {
				cv := genericStorage.NewSystemScan()
				err = event.Unmarshal(cv)
				if err != nil {
					in.logger.Errorf("Could not unmarshal incoming event due to: %v", err)
					panic(err)
				}
				if !cache.Check(cv.GetName()) {
					continue
				}
				switch event.Type {
				case genericStorage.CREATE, genericStorage.UPDATE:
					cache.Set(cv.GetName(), cv)
				case genericStorage.DELETE:
					cache.Pop(cv.GetName())
				}

				all := cache.Flush()
				for i := range all {
					result = append(result, all[i].(*genericStorage.SystemScan))
				}
				err = conn.WriteJSON(result)
				if err != nil {
					in.logger.Errorf("Failed to send result due to: %v", err)
					panic(err)
				}
				result = result[:0]
			}
			in.logger.Debugf("System scanning event observer exited.")
		}()

		go func() {
			defer in.logger.Sync()
			for {
				mt, r, err := conn.NextReader()
				if err != nil {
					if err != io.EOF {
						in.logger.Errorf("Could not read from the message due to: %v", err)
					}
					close(clzChan)
					break
				}
				if mt == websocket.BinaryMessage {
					in.logger.Error("Ignored unsupported binary message.")
					continue
				}

				order := new(WSOrderRequest)
				dec := json.NewDecoder(r)
				err = dec.Decode(order)
				if err != nil {
					in.logger.Errorf("Ignored invalid order message due to: %v", err)
					continue
				}

				for i := range order.Commands {
					v := cache.Get(order.Commands[i].Target)
					if v == nil {
						continue
					}
					scan := v.(*genericStorage.SystemScan)
					switch scan.State {
					case genericStorage.SuccessState, genericStorage.FailureState:
						scan.State = genericStorage.StartedState
						err = in.storage.Update(scan)
						if err != nil {
							in.logger.Errorf("Could not start scanning host '%s' due to: %v", err)
						}
					}
				}
			}
			in.logger.Debugf("System scanning websocket sub-handler exited.")
		}()

		// Server will never exit unless client is closed first.
		<-clzChan

	case "cmd":
		var (
			observer  genericStorage.Watcher
			eventChan <-chan genericStorage.WatchEvent
		)

		mt, r, err := conn.NextReader()
		if err != nil {
			if err != io.EOF {
				in.logger.Errorf("Could not read from the message due to: %v", err)
			}
			return
		}
		if mt == websocket.BinaryMessage {
			in.logger.Error("Ignored unsupported binary message.")
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Unsupported binary data"),
				time.Time{}, // Exit right now.
			)
			return
		}
		order := new(WSOrderRequest)
		dec := json.NewDecoder(r)
		err = dec.Decode(order)
		if err != nil {
			in.logger.Errorf("Ignored invalid order message due to: %v", err)
			panic(err)
		}

		commands := make(map[string]*WSCommand)
		for i := range order.Commands {
			host := genericStorage.NewHost()
			host.SetName(order.Commands[i].Target)
			err = in.storage.Get(host)
			if err != nil {
				if genericStorage.IsInternalError(err) {
					in.logger.Errorf("Unexpected storage error: %v", err)
					panic(err)
				}
				in.logger.Errorf("Abort request due to: %v", err)
				continue
			}
			commands[host.GetName()] = &order.Commands[i]
		}

		observer, err = in.storage.Watch(genericStorage.NewHostOperation(), genericStorage.WatchOnKind)
		if err != nil {
			in.logger.Errorf("Could not watch on system updates due to: %v", err)
			panic(err)
		}
		defer observer.Close()
		eventChan = observer.Output()

		for name, cmd := range commands {
			op := genericStorage.NewHostOperation()
			op.SetGUID(uuid.NewV4().String())
			op.SetName(op.GetGUID())
			op.SetNamespace(name)
			op.Type = genericStorage.UserOperation
			op.Command = cmd.Command
			op.Method = genericStorage.CombinedOutputMethod
			op.State = genericStorage.StartedState
			err = in.storage.Create(op)
			if err != nil {
				in.logger.Errorf("Unexpected storage error: %v", err)
				panic(err)
			}
			cache.Set(op.GetName(), op)
		}

		for finished := 0; finished >= len(commands); {
			event := <-eventChan
			cv := genericStorage.NewHostOperation()
			err = event.Unmarshal(cv)
			if err != nil {
				in.logger.Errorf("Could not unmarshal incoming event due to: %v", err)
				panic(err)
			}
			if cache.Get(cv.GetName()) == nil {
				continue
			}
			cache.Set(cv.GetName(), cv)
			err = conn.WriteJSON(cv)
			if err != nil {
				in.logger.Errorf("Failed to send result: %v", err)
				panic(err)
			}
			switch cv.State {
			case genericStorage.SuccessState, genericStorage.FailureState:
				finished++
			}
		}

		// Anyway, we still tell the client that we are exiting. This is preserved for
		// extra stability assurance.
		if len(commands) != len(order.Commands) {
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Order is partial performed, which may caused by invalid data."),
				time.Time{}, // Exit right now.
			)
			return
		}
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Completed."),
			time.Time{}, // Exit right now.
		)
	}

}

// WSOrderRequest is the inner request of websocket. Once the server accepts an order, it
// will accept no more order until the order has been completely performed. It is the client
// side's responsibility to issue a sane order.
type WSOrderRequest struct {
	// The group of commands, their value relies on Enforce.
	Commands []WSCommand `json:"commands,omitempty"`
}

// WSCommand is the atomic order on single host.
type WSCommand struct {
	// The command to perform. If Scan is also present, this will be ignored.
	// This has no effect if EnforceOnScan was specified.
	Command string `json:"command,omitempty"`
	// The target hosts. If a target is present during scanning, then we will rescan the presented
	// target.
	Target string `json:"target,omitempty"`
}

// Cache is a fast cacher for websocket.
type Cache struct {
	Loose bool
	lock  sync.RWMutex
	buf   map[string]interface{}
}

// Check returns true if a key is present in cache.
func (in *Cache) Check(key string) bool {
	if in.Loose {
		return true
	}
	in.lock.RLock()
	defer in.lock.RUnlock()
	_, ok := in.buf[key]
	return ok
}

// Set set value for a given key
func (in *Cache) Set(key string, value interface{}) {
	in.lock.Lock()
	defer in.lock.Unlock()
	in.buf[key] = value
}

// Reset cleanup the entire storage.
func (in *Cache) Reset() {
	in.lock.Lock()
	defer in.lock.Unlock()
	in.buf = make(map[string]interface{})
}

// Get retrieve value of a given key, returns nil if not found.
func (in *Cache) Get(key string) interface{} {
	in.lock.RLock()
	defer in.lock.RUnlock()
	if v, ok := in.buf[key]; ok {
		return v
	}
	return nil
}

// Pop removes a key from cache if it exists, and returns its value.
func (in *Cache) Pop(key string) interface{} {
	in.lock.Lock()
	defer in.lock.Unlock()
	if v, ok := in.buf[key]; ok {
		delete(in.buf, key)
		return v
	}
	return nil
}

// Len returns the buffer length
func (in *Cache) Len() int {
	return len(in.buf)
}

// Flush returns all stored data.
func (in *Cache) Flush() []interface{} {
	in.lock.RLock()
	defer in.lock.RUnlock()
	var list []interface{}
	for _, each := range in.buf {
		list = append(list, each)
	}
	return list
}

// NewCache initiates a new cache.
func NewCache() *Cache {
	return &Cache{buf: make(map[string]interface{})}
}
