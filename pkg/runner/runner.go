package runner

import (
	"sync"

	"github.com/QuestsInc/errs"
)

type Runner interface {
	// Run function must wait till Stop called or error happened
	// in case if it is not wait runner will stop execution
	// of all other runners immediately
	Run() error
	Stop() error
}

type numberedError struct {
	n   int
	err error
}

type runner struct {
	runners  []Runner
	stopChan chan struct{}
	run      bool
	mu       sync.Mutex
}

func New(runners ...Runner) *runner {
	return &runner{
		runners: runners,
		run:     false,
		mu:      sync.Mutex{},
	}
}

func (r *runner) Run() error {
	errChan := make(chan numberedError, len(r.runners))
	r.stopChan = make(chan struct{})

	for i := range r.runners {
		i := i
		go func() {
			err := r.runners[i].Run()
			errChan <- numberedError{
				n:   i,
				err: err,
			}
		}()
	}

	r.mu.Lock()
	r.run = true
	r.mu.Unlock()

	select {
	case e := <-errChan:
		// Error happened and r.Stop() must have no effect
		r.mu.Lock()
		r.run = false
		r.mu.Unlock()

		err := r.stopRunners(e.n)
		return errs.Combine(e.err, err)
	case _ = <-r.stopChan:
		n := len(r.runners)
		errors := make([]error, n)

		for i := 0; i < n; i++ {
			e := <-errChan
			errors = append(errors, e.err)
		}

		return errs.Combine(errors...)
	}
}

func (r *runner) Stop() error {
	r.mu.Lock()
	run := r.run
	r.mu.Unlock()

	if !run {
		return nil
	}

	r.stopChan <- struct{}{}
	return r.stopRunners(-1)
}

func (r *runner) stopRunners(n int) error {
	var errors []error
	for i := 0; i < len(r.runners); i++ {
		if i == n {
			continue
		}

		err := r.runners[i].Stop()
		errors = append(errors, err)
	}

	return errs.Combine(errors...)
}
