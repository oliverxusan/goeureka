package goeureka

import (
	"fmt"
	"testing"
)

func TestShuffle(t *testing.T) {
	var r = NewRand()
	var cnt1 = map[int]int{}
	for i := 0; i < 1000000; i++ {
		var sl = []int{0, 1, 2, 3, 4, 5, 6}
		r.RandRobin(sl)
		cnt1[sl[0]]++
	}

	var cnt2 = map[int]int{}
	for i := 0; i < 1000000; i++ {
		sl := r.RandRobin2(7)
		cnt2[sl[0]]++
	}

	fmt.Println(cnt1, "\n", cnt2)
}
