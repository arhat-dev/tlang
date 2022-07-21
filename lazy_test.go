package tlang

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetLazyValue(t *testing.T) {
	v := &LazyValue[int64]{
		Create: func() int64 { return 1 },
	}

	ret := GetLazyValue(reflect.ValueOf(v))
	t.Log(ret.String())
	assert.Equal(t, int64(1), reflect.Indirect(ret).Int())
}

func TestLazyValue_Get(t *testing.T) {
	const (
		testdata = "test"
	)

	var called int32
	lv := &LazyValue[string]{
		Create: func() string {
			_ = atomic.AddInt32(&called, 1)

			time.Sleep(5 * time.Second)
			return testdata
		},
	}

	startSig := make(chan struct{})

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			<-startSig

			assert.EqualValues(t, testdata, lv.GetLazyValue())
		}()
	}

	close(startSig)
	wg.Wait()
	assert.EqualValues(t, 1, called)
}
