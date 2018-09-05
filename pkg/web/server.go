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
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	genericStorage "github.com/universonic/panther/pkg/storage/generic"
	zap "go.uber.org/zap"
)

const (
	// DefaultWriteTimeout is the default timeout duration of writing response.
	DefaultWriteTimeout = time.Second * 15
	// DefaultReadTimeout is the default timeout duration of reading requests.
	DefaultReadTimeout = time.Second * 15
	// DefaultIdleTimeout is the default timeout duration of keep-alive if enabled.
	DefaultIdleTimeout = time.Second * 60
)

// Server is a combined http API server
type Server struct {
	closeWG  sync.WaitGroup
	closeCh  chan struct{}
	clzSubCh []chan struct{}
	grace    time.Duration
	unixAddr string
	tcpAddr  string
	unixSrv  *http.Server // Unix socket is used for local CLI
	tcpSrv   *http.Server
	wwwroot  string
}

// Prepare initialize a new router for inner server
func (in *Server) Prepare(storage genericStorage.Storage, logger *zap.SugaredLogger) {
	router := NewHandler(storage, logger, in.wwwroot)
	in.unixSrv.Handler = router
	in.tcpSrv.Handler = router
}

// Subscribe attach an external channel to be used for callback function when server exited.
func (in *Server) Subscribe() <-chan struct{} {
	subscription := make(chan struct{}, 1)
	in.clzSubCh = append(in.clzSubCh, subscription)
	return subscription
}

// Serve listen and serve on desired address, and returns gracefully if Stop was called.
func (in *Server) Serve() error {
	defer func() {
		for i := range in.clzSubCh {
			close(in.clzSubCh[i])
		}
		in.clzSubCh = in.clzSubCh[:0]
	}()
	unixListener, err := net.Listen("unix", in.unixAddr)
	if err != nil {
		return err
	}
	tcpListener, err := net.Listen("tcp", in.tcpAddr)
	if err != nil {
		return err
	}
	go in.serveUnix(unixListener)
	go in.serveTCP(tcpListener)
	<-in.closeCh

	in.closeWG.Add(2)
	go in.stopUnix()
	go in.stopTCP()
	in.closeWG.Wait()
	return nil
}

// Stop shutdown the server gracefully.
func (in *Server) Stop() {
	close(in.closeCh)
}

func (in *Server) serveUnix(listener net.Listener) error {
	return in.unixSrv.Serve(listener)
}

func (in *Server) stopUnix() error {
	defer in.closeWG.Done()
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), in.grace)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	return in.unixSrv.Shutdown(ctx)
}

func (in *Server) serveTCP(listener net.Listener) error {
	return in.tcpSrv.Serve(listener)
}

func (in *Server) stopTCP() error {
	defer in.closeWG.Done()
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), in.grace)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	return in.tcpSrv.Shutdown(ctx)
}

// NewServer generates a new server with given addresses and grace.
func NewServer(unix, tcp string, grace int, wwwroot string) (*Server, error) {
	us := &http.Server{
		WriteTimeout: DefaultWriteTimeout,
		ReadTimeout:  DefaultReadTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
	ts := &http.Server{
		WriteTimeout: DefaultWriteTimeout,
		ReadTimeout:  DefaultReadTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
	s := &Server{
		closeCh:  make(chan struct{}, 1),
		grace:    time.Duration(grace) * time.Second,
		unixAddr: unix,
		tcpAddr:  tcp,
		unixSrv:  us,
		tcpSrv:   ts,
		wwwroot:  wwwroot,
	}
	return s, nil
}
