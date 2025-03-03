// Package chitian/errgroup is an extension of the sync/errgroup which has this header comment:
//
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package errgroup

import (
        "context"
        "errors"
        "runtime"
        "sync"

        "code.byted.org/gopkg/logs"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
        doneOnce sync.Once
        done     chan struct{}

        wg      sync.WaitGroup
        errOnce sync.Once
        err     error

        ch chan struct{}
}

// default cancel on error.
func WithCancel() *Group {
        return &Group{done: make(chan struct{})}
}

// Not Recommended, just to be compatibility with package "golang.org/x/sync/errgroup"
func WithContext(ctx context.Context) (*Group, context.Context) {
        return &Group{done: make(chan struct{})}, ctx
}

// set max goroutine to work.
func (g *Group) SetLimit(n int) {
        g.ch = make(chan struct{}, n)
}

// close channel only once
func (g *Group) cancel() {
        if g.done == nil {
                return
        }
        g.doneOnce.Do(func() {
                close(g.done)
        })
}

func (g *Group) done_() {
        g.wg.Done()
        if g.ch != nil {
                <-g.ch
        }
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
        g.wg.Wait()
        g.cancel()
        if g.ch != nil {
                close(g.ch)
        }
        return g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *Group) Go(f func() error) {
        if g.ch != nil {
                select {
                case <-g.done:
                        return
                case g.ch <- struct{}{}:
                        break
                }
        }

        g.wg.Add(1)
        go func() {
                defer g.done_()
                localDone := make(chan error, 2)
                go func() {
                        defer func() {
                                if err := recover(); err != nil {
                                        const size = 64 << 10
                                        buf := make([]byte, size)
                                        buf = buf[:runtime.Stack(buf, false)]
                                        logs.Error("panic: err=%s\n%s",
                                                logs.SecMark("error", err),
                                                logs.SecMark("stack", string(buf)))
                                        localDone <- errors.New("panic")
                                }
                                close(localDone)
                        }()
                        if err := f(); err != nil {
                                localDone <- err
                        }
                }()

                select {
                case err := <-localDone:
                        if err != nil {
                                g.errOnce.Do(func() {
                                        g.err = err
                                })
                                g.cancel()
                        }
                case <-g.done: // when g.done is nil, this case will be blocked
                }
        }()
}

