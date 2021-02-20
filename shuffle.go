package goeureka

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Shuffle interface {
	RandRobin(indexes []int)
	RandRobin2(n int) []int
}

type Rand struct {
}

func NewRand() *Rand {
	return new(Rand)
}
func (r *Rand) RandRobin(indexes []int) {
	for i := len(indexes); i > 0; i-- {
		lastIdx := i - 1
		idx := rand.Intn(i)
		indexes[lastIdx], indexes[idx] = indexes[idx], indexes[lastIdx]
	}
}

func (r *Rand) RandRobin2(n int) []int {
	return rand.Perm(n)
}
