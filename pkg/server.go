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

package server

import (
	"os"
	"os/signal"
	"syscall"

	executor "github.com/universonic/panther/pkg/executor"
	genericStorage "github.com/universonic/panther/pkg/storage/generic"
	web "github.com/universonic/panther/pkg/web"
	zap "go.uber.org/zap"
)

// Server indicates the main server process
type Server struct {
	storage  genericStorage.Storage
	closeCh  chan error
	logger   *zap.SugaredLogger
	web      *web.Server
	executor *executor.Server
}

// Run start the server as a main process.
func (in *Server) Run() error {
	defer in.logger.Sync()
	defer in.logger.Info("Server exited")

	executorCltz := in.executor.Subscribe()
	apiClz := in.web.Subscribe()

	in.logger.Info("Server is starting...")
	go func() {
		err := in.executor.Serve()
		if err != nil {
			in.closeCh <- err
		}
	}()
	go func() {
		err := in.web.Serve()
		if err != nil {
			in.closeCh <- err
		}
	}()
	in.logger.Info("Server is ready.")
	in.logger.Sync()

	clz := make(chan os.Signal, 1)
	defer close(clz)
	// We accept graceful shutdowns when quit via SIGINT (Ctrl+C) and SIGTERM (Ctrl+/).
	// SIGKILL or SIGQUIT will not be caught.
	signal.Notify(clz, syscall.SIGINT, syscall.SIGTERM)

LOOP:
	for {
		select {
		case <-clz:
			break LOOP
		case e := <-in.closeCh:
			if e != nil {
				in.logger.Error(e)
			} else {
				in.logger.Fatal("Subprocess exited before server is terminated!")
			}
			break LOOP
		}
	}

	in.web.Stop()
	in.executor.Stop()
	<-apiClz
	<-executorCltz
	return nil
}
