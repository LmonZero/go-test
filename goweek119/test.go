package main

import "fmt"

type integer int

type a struct {
	s string
	c int
}

func (i integer) String() string {

	return "hello"
}

func main() {
	fmt.Println(integer(10))
	fmt.Printf("%T\r\n", integer(10))
	fmt.Printf("%T", a{"1", 10})
}
