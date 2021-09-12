package localworker_test

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ihexxa/quickshare/src/worker"
	"github.com/ihexxa/quickshare/src/worker/localworker"
)

func TestWorkerPools(t *testing.T) {
	type tinput struct {
		ID int `json:"id"`
	}

	workersTest := func(workers worker.IWorkerPool, t *testing.T) {
		records := &sync.Map{}
		mType1, mType2 := "mtype1", "mtype2"

		handler1 := func(msg worker.IMsg) error {
			input := &tinput{}
			err := json.Unmarshal([]byte(msg.Body()), input)
			if err != nil {
				t.Fatal(err)
			}

			records.Store(mType1, input.ID)
			return nil
		}
		handler2 := func(msg worker.IMsg) error {
			input := &tinput{}
			err := json.Unmarshal([]byte(msg.Body()), input)
			if err != nil {
				t.Fatal(err)
			}

			records.Store(mType2, input.ID)
			return nil
		}

		workers.AddHandler(mType1, handler1)
		workers.AddHandler(mType2, handler2)
		workers.Start()

		count := 3
		for i := 0; i < count; i++ {
			body, _ := json.Marshal(&tinput{ID: i})
			workers.TryPut(localworker.NewMsg(
				uint64(i),
				map[string]string{localworker.MsgTypeKey: mType1},
				string(body),
			))
			workers.TryPut(localworker.NewMsg(
				uint64(i*10),
				map[string]string{localworker.MsgTypeKey: mType2},
				string(body),
			))
		}

		workers.Stop()
		workers.DelHandler(mType1)
		workers.DelHandler(mType2)

		val1, ok := records.Load(mType1)
		if !ok {
			t.Fatal("mtype1 not found")
		}
		count1 := val1.(int)
		if count1 != count-1 {
			t.Fatalf("incorrect count %d", count1)
		}

		val2, ok := records.Load(mType2)
		if !ok {
			t.Fatal("mtype2 not found")
		}
		count2 := val2.(int)
		if count1 != count-1 {
			t.Fatalf("incorrect count %d", count2)
		}
	}

	t.Run("test bolt provider", func(t *testing.T) {
		// rootPath, err := ioutil.TempDir("./", "quickshare_kvstore_test_")
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// defer os.RemoveAll(rootPath)

		stdoutWriter := zapcore.AddSync(os.Stdout)
		multiWriter := zapcore.NewMultiWriteSyncer(stdoutWriter)
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			multiWriter,
			zap.InfoLevel,
		)

		workers := localworker.NewWorkerPool(1024, 1, 2, zap.New(core).Sugar())
		workersTest(workers, t)
	})
}
