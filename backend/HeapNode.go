package main

// An HeapNode is a min-heap of ints.
type HeapNode []*node

func (h HeapNode) Len() int           { return len(h) }
func (h HeapNode) Less(i, j int) bool { return h[i].fcost < h[j].fcost }
func (h HeapNode) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// Push func
func (h *HeapNode) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*node))
}

// Pop func
func (h *HeapNode) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
