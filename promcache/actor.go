package promcache

type Actor struct {
	inbox chan func()
}

func (a *Actor) Run() {
	a.inbox = make(chan func(), 30)
	go func() {
		for fn := range a.inbox {
			fn()
		}
	}()
}

func (a *Actor) Stop() {
	close(a.inbox)
}

func (a *Actor) Tell(f func()) {
	a.inbox <- f
}

func (a *Actor) Ask(f func()) {
	done := make(chan bool)
	a.inbox <- func() {
		f()
		close(done)
	}
	<-done
}
