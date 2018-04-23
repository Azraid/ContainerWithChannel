// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Data struct {
	num int
}

type DataAction struct {
	do   func(data *Data) Data
	data *Data
}

func Process(in <-chan *DataAction, out chan<- Data) {
	for d := range in {
		result := d.do(d.data)
		out <- result
	}
}

type Event struct {
	action int
	param  interface{}
	do     func(*Container, interface{})
}

type Container struct {
	dataM  map[int]*Data
	eventC chan Event
	closeC chan bool
}

func NewContainer() *Container {
	c := &Container{}
	c.dataM = make(map[int]*Data)
	c.eventC = make(chan Event)
	c.closeC = make(chan bool)
	return c
}

func (c *Container) Process() {
	for {
		select {
		case ev := <-c.eventC:
			fmt.Println("event:", ev.action)
			ev.do(c, ev.param)

		case close := <-c.closeC:
			if close {
				fmt.Println("Process closed")
				return
			}
		}
	}
}

func main() {
	runCount := 100
	rand.Seed(100)
	cont := NewContainer()
	go cont.Process()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < runCount; i++ {
			cont.eventC <- Event{action: 1,
				param: i,
				do: func(c *Container, param interface{}) {
					c.dataM[rand.Intn(10)] = &Data{num: param.(int)}
				},
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < runCount; i++ {
			cont.eventC <- Event{action: 2,
				do: func(c *Container, param interface{}) {
					delete(c.dataM, rand.Intn(10))
				},
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < runCount; i++ {
			cont.eventC <- Event{action: 3,
				do: func(c *Container, param interface{}) {
					if v, ok := c.dataM[rand.Intn(10)]; ok {
						fmt.Println("get", v.num)
					}
				},
			}
		}
	}()

	wg.Wait()
	cont.closeC <- true

	time.Sleep(time.Second * 2)
}
