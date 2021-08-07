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

// var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// const rndCap = 10000
// const addCap = 1

// // how to set time
// // extend: wait can be greater than ttl/2
// // cyc is smaller than ttl and wait, then it can be clean in time
// const cap = 40
// const ttl = 3
// const cyc = 1
// const bucketCap = 2
// const id1 = "id1"
// const id2 = "id2"
// const op1 int16 = 0
// const op2 int16 = 1

// var customCaps = map[int16]int16{
// 	op2: 1000,
// }

// const wait = 1

// var limiter = New(cap, cyc, bucketCap)

// var idSeed = 0

// func randId() string {
// 	idSeed++
// 	return fmt.Sprintf("%d", idSeed)
// }

// func TestAccess(t *testing.T) {
// 	func(t *testing.T) {
// 		canAccess := limiter.Access(id1)
// 		if !canAccess {
// 			t.Fatal("access: fail")
// 		}

// 		for i := 0; i < bucketCap; i++ {
// 			canAccess = limiter.Access(id1)
// 		}

// 		if canAccess {
// 			t.Fatal("access: fail to deny access")
// 		}

// 		time.Sleep(time.Duration(limiter.GetCyc()) * time.Second)

// 		canAccess = limiter.Access(id1)
// 		if !canAccess {
// 			t.Fatal("access: fail to refresh tokens")
// 		}
// 	}(t)
// }

// func TestCap(t *testing.T) {
// 	originalCap := limiter.GetCap()
// 	fmt.Printf("cap:info: %d\n", originalCap)

// 	ok := limiter.ExpandCap(originalCap + addCap)

// 	if !ok || limiter.GetCap() != originalCap+addCap {
// 		t.Fatal("cap: fail to expand")
// 	}

// 	ok = limiter.ExpandCap(limiter.GetSize() - addCap)
// 	if ok {
// 		t.Fatal("cap: shrink cap")
// 	}

// 	ids := []string{}
// 	for limiter.GetSize() < limiter.GetCap() {
// 		id := randId()
// 		ids = append(ids, id)

// 		ok := limiter.Access(id, 0)
// 		if !ok {
// 			t.Fatal("cap: not full")
// 		}
// 	}

// 	if limiter.GetSize() != limiter.GetCap() {
// 		t.Fatal("cap: incorrect size")
// 	}

// 	if limiter.Access(randId(), 0) {
// 		t.Fatal("cap: more than cap")
// 	}

// 	limiter.truncate()
// }

// func TestTtl(t *testing.T) {
// 	var addTtl int32 = 1
// 	originalTTL := limiter.GetTTL()
// 	fmt.Printf("ttl:info: %d\n", originalTTL)

// 	limiter.UpdateTTL(originalTTL + addTtl)
// 	if limiter.GetTTL() != originalTTL+addTtl {
// 		t.Fatal("ttl: update fail")
// 	}
// }

// func cycTest(t *testing.T) {
// 	var addCyc int32 = 1
// 	originalCyc := limiter.GetCyc()
// 	fmt.Printf("cyc:info: %d\n", originalCyc)

// 	limiter.UpdateCyc(originalCyc + addCyc)
// 	if limiter.GetCyc() != originalCyc+addCyc {
// 		t.Fatal("cyc: update fail")
// 	}
// }

// func autoCleanTest(t *testing.T) {
// 	ids := []string{
// 		randId(),
// 		randId(),
// 	}

// 	for _, id := range ids {
// 		ok := limiter.Access(id, 0)
// 		if ok {
// 			t.Fatal("autoClean: warning: add fail")
// 		}
// 	}

// 	time.Sleep(time.Duration(limiter.GetTTL()+wait) * time.Second)

// 	for _, id := range ids {
// 		_, exist := limiter.get(id)
// 		if exist {
// 			t.Fatal("autoClean: item still exist")
// 		}
// 	}
// }

// func snapshotTest(t *testing.T) {
// }
