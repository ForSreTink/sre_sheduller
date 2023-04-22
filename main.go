package main

import (
	"fmt"
	"time"
)

func main() {
	t := "05:44"
	t_n, err := time.Parse("15:04", t)
	fmt.Println(t_n)
	fmt.Println(err)
}
