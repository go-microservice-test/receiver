package main

import "fmt"

func main() {
	var nums [3]int = [...]int{0: 3, 2: 2, 1: 3}
	for _, number := range nums {
		fmt.Println(number)
	}

}
