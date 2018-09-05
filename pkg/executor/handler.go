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

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	genericStorage "github.com/universonic/panther/pkg/storage/generic"
	sshutil "github.com/universonic/panther/pkg/utils/ssh"
	zap "go.uber.org/zap"
)

// Handler is indeed an external executer caller.
type Handler struct {
	lock    sync.RWMutex
	total   int
	busy    int
	storage genericStorage.Storage
	logger  *zap.SugaredLogger
	wg      sync.WaitGroup
	queue   chan workload
	clzChan chan struct{}
}

func (in *Handler) worker() {
	in.lock.Lock()
	in.total++
	in.lock.Unlock()
	for job := range in.queue {
		in.lock.Lock()
		in.busy++
		in.lock.Unlock()
		if job.gc {
			switch job.val.GetKind() {
			case genericStorage.RESOURCE_HOST:
				in.gcHost(job.val.(*genericStorage.Host))
			}
		} else {
			switch job.val.GetKind() {
			case genericStorage.RESOURCE_HOST:
				in.handleHost(job.val.(*genericStorage.Host))
			case genericStorage.RESOURCE_SYSTEM_SCAN:
				in.handleScan(job.val.(*genericStorage.SystemScan))
			case genericStorage.RESOURCE_HOST_OPERATION:
				in.handleOp(job.val.(*genericStorage.HostOperation))
			}
		}
		in.lock.Lock()
		in.busy--
		in.lock.Unlock()
	}
	in.wg.Done()
	in.lock.Lock()
	in.total--
	in.lock.Unlock()
}

func (in *Handler) sendJob(obj genericStorage.Object, gc bool) {
	defer recover() // prevent panic while trying to send on a closed handler
	in.queue <- workload{
		gc:  gc,
		val: obj,
	}
}

func (in *Handler) reportWorkerState() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			in.logger.Debugw("Workers state =>", "busy", in.busy, "total", in.total)
			in.logger.Sync()
		case <-in.clzChan:
			return
		}
	}
}

// State returns the number of workers in busy state and the number of total workers.
func (in *Handler) State() (busy, total int) {
	in.lock.RLock()
	defer in.lock.RUnlock()
	return in.busy, in.total
}

// ScanAllHost scans on all existing host.
func (in *Handler) ScanAllHost() {
	defer in.logger.Sync()
	list := genericStorage.NewHostList()
	err := in.storage.List(list)
	if err != nil {
		in.logger.Errorf("Could not retrieve host list from storage due to: %v", err)
		return
	}
	for i := range list.Members {
		in.sendJob(&list.Members[i], false)
	}
}

// HandleHostEvent handles host event.
func (in *Handler) HandleHostEvent(event genericStorage.WatchEvent) {
	defer in.logger.Sync()
	host := genericStorage.NewHost()
	err := event.Unmarshal(host)
	if err != nil {
		in.logger.Errorf("Received host event seems to be invalid: %v", err)
		return
	}
	in.sendJob(host, false)
}

// HandleScanEvent handles system scan event
func (in *Handler) HandleScanEvent(event genericStorage.WatchEvent) {
	defer in.logger.Sync()
	scan := genericStorage.NewSystemScan()
	err := event.Unmarshal(scan)
	if err != nil {
		in.logger.Errorf("Received system scan event seems to be invalid: %v", err)
		return
	}
	in.sendJob(scan, false)
}

// HandleOpEvent executes command on host, which is driven by event.
func (in *Handler) HandleOpEvent(event genericStorage.WatchEvent) {
	defer in.logger.Sync()
	op := genericStorage.NewHostOperation()
	err := event.Unmarshal(op)
	if err != nil {
		in.logger.Errorf("Received host operation event seems to be invalid: %v", err)
		return
	}
	in.sendJob(op, false)
}

// HandleHostCleanupEvent cleanup resources that is assigned with the event host.
func (in *Handler) HandleHostCleanupEvent(event genericStorage.WatchEvent) {
	defer in.logger.Sync()
	host := genericStorage.NewHost()
	err := event.Unmarshal(host)
	if err != nil {
		in.logger.Errorf("Received host event seems to be invalid: %v", err)
		return
	}
	in.sendJob(host, true)
}

// handleHost scans updates on single host and stores result into database.
func (in *Handler) handleHost(host *genericStorage.Host) {
	defer in.logger.Sync()

	scan := genericStorage.NewSystemScan()
	scan.SetName(host.GetName())
	err := in.storage.Get(scan)
	if err != nil {
		if genericStorage.IsInternalError(err) {
			in.logger.Errorf("Unexpected storage error while trying to scan host '%s': %v", host.GetName(), err)
			return
		}
		scan.State = genericStorage.StartedState
		err = in.storage.Create(scan)
		if err != nil {
			in.logger.Errorf("Could not initiate host scan result: %v", err)
			return
		}
	} else {
		switch scan.State {
		case genericStorage.SuccessState, genericStorage.FailureState:
			scan.State = genericStorage.StartedState
			err = in.storage.Update(scan)
			if err != nil {
				in.logger.Errorf("Abort to scan host '%s' since we could not initiate a scan result due to: %v", host.GetName(), err)
				return
			}
		}
	}
}

func (in *Handler) handleScan(scan *genericStorage.SystemScan) {
	defer in.logger.Sync()

	if scan.State != genericStorage.StartedState {
		return
	}
	scan.State = genericStorage.InProgressState
	scan.Security = scan.Security[:0]
	err := in.storage.Update(scan)
	if err != nil {
		in.logger.Errorf("Abort to scan host '%s' due to: %v", scan.GetName(), err)
		return
	}

	var (
		lines    [][]byte
		observer genericStorage.Watcher
		upstream <-chan genericStorage.WatchEvent
		done     chan error
	)
	op := genericStorage.NewHostOperation()

	host := genericStorage.NewHost()
	host.SetName(scan.GetName())
	err = in.storage.Get(host)
	if err != nil {
		in.logger.Errorf("Abort to scan host '%s' due to: %v", scan.GetName(), err)
		scan.State = genericStorage.FailureState
		goto FINALIZE
	}
	op.SetGUID(uuid.NewV4().String())
	op.SetName(op.GetGUID())
	op.SetNamespace(host.GetName())
	op.Type = genericStorage.InternalOperation
	op.Command = "yum updateinfo list cve"
	op.Method = genericStorage.OutputMethod
	op.State = genericStorage.StartedState

	observer, err = in.storage.Watch(op, genericStorage.WatchOnName)
	if err != nil {
		in.logger.Errorf("Abort to scan host '%s' since we could not initiate observer due to: %v", host.GetName(), err)
		scan.State = genericStorage.FailureState
		goto FINALIZE
	}
	defer observer.Close()
	upstream = observer.Output()
	done = make(chan error, 1)

	go func() {
		for {
			event := <-upstream
			switch event.Type {
			case genericStorage.CREATE:
				continue
			case genericStorage.DELETE:
				done <- fmt.Errorf("Target has been deleted before we can proceed")
				return
			case genericStorage.ERROR:
				done <- fmt.Errorf("%s", event.Value)
				return
			}
			cv := genericStorage.NewHostOperation()
			err := event.Unmarshal(cv)
			if err != nil {
				done <- err
				return
			}
			switch cv.State {
			case genericStorage.SuccessState:
				op = cv
				done <- nil
				return
			case genericStorage.FailureState:
				done <- fmt.Errorf("%s", cv.Data)
				return
			}
		}
	}()

	err = in.storage.Create(op)
	if err != nil {
		in.logger.Errorf("Could not initiate operation for host '%s' due to: %v", host.GetName(), err)
		scan.State = genericStorage.FailureState
		goto FINALIZE
	}
	err = <-done
	close(done)
	if err != nil {
		in.logger.Errorf("Failed to scan on host '%s' due to: %v", host.GetName(), err)
		scan.State = genericStorage.FailureState
		goto FINALIZE
	}
	lines = bytes.Split(op.Data, []byte("\n"))
	for _, each := range lines {
		line := bytes.TrimSpace(each)
		if !bytes.HasPrefix(line, []byte("CVE-")) {
			continue
		}
		parsed := bytes.Fields(line)
		if len(parsed) == 3 {
			entity := genericStorage.SecurityUpdate{
				CVEID:   string(parsed[0]),
				Package: string(parsed[2]),
			}
			switch string(parsed[1]) {
			case "Critical/Sec.":
				entity.Severity = genericStorage.CriticalSec
			case "Important/Sec.":
				entity.Severity = genericStorage.ImportantSec
			case "Moderate/Sec.":
				entity.Severity = genericStorage.ModerateSec
			}
			scan.Security = append(scan.Security, entity)
		}
	}
	scan.State = genericStorage.SuccessState

FINALIZE:
	err = in.storage.Update(scan)
	if err != nil {
		in.logger.Errorf("Could not save scan result for host '%s' due to: %v", host.GetName(), err)
	}
}

func (in *Handler) handleOp(op *genericStorage.HostOperation) {
	defer in.logger.Sync()
	// Ignore those ops that is being handled by other workers.
	if op.State != genericStorage.StartedState {
		return
	}
	var (
		host *genericStorage.Host
		err  error
		conn *sshutil.Conn
		dAtA []byte
	)
	done := make(chan struct{}, 1)
	defer close(done)
	if op.GetNamespace() == "" {
		in.logger.Warnf("Abort to perform command `%s` due to no assigned host.", op.Command)
		op.State = genericStorage.AbortState
		goto FINALIZE
	}
	if op.Type == genericStorage.UnknownOperation {
		in.logger.Warnf("Abort to perform command `%s` on host '%s' due to no assigned type.", op.Command, op.GetNamespace())
		op.State = genericStorage.AbortState
		goto FINALIZE
	}
	if op.Method == genericStorage.UnknownMethod {
		in.logger.Warnf("Abort to perform command `%s` on host '%s' due to no assigned method.", op.Command, op.GetNamespace())
		op.State = genericStorage.AbortState
		goto FINALIZE
	}

	host = genericStorage.NewHost()
	host.SetName(op.GetNamespace())
	err = in.storage.Get(host)
	if err != nil {
		in.logger.Errorf("Could not perform command `%s` on host '%s' due to a storage error: %v", op.Command, op.GetNamespace(), err)
		op.State = genericStorage.AbortState
		goto FINALIZE
	}
	op.State = genericStorage.InProgressState
	err = in.storage.Update(op)
	if err != nil {
		in.logger.Errorf("Could not refresh operation state for host '%s' due to a storage error: %v", op.GetNamespace(), err)
		op.State = genericStorage.FailureState
		op.Data = []byte(err.Error())
		goto FINALIZE
	}
	conn, err = sshutil.NewConn(host.SSHAddress, host.SSHPort, host.SSHCredential.User, string(host.SSHCredential.Password))
	if err != nil {
		in.logger.Errorf("Failed to connect to host '%s' due to: %v", host.GetName(), err)
		op.State = genericStorage.FailureState
		op.Data = []byte(err.Error())
		goto FINALIZE
	}
	defer conn.Close()
	if host.OpCredential.User != "" {
		err = conn.Su(host.OpCredential.User, string(host.OpCredential.Password))
		if err != nil {
			in.logger.Errorf("Unexpected privilege error on host %s: %v", host.GetName(), err)
			op.State = genericStorage.FailureState
			op.Data = []byte(err.Error())
			goto FINALIZE
		}
	}

	switch op.Method {
	case genericStorage.RunMethod:
		err = conn.Run(op.Command, op.Timeout)
	case genericStorage.OutputMethod:
		dAtA, err = conn.Output(op.Command, op.Timeout)
	case genericStorage.CombinedOutputMethod:
		dAtA, err = conn.CombinedOutput(op.Command, op.Timeout)
	}
	if err != nil {
		in.logger.Errorf("Failed to perform command `%s` on host '%s' due to: %v", op.Command, host.GetName(), err)
		op.State = genericStorage.FailureState
		op.Data = []byte(err.Error())
		goto FINALIZE
	}
	op.Data = dAtA
	op.State = genericStorage.SuccessState

FINALIZE:
	err = in.storage.Update(op)
	if err != nil {
		in.logger.Errorf("Could not store execution result to database due to: %v", err)
	}
}

func (in *Handler) gcHost(host *genericStorage.Host) {
	defer in.logger.Sync()

	scan := genericStorage.NewSystemScan()
	scan.SetName(host.GetName())
	err := in.storage.Delete(scan)
	if err != nil {
		if genericStorage.IsInternalError(err) {
			in.logger.Errorf("Could cleanup system scanning result that is related to host '%s' due to an internal error: %v", host.GetName(), err)
		} else {
			in.logger.Warnf("System scanning result that is related to host '%s' seems has been removed from the storage: %v", host.GetName(), err)
		}
	}
}

// Close aims to shutdown handler gracefully if possible.
func (in *Handler) Close() error {
	close(in.queue)
	close(in.clzChan)

	clz := make(chan struct{}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		in.wg.Wait()
		close(clz)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-clz:
		return nil
	}
}

// NewHandler return a new Handler instance.
func NewHandler(storage genericStorage.Storage, logger *zap.SugaredLogger, workers int) *Handler {
	h := &Handler{
		storage: storage,
		logger:  logger,
		queue:   make(chan workload, 100),
	}
	h.wg.Add(workers)
	for w := 0; w < workers; w++ {
		go h.worker()
	}
	go h.reportWorkerState()
	return h
}

type workload struct {
	gc  bool
	val genericStorage.Object
}
