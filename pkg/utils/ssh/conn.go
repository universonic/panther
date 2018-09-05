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

package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/user"
	"sync"
	"time"

	ssh "golang.org/x/crypto/ssh"
)

const (
	// idleTimeoutIntervalPoll is the duration threshold for waiting input hint on interactive commands like `su`.
	// For Run(), Output(), and CombinedOutput(), this feature is supported for internal commands only, which means
	// you cannot perform such a command like `Run("su - test")` directly.
	idleTimeoutIntervalPoll = 3 * time.Second
)

var (
	// ErrOSExecTimeout represents an error that timeout duration are reached while waiting for command to return.
	// This might cause by many complicated reasons on remote system, such as overflow of `open_file_limit` threshold.
	ErrOSExecTimeout = errors.New("Timeout duration exceeded while calling command")
	// ErrInvalidCredential represents an error that an invalid pair of username and password has been given from user.
	ErrInvalidCredential = errors.New("Invalid username or password")
)

type sess struct {
	*ssh.Session
	errChan    chan error
	waitChan   chan error
	lock       sync.Mutex
	timeout    uint
	Stdin      io.WriteCloser
	Stdout     io.Reader
	Stderr     io.Reader
	opUsername string
	opPassword string
}

func (in *sess) Start(cmd string) (err error) {
	in.lock.Lock() // lock this session to prevent conflict
	if in.opUsername == "" {
		err = in.Session.Start(cmd)
	} else {
		err = in.Session.Start(fmt.Sprintf(`su -l %s -c "%s"`, in.opUsername, cmd))
	}
	if err != nil {
		return err
	}
	go in.waitBg()
	return nil
}
func (in *sess) Wait() error {
	return <-in.waitChan
}

func (in *sess) Run(cmd string) error {
	err := in.Start(cmd)
	if err != nil {
		return err
	}
	return in.Wait()
}

func (in *sess) Output(cmd string) ([]byte, error) {
	err := in.Start(cmd)
	if err != nil {
		return nil, err
	}
	err = in.Wait()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(in.Stdout)
}

func (in *sess) CombinedOutput(cmd string) ([]byte, error) {
	err := in.Start(cmd)
	if err != nil {
		return nil, err
	}
	err = in.Wait()
	if err != nil {
		return nil, err
	}
	combined := io.MultiReader(in.Stdout, in.Stderr)
	return ioutil.ReadAll(combined)
}

func (in *sess) Close() error {
	err := in.Session.Close()
	in.Stdin.Close()
	close(in.waitChan)
	close(in.errChan)
	return err
}

func (in *sess) runBg(errChan chan<- error) {
	err := in.Session.Wait()
	if in.opUsername != "" && err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 1:
				errChan <- fmt.Errorf("%v which might due to: %v", e, ErrInvalidCredential)
				return
			}
		}
	}
	errChan <- err
}

func (in *sess) waitBg() {
	defer in.lock.Unlock()
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	done := make(chan error, 1)

	if in.opUsername != "" {
		go in.enterPasswordBg()
		err := <-in.errChan
		if err != nil {
			in.waitChan <- err
			return
		}
	}

	if in.timeout == 0 {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(in.timeout)*time.Second)
	}
	defer cancel()

	go in.runBg(done)
	select {
	case err := <-done:
		in.waitChan <- err
	case <-ctx.Done():
		// FIXME: This is currently a workaround for this, as the signal does not work on OpenSSL properly.
		in.waitChan <- ctx.Err()
		in.Session.Signal(ssh.SIGKILL)
	}
}

func (in *sess) enterPasswordBg() {
	defer recover()
	ctx, cancel := context.WithTimeout(context.Background(), idleTimeoutIntervalPoll)
	defer cancel()
	var (
		prev []byte
		err  error
	)
	errChan := make(chan error, 1)
	defer close(errChan)
	buf := make([]byte, 9)
	go func() {
		for {
			prev = buf
			n, err := in.Stdout.Read(buf)
			if n == 9 && bytes.Contains(append(prev, buf...), []byte("Password:")) {
				_, err := in.Stdin.Write([]byte(in.opPassword + "\n"))
				errChan <- err
				return
			} else if err != nil && err != io.EOF {
				errChan <- err
				return
			}
		}
	}()
	select {
	case err = <-errChan:
		in.errChan <- err
		break
	case <-ctx.Done():
		in.errChan <- ctx.Err()
		break
	}
}

func newSess(se *ssh.Session, opUser, opPass string, timeout uint) (*sess, error) {
	s := &sess{
		Session:    se,
		errChan:    make(chan error, 1),
		waitChan:   make(chan error, 1),
		timeout:    timeout,
		opUsername: opUser,
		opPassword: opPass,
	}
	var err error
	s.Stdin, err = s.StdinPipe()
	if err != nil {
		return nil, err
	}
	s.Stdout, err = s.StdoutPipe()
	if err != nil {
		return nil, err
	}
	s.Stderr, err = s.StderrPipe()
	if err != nil {
		return nil, err
	}
	err = s.Setenv("LANG", "C")
	if err != nil {
		return nil, err
	}
	// Set up terminal modes
	// Documentation:
	// - https://www.ietf.org/rfc/rfc4254.txt
	// - https://godoc.org/golang.org/x/crypto/ssh
	err = s.RequestPty("vt100", 200, 300, ssh.TerminalModes{
		ssh.ECHO:  0, // Disable echoing for general purpose.
		ssh.IGNCR: 1, // Ignore CR on input for cross-platform capability.
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Conn indicates an SSH session that can be reused.
type Conn struct {
	lock       sync.Mutex
	session    *sess
	client     *ssh.Client
	gcChan     chan struct{}
	clzChan    chan struct{}
	errChan    chan error
	done       chan struct{}
	opUsername string
	opPassword string
}

func (in *Conn) gcBg() {
LOOP:
	for {
		select {
		case <-in.clzChan:
			break LOOP
		case <-in.gcChan:
			in.session.Close()
			in.session = nil
		}
	}

	in.session.Close()
	close(in.gcChan)
	close(in.errChan)
	close(in.done)
}

func (in *Conn) async() {
	defer in.lock.Unlock()
	e := in.session.Wait()
	in.errChan <- e
}

// Start runs a command at background, returns any encountered error.
func (in *Conn) Start(cmd string, timeout ...uint) (err error) {
	in.lock.Lock()
	s, err := in.client.NewSession()
	if err != nil {
		return err
	}
	var t uint
	if len(timeout) > 0 {
		t = timeout[0]
	}
	in.session, err = newSess(s, in.opUsername, in.opPassword, t)
	if err != nil {
		return err
	}
	go in.async()
	return in.session.Start(cmd)
}

// Su redirects you to another use with `su` command, and returns any encountered error.
func (in *Conn) Su(username, password string, timeout ...uint) (err error) {
	in.lock.Lock()
	defer in.lock.Unlock()
	s, err := in.client.NewSession()
	if err != nil {
		return err
	}
	var t uint
	if len(timeout) > 0 {
		t = timeout[0]
	}
	in.session, err = newSess(s, in.opUsername, in.opPassword, t)
	if err != nil {
		return err
	}
	err = in.session.Run("stty -a")
	if err != nil {
		return err
	}
	in.opUsername = username
	in.opPassword = password
	return nil
}

// Wait standby until the running command exits. Keep in mind this should not be called if you are using
// Run, Output, or CombinedOutput.
func (in *Conn) Wait() error {
	return <-in.errChan
}

// Run executes a command and returns its result as an error.
func (in *Conn) Run(cmd string, timeout ...uint) error {
	in.lock.Lock()
	defer in.lock.Unlock()
	s, err := in.client.NewSession()
	if err != nil {
		return err
	}
	var t uint
	if len(timeout) > 0 {
		t = timeout[0]
	}
	in.session, err = newSess(s, in.opUsername, in.opPassword, t)
	if err != nil {
		return err
	}
	return in.session.Run(cmd)
}

// Output executes a command at foreground, returns stdout and any encountered error.
func (in *Conn) Output(cmd string, timeout ...uint) ([]byte, error) {
	in.lock.Lock()
	defer in.lock.Unlock()
	s, err := in.client.NewSession()
	if err != nil {
		return nil, err
	}
	var t uint
	if len(timeout) > 0 {
		t = timeout[0]
	}
	in.session, err = newSess(s, in.opUsername, in.opPassword, t)
	if err != nil {
		return nil, err
	}
	return in.session.Output(cmd)
}

// CombinedOutput executes a command at foreground, returns a combined output and any encountered error.
func (in *Conn) CombinedOutput(cmd string, timeout ...uint) ([]byte, error) {
	in.lock.Lock()
	defer in.lock.Unlock()
	s, err := in.client.NewSession()
	if err != nil {
		return nil, err
	}
	var t uint
	if len(timeout) > 0 {
		t = timeout[0]
	}
	in.session, err = newSess(s, in.opUsername, in.opPassword, t)
	if err != nil {
		return nil, err
	}
	return in.session.CombinedOutput(cmd)
}

// Close shutdown the current ssh client and returns any error that encountered.
func (in *Conn) Close() error {
	close(in.clzChan)
	<-in.done
	return in.client.Close()
}

// NewConn returns a new ssh client with given params, returns error if encountered.
func NewConn(host string, port uint16, username, password string) (*Conn, error) {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 22
	}
	if username == "" {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = u.Username
	}
	sshCfg := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	sshCfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshCfg)
	if err != nil {
		return nil, err
	}
	conn := &Conn{
		client:  client,
		gcChan:  make(chan struct{}, 1),
		clzChan: make(chan struct{}, 1),
		errChan: make(chan error, 1),
		done:    make(chan struct{}, 1),
	}
	go conn.gcBg()
	return conn, nil
}
