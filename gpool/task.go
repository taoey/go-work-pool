package gpool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Task struct {
	Name string
	Func func(ctx context.Context, data interface{}) error
	Data interface{}

	exit chan struct{}
	wg   *sync.WaitGroup
}

func NewTask(name string, f func(context.Context, interface{}) error, data interface{}) *Task {
	return &Task{
		Name: name,
		Func: f,
		Data: data,
		exit: make(chan struct{}),
	}
}

func (t *Task) run(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println(fmt.Sprintf("recover from %v", r))
		}
	}()

	if err := t.Func(ctx, t.Data); err != nil {
		log.Println(err)
		return err
	}
	t.exit <- struct{}{}
	return nil
}

func (t *Task) Run(ctx context.Context) error {
	return t.run(ctx)
}

func (t *Task) RunWithTimeout(ctx context.Context, timeout time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	t.wg.Add(1)

	go t.run(ctx)

	select {
	case <-ctx.Done():
		t.wg.Done()
		log.Println(fmt.Sprintf("任务超时:%s", t.Name))
		return nil
	case <-t.exit:
		t.wg.Done()
		log.Println(fmt.Sprintf("任务执行完毕:%s", t.Name))
		return nil
	}
}
