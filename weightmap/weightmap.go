package weightmap

import (
	"bytes"
	"fmt"
)

// weightMap is used to create a map of integes that can be returned
// according to a weighted chance.
// Typically used with retrieving randomized items where some of the values
// should hav a better chance than others to be returned.
type weightMap [][2]int

func NewWeightMap() *weightMap {
	wm := new(weightMap)
	return wm
}
func (c *weightMap) Add(weight int, value int) *weightMap {
	// *c = append(*c, [2]int{c.getmax() + weight, value})
	*c = append(*c, [2]int{weight, value})
	return c
}

func (c weightMap) getmax() int {
	var max int
	length := len(c)
	for i := 0; i < length; i++ {
		max += (c)[i][0]
	}
	return max
}
func (c weightMap) String() string {
	s := bytes.Buffer{}
	max := float64(c.getmax()) / 100
	for _, v := range c {
		c := float64(v[0])
		fmt.Fprintf(&s, "%d: %05.2f%% (%v/%v)\n", v[1], c/max, v[0], max)
	}
	return s.String()
}
func (c weightMap) GetDistribution() map[int]float64 {
	dist := map[int]float64{}
	max := float64(c.getmax())
	for _, v := range c {
		c := float64(v[0])
		dist[v[1]] = c / max
	}
	return dist
}

func (c weightMap) Get(nonce int) int {
	length := len(c)
	if nonce == 0 {
		nonce = 1
	}
	h := evenlyDistributedHash(nonce)
	max := c.getmax()
	r := h % max
	if length == 0 {
		return -1
	}
	prev := 0
	for i := 0; i < length; i++ {
		if (c[i][0] + prev) > r {
			return c[i][1]
		}
		prev += c[i][0]
	}
	return -2
}

// based on https://stackoverflow.com/a/12996028
func evenlyDistributedHash(x int) int {
	x = ((x >> 16) ^ x) * 0x45d9f3b
	x = ((x >> 16) ^ x) * 0x45d9f3b
	x = (x >> 16) ^ x
	return x
}
