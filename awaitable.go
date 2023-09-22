package awaitable

import (
	"sync"
	"time"
)

type Awaitable struct {
	tasks       []string
	taskChanged chan bool
	mutex       sync.RWMutex
}

func (j *Awaitable) Add(label string) {
	j.mutex.Lock()
	j.tasks = append(j.tasks, label)
	j.mutex.Unlock()
	select {
	case j.taskChanged <- true:
	default:
	}
}

func (j *Awaitable) Remove(label string) {
	j.mutex.Lock()
	for i, v := range j.tasks {
		if v == label {
			j.tasks = append(j.tasks[:i], j.tasks[i+1:]...)
			break
		}
	}
	j.mutex.Unlock()
	select {
	case j.taskChanged <- true:
	default:
	}
}

func (j *Awaitable) Snapshot() []string {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	return j.tasks
}

func (j *Awaitable) Done(timeout time.Duration) []string {
	completeChan := make(chan bool, 1)

	go func() {
		for {
			j.mutex.RLock()
			if len(j.tasks) == 0 {
				j.mutex.RUnlock()
				completeChan <- true
				close(completeChan)
				return
			}
			j.mutex.RUnlock()
			<-j.taskChanged
		}
	}()

	if timeout > 0 {
		select {
		case <-completeChan:
			return nil
		case <-time.After(timeout):
			return j.Snapshot()
		}
	} else {
		<-completeChan
		return nil
	}
}

func NewAwaitable(jobs ...string) *Awaitable {
	j := &Awaitable{
		tasks:       jobs,
		taskChanged: make(chan bool),
	}
	return j
}
