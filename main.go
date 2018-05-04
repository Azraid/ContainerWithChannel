package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Data struct {
	num   int
	exter int
}

type ChannelMessage struct {
	param  interface{}
	result chan interface{}
	do     func(*Container, interface{}, chan<- interface{})
}

type Container struct {
	dataM  map[int]*Data
	reqC   chan ChannelMessage
	closeC chan bool
}

func NewContainer() *Container {
	c := &Container{}
	c.dataM = make(map[int]*Data)
	c.reqC = make(chan ChannelMessage)
	c.closeC = make(chan bool)
	return c
}

func (c *Container) Go() {
	for {
		select {
		case req := <-c.reqC:
			req.do(c, req.param, req.result)

		case close := <-c.closeC:
			if close {
				fmt.Println("Process closed")
				return
			}
		}
	}
}

func (c *Container) Add(data *Data) {
	c.reqC <- ChannelMessage{
		param: data,
		do: func(c *Container, param interface{}, result chan<- interface{}) {
			d := param.(*Data)
			c.dataM[d.num] = d
		},
	}
}

func (c *Container) Remove(num int) {
	c.reqC <- ChannelMessage{
		param: num,
		do: func(c *Container, param interface{}, result chan<- interface{}) {
			num := param.(int)
			delete(c.dataM, num)
		},
	}
}

func (c *Container) Get(num int) (*Data, bool) {
	result := make(chan interface{})

	c.reqC <- ChannelMessage{
		param:  num,
		result: result,
		do: func(c *Container, param interface{}, result chan<- interface{}) {
			result <- c.dataM[param.(int)]
		},
	}

	data := <-result
	return data.(*Data), (data.(*Data) != nil)
}

func (c *Container) Do(num int, do func(*Data) error) error {
	result := make(chan interface{})

	c.reqC <- ChannelMessage{
		param:  num,
		result: result,
		do: func(c *Container, param interface{}, result chan<- interface{}) {
			var err error
			if v, ok := c.dataM[param.(int)]; ok {
				err = do(v)
			}
			result <- err
		},
	}

	e := <-result
	if e == nil {
		return nil
	} else {
		return e.(error)
	}
}

func main() {

	rand.Seed(100)
	cont := NewContainer()
	go cont.Go()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		cont.Add(&Data{num: 1})
		cont.Add(&Data{num: 2})
		cont.Remove(2)
		cont.Add(&Data{num: 3})
	}()

	time.Sleep(1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		cont.Do(1, func(d *Data) error {
			d.exter = 3333
			return nil
		})

		data, ok := cont.Get(1)
		if ok {
			fmt.Println(data.num, data.exter)
			data.num++
		} else {
			fmt.Println("no data")
		}
	}()

	wg.Wait()
	cont.closeC <- true

	time.Sleep(time.Second * 2)
}
