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

package storage

import (
	etcd "github.com/universonic/panther/pkg/storage/etcd"
	generic "github.com/universonic/panther/pkg/storage/generic"
	zap "go.uber.org/zap"
)

// Config is a generic type of storage configuration
type Config interface {
	Open(logger *zap.SugaredLogger) (generic.Storage, error)
}

// ConfigInitiator is used for initialize a storage config
type ConfigInitiator struct {
	Adapter string `json:"adapter,omitempty" yaml:"adapter,omitempty" toml:"adapter,omitempty"`
}

// NewConfigInitiator returns an empty set of ConfigInitiator
func NewConfigInitiator() *ConfigInitiator {
	return new(ConfigInitiator)
}

// QualifiedConfig is a fulfilled top layer storage configuration.
type QualifiedConfig struct {
	Adapter      string `json:"adapter,omitempty" yaml:"adapter,omitempty" toml:"adapter,omitempty"`
	*etcd.Config `json:"config,omitempty" yaml:"config,omitempty" toml:"config,omitempty"`
}

// NewQualifiedConfig returns a new QualifiedConfig as config carrier by given adapter name, and any encountered error if any.
func NewQualifiedConfig(of string) (*QualifiedConfig, error) {
	switch of {
	case "etcd":
		return &QualifiedConfig{
			Adapter: of,
			Config:  etcd.New(),
		}, nil
	}
	return nil, ErrUnknownStorageAdapter
}
