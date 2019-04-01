package main

import (
    "fmt"
)

type circularBuffer struct{
	size int
	head int
	tail int
	data []*httpRequestData
}

func (c *circularBuffer) init(size int) {
	c.size = size
	c.data = make([]*httpRequestData, size)
	c.head = 0
	c.tail = 0
}
func (c *circularBuffer) empty() bool{
	return (c.head == c.tail && c.data[c.head] == nil)
}

func (c *circularBuffer) insert(data *httpRequestData) bool{
	if c.data[c.head] != nil {
		fmt.Printf("the buffer is full... slow your role")
		return false
	}
	c.data[c.head] = data
	c.head = (c.head + 1)%c.size
	fmt.Printf("%+v\n",c)
	return true
}

func (c *circularBuffer) get() *httpRequestData{
	if c.data[c.tail] == nil{
		return nil
	}
	return c.data[c.tail]
}

func (c *circularBuffer) pop(){
	if c.data[c.tail] == nil{
		fmt.Println("you are trying to delete from an empty buffer")
		return
	}
	c.data[c.tail] = nil
	c.tail = (c.tail + 1)%c.size
	fmt.Printf("%+v\n",c)
}

func (c *circularBuffer) addRequest() int {
	return 1
}



