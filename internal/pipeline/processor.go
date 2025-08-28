package pipeline

import "sync"

type Job struct{ Run func() error }

type Processor struct {
	wg   sync.WaitGroup
	jobs chan Job
}

func NewProcessor(workers, buf int) *Processor {
	p := &Processor{jobs: make(chan Job, buf)}
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for j := range p.jobs {
				_ = j.Run()
			}
		}()
	}
	return p
}

func (p *Processor) Submit(j Job) { p.jobs <- j }
func (p *Processor) Close()       { close(p.jobs); p.wg.Wait() }
