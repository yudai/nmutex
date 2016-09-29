package nmutex

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

var names = 8
var users = names * 100
var maxSleep = 10
var maxTries = 20

func TestNamedMutex(t *testing.T) {
	m := New()

	var wg sync.WaitGroup
	wg.Add(users)

	logs := make(map[string]chan bool)
	for n := 0; n < names; n++ {
		logs[strconv.Itoa(n)] = make(chan bool, maxTries*users)
	}

	for u := 0; u < users; u++ {
		go func(u int) {
			defer wg.Done()

			name := strconv.Itoa(u % names)
			log := logs[name]

			tries := rand.Intn(maxTries)
			for try := 0; try < tries; try++ {
				unlock := m.Lock(name)
				log <- true
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(maxSleep)))
				log <- false
				unlock()
			}
		}(u)
	}

	wg.Wait()

	for _, log := range logs {
		prev := false
		close(log)
		for l := range log {
			if l == prev {
				t.Fatalf("detected broken lock")
			}
			prev = l
		}
	}

	if len(m.sets) != 0 {
		t.Fatalf("detected stale set")
	}
}
