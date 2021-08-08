package golimiter

import (
	"fmt"
	// "math/rand"
	"sync"
	"testing"
	// "time"
)

func TestLimiter(t *testing.T) {
	t.Run("access count is limited", func(t *testing.T) {
		clientCount := 3
		tokenMaxCount := 3
		wg := sync.WaitGroup{}
		counts := make([]int, clientCount)

		limiter := New(clientCount, 2000)
		client := func(id int) {
			lid := fmt.Sprint(id)
			for i := 0; i < tokenMaxCount*2; i++ {
				ok := limiter.Access(lid, tokenMaxCount, 1)
				if ok {
					counts[id]++
				}
			}

			wg.Done()
		}

		for i := 0; i < clientCount; i++ {
			wg.Add(1)
			go client(i)
		}

		wg.Wait()

		for id := range counts {
			if counts[id] != tokenMaxCount {
				t.Fatalf("id(%d): accessed(%d) tokenMaxCount(%d) don't match", id, counts[id], tokenMaxCount)
			}
		}
	})
}
