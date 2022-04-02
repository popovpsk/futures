package future

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestRun_Success(t *testing.T) {
	f := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 10)
		return "test_result", nil
	})

	checkSuccessfullCase(f)
}

func TestRun_Error(t *testing.T) {
	var customError = errors.New("test error")
	f := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 10)
		return "", customError
	})
	_, err := f.GetResult()
	if err == nil {
		log.Panic("err == nil")
	}
	if err != customError {
		log.Panic("err != customError")
	}
}

func TestWait_IsCopyable(t *testing.T) {
	future := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 10)
		return "test_result", nil
	})

	future2 := future

	checkSuccessfullCase(future2)
	checkSuccessfullCase(future)
}

func TestGetResult_MultiConsumer(t *testing.T) {
	f := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 10)
		return "test_result", nil
	})

	const consumersCount = 1000
	wg := sync.WaitGroup{}
	wg.Add(consumersCount)
	for i := 0; i < consumersCount; i++ {
		go func() {
			defer wg.Done()
			checkSuccessfullCase(f)
		}()
	}

	wg.Wait()
}

func TestGetResult_Panic(t *testing.T) {
	future := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 10)
		panic("test panic")
	})

	_, err := future.GetResult()
	if err == nil {
		log.Panic("err == nil")
	}
}

func TestWait(t *testing.T) {
	future := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 20)
		return "test_result", nil
	})

	if future.IsCompleted() {
		log.Panic("future.IsCompleted() == true before finish")
	}
	future.Wait()
	checkSuccessfullCase(future)

	if !future.IsCompleted() {
		log.Panic("future.IsCompleted() == false after wait")
	}

	future.Wait()
	checkSuccessfullCase(future)
}

func TestGetResult_OnCompleted(t *testing.T) {
	future := Run(func() (string, error) {
		time.Sleep(time.Millisecond * 20)
		return "test_result", nil
	})
	checkSuccessfullCase(future)
}

func TestWaitAll(t *testing.T) {
	start := time.Now()

	const count = 100
	futures := make([]Future[int], 100)
	for i := 0; i < count; i++ {
		cp := i
		futures[i] = Run(func() (int, error) {
			time.Sleep(time.Duration((rand.Int() % 100)) * time.Millisecond)
			return cp, nil
		})
	}

	WaitAll(futures...)
	for _, f := range futures {
		if !f.IsCompleted() {
			log.Panic("!f.IsCompleted()")
		}
	}

	if time.Since(start) > time.Second {
		log.Panic("time limit exceeded")
	}
}

func TestWhenAll(t *testing.T) {
	start := time.Now()

	const count = 100
	futures := make([]Future[int], 100)
	for i := 0; i < count; i++ {
		cp := i
		futures[i] = Run(func() (int, error) {
			time.Sleep(time.Duration((rand.Int() % 100)) * time.Millisecond)
			return cp, nil
		})
	}

	results := WhenAll(futures...)
	set := make(map[int]struct{}, count)
	for _, r := range results {
		if _, ok := set[r.Value]; ok {
			log.Panicf("value: %d already in set", r.Value)
		}
		set[r.Value] = struct{}{}
	}

	if time.Since(start) > time.Second {
		log.Panic("time limit exceeded")
	}
}

func checkSuccessfullCase(f Future[string]) {
	res, err := f.GetResult()
	if err != nil {
		log.Panic("err != nil")
	}
	if res != "test_result" {
		log.Panic("res != test_result")
	}
}
