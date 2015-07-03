package main

import (
	"fmt"

	"github.com/mikkeloscar/calbox/golfbox"
)

func main() {

	times := golfbox.GetTimes("14-1644", "2428")
	for _, t := range times {
		fmt.Println(t.Club)
		fmt.Println(t.Time)
		for _, p := range t.Players {
			fmt.Printf("%#v\n", p)
		}
	}
}
