package main

import (
	"github.com/vmware/govmomi/vim25/progress"
	"sync"
)

// copied from the real progress logger for vsphere-images commands
// which in turn is copied from the govc progress logger
//
// this one is modified to not print anything
// the Sinker interface is surprisingly complicated to implement correctly

type progressLogger struct {
	wg sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func newProgressLogger() *progressLogger {
	p := &progressLogger{
		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(ch)
			if err != nil {
				stop = true
			}
		case <-p.done:
			stop = true
		}
	}
}

func (p *progressLogger) loopB(ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		}
	}

	return err
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
}

func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}
