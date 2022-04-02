package future

import (
	"fmt"
	"sync"
)

type Future[T any] interface {
	GetResult() (T, error)
	IsCompleted() bool
	Wait()
}

type future[T any] struct {
	Result[T]
	result    *T
	err       error
	completed bool
	cv        *sync.Cond
}

type Result[T any] struct {
	Value T
	Err   error
}

func Run[T any](cb func() (T, error)) Future[T] {
	f := new(future[T])
	f.cv = sync.NewCond(&sync.Mutex{})

	go func(f *future[T]) {
		result, err := do(cb)

		f.cv.L.Lock()
		f.result = &result
		f.err = err
		f.completed = true
		f.cv.Broadcast()
		f.cv.L.Unlock()
	}(f)

	return f
}

func (f *future[T]) GetResult() (T, error) {
	f.Wait()
	return *f.result, f.err
}

func (f *future[T]) IsCompleted() bool {
	f.cv.L.Lock()
	result := f.completed
	f.cv.L.Unlock()
	return result
}

func (f *future[T]) Wait() {
	f.cv.L.Lock()
	for !f.completed {
		f.cv.Wait()
	}
	f.cv.L.Unlock()
}

func WaitAll[T any](futures ...Future[T]) {
	for i := range futures {
		futures[i].Wait()
	}
}

func WhenAll[T any](futures ...Future[T]) []Result[T] {
	result := make([]Result[T], len(futures))
	for i := range futures {
		v, err := futures[i].GetResult()
		result[i] = Result[T]{Value: v, Err: err}
	}
	return result
}

func do[T any](cb func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occured: %v", r)
		}
	}()

	result, err = cb()
	return
}
