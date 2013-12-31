package main

import (
	"fmt"
)

func main() {
	a := make([]int, 2)
	for i:= range a {
		a[i] = i
	}
	fmt.Println(a[1:1])
}
