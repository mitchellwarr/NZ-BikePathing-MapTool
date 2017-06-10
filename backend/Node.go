package main

type node struct {
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lng"`
	Elevation float64 `json:"elevation"`
	parent    *node
	gcost     float64
	fcost     float64
	hcost     float64
	id        int64
}

func (n *node) calcCost(g float64, h float64) {
	n.gcost = g
	n.hcost = h
	n.fcost = g + h
}

func (n *node) isShorterThan(n2 *node) bool {
	shorter := false
	if n.fcost < n2.fcost {
		shorter = true
	} else if n.fcost == n2.fcost {
		if n.gcost < n2.gcost {
			shorter = true
		}
	}
	return shorter
}

func ContainsNode(list []*node, elem *node) bool {
	for _, t := range list {
		if t.id == elem.id {
			return true
		}
	}
	return false
}
