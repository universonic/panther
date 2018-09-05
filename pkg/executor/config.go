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

package executor

// Config is the configuration of executor
type Config struct {
	Schedule string `json:"schedule,omitempty" yaml:"schedule,omitempty" toml:"schedule,omitempty"`
	Workers  int    `json:"workers,omitempty" yaml:"workers,omitempty" toml:"workers,omitempty"`
}

// Complete fulfills the empty fields of Config
func (in *Config) Complete() {
	if in.Schedule == "" {
		in.Schedule = "@daily"
	}
	if in.Workers <= 2 {
		in.Workers = 8
	}
}

// Apply spawns a new API server with configuration, and returns any encountered error.
func (in *Config) Apply() (*Server, error) {
	return NewServer(in.Schedule, in.Workers)
}
