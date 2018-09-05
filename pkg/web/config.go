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

import "fmt"

const (
	// DefaultShutdownTimeout is a default time duration (in seconds) to wait for a graceful
	// server shutdown.
	DefaultShutdownTimeout = 10
)

// Config is the configuration of API server
type Config struct {
	Socket  string `json:"socket,omitempty" yaml:"socket,omitempty" toml:"socket,omitempty"`
	Address string `json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	Port    int    `json:"port,omitempty" yaml:"port,omitempty" toml:"port,omitempty"`
	WWWRoot string `json:"www_root,omitempty" yaml:"www_root,omitempty" toml:"www_root,omitempty"`
}

// Complete fulfills the empty fields of Config
func (in *Config) Complete() {
	if in.Socket == "" {
		in.Socket = "/var/run/panther.sock"
	}
	if in.Address == "" {
		in.Address = "0.0.0.0"
	}
	if in.Port == 0 {
		in.Port = 8080
	}
	if in.WWWRoot == "" {
		in.WWWRoot = "/usr/share/panther/wwwroot"
	}
}

// Apply spawns a new API server with configuration, and returns any encountered error.
func (in *Config) Apply() (*Server, error) {
	return NewServer(in.Socket, fmt.Sprintf("%s:%d", in.Address, in.Port), DefaultShutdownTimeout, in.WWWRoot)
}
