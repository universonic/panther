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

package etcd

import (
	"path/filepath"
	"strings"
	"time"

	clientv3 "github.com/coreos/etcd/clientv3"
	namespace "github.com/coreos/etcd/clientv3/namespace"
	transport "github.com/coreos/etcd/pkg/transport"
	generic "github.com/universonic/panther/pkg/storage/generic"
	zap "go.uber.org/zap"
)

const (
	defaultDialTimeout = 2 * time.Second
)

// SSLOptions indicates the SSL options to be used in a etcd v3 connection
type SSLOptions struct {
	Enabled    bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`
	ServerName string `json:"server_name,omitempty" yaml:"server_name,omitempty" toml:"server_name,omitempty"`
	CACert     string `json:"ca_cert,omitempty" yaml:"ca_cert,omitempty" toml:"ca_cert,omitempty"`
	SSLKey     string `json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
	SSLCert    string `json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
}

// Config indicates the etcd configuration
type Config struct {
	Endpoints  []string    `json:"endpoints,omitempty" yaml:"endpoints,omitempty" toml:"endpoints,omitempty"`
	Namespace  []string    `json:"-" yaml:"-" toml:"-"`
	User       string      `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Password   string      `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
	SSLOptions *SSLOptions `json:"ssl,omitempty" yaml:"ssl,omitempty" toml:"ssl,omitempty"`
}

// Open is used for initiating a new connection with etcd cluster.
func (in *Config) Open(logger *zap.SugaredLogger) (generic.Storage, error) {
	cfg := clientv3.Config{
		Endpoints:   in.Endpoints,
		DialTimeout: defaultDialTimeout * time.Second,
		Username:    in.User,
		Password:    in.Password,
	}

	var cfgtls *transport.TLSInfo
	if in.SSLOptions != nil {
		tlsinfo := transport.TLSInfo{}

		if in.SSLOptions.SSLCert != "" {
			tlsinfo.CertFile = in.SSLOptions.SSLCert
			cfgtls = &tlsinfo
		}
		if in.SSLOptions.SSLKey != "" {
			tlsinfo.KeyFile = in.SSLOptions.SSLKey
			cfgtls = &tlsinfo
		}
		if in.SSLOptions.CACert != "" {
			tlsinfo.CAFile = in.SSLOptions.CACert
			cfgtls = &tlsinfo
		}
		if in.SSLOptions.ServerName != "" {
			tlsinfo.ServerName = in.SSLOptions.ServerName
			cfgtls = &tlsinfo
		}

		if cfgtls != nil {
			clientTLS, err := cfgtls.ClientConfig()
			if err != nil {
				return nil, err
			}
			cfg.TLS = clientTLS
		}
	}

	cfg.DialTimeout = 3 * time.Second

	db, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	if len(in.Namespace) == 0 {
		in.Namespace = []string{"/com.redhat", "panther"}
	}
	if !strings.HasPrefix(in.Namespace[0], "/") {
		in.Namespace[0] = "/" + in.Namespace[0]
	}
	prefix := filepath.Join(in.Namespace...)
	db.KV = namespace.NewKV(db.KV, prefix)
	c := &conn{
		prefix: prefix,
		db:     db,
		logger: logger,
	}
	return c, nil
}

// New returns an empty configuration carrier.
func New() *Config {
	return new(Config)
}
