// Much of the KNN code has been adapted from the
// github.com/dhconnelly/rtreego project.
//
// Copyright 2012 Daniel Connelly.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package rtree

import (
	"math"
	"sort"
)

// NearestNeighbors gets the closest Spatials to the Point.
func (tr *RTree) NearestNeighbors(k int, x, y, z float64) []Item {
	if tr.tr.root == nil {
		return nil
	}
	dists := make([]float64, k)
	objs := make([]Item, k)
	for i := 0; i < k; i++ {
		dists[i] = math.MaxFloat64
		objs[i] = nil
	}
	objs, _ = tr.nearestNeighbors(k, x, y, z, tr.tr.root, dists, objs)
	//for i := 0; i < len(objs); i++ {
	//	fmt.Printf("%v\n", objs[i])
	//}
	for i := 0; i < len(objs); i++ {
		if objs[i] == nil {
			return objs[:i]
		}
	}
	return objs
}

// minDist computes the square of the distance from a point to a rectangle.
// If the point is contained in the rectangle then the distance is zero.
//
// Implemented per Definition 2 of "Nearest Neighbor Queries" by
// N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.
func minDist(x, y, z float64, r d3rectT) float64 {
	sum := 0.0
	p := [3]float64{x, y, z}
	rp := [3]float64{
		float64(r.min[0]), float64(r.min[1]), float64(r.min[2]),
	}
	rq := [3]float64{
		float64(r.max[0]), float64(r.max[1]), float64(r.max[2]),
	}
	for i := 0; i < 3; i++ {
		if p[i] < float64(rp[i]) {
			d := p[i] - float64(rp[i])
			sum += d * d
		} else if p[i] > float64(rq[i]) {
			d := p[i] - float64(rq[i])
			sum += d * d
		}
	}
	return sum
}

func (tr *RTree) nearestNeighbors(k int, x, y, z float64, n *d3nodeT, dists []float64, nearest []Item) ([]Item, []float64) {
	if n.isLeaf() {
		for i := 0; i < n.count; i++ {
			e := n.branch[i]
			dist := math.Sqrt(minDist(x, y, z, e.rect))
			dists, nearest = insertNearest(k, dists, nearest, dist, e.data.(Item))
		}
	} else {
		branches, branchDists := sortEntries(x, y, z, n.branch[:n.count])
		branches = pruneEntries(x, y, z, branches, branchDists)
		for _, e := range branches {
			nearest, dists = tr.nearestNeighbors(k, x, y, z, e.child, dists, nearest)
		}
	}
	return nearest, dists
}

// insert obj into nearest and return the first k elements in increasing order.
func insertNearest(k int, dists []float64, nearest []Item, dist float64, obj Item) ([]float64, []Item) {
	i := 0
	for i < k && dist >= dists[i] {
		i++
	}
	if i >= k {
		return dists, nearest
	}

	left, right := dists[:i], dists[i:k-1]
	updatedDists := make([]float64, k)
	copy(updatedDists, left)
	updatedDists[i] = dist
	copy(updatedDists[i+1:], right)

	leftObjs, rightObjs := nearest[:i], nearest[i:k-1]
	updatedNearest := make([]Item, k)
	copy(updatedNearest, leftObjs)
	updatedNearest[i] = obj
	copy(updatedNearest[i+1:], rightObjs)

	return updatedDists, updatedNearest
}

type entrySlice struct {
	entries []d3branchT
	dists   []float64
	x, y, z float64
}

func (s entrySlice) Len() int { return len(s.entries) }

func (s entrySlice) Swap(i, j int) {
	s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
	s.dists[i], s.dists[j] = s.dists[j], s.dists[i]
}
func (s entrySlice) Less(i, j int) bool {
	return s.dists[i] < s.dists[j]
}

func sortEntries(x, y, z float64, entries []d3branchT) ([]d3branchT, []float64) {
	sorted := make([]d3branchT, len(entries))
	dists := make([]float64, len(entries))
	for i := 0; i < len(entries); i++ {
		sorted[i] = entries[i]
		dists[i] = minDist(x, y, z, entries[i].rect)
	}
	sort.Sort(entrySlice{sorted, dists, x, y, z})
	return sorted, dists
}

func pruneEntries(x, y, z float64, entries []d3branchT, minDists []float64) []d3branchT {
	minMinMaxDist := math.MaxFloat64
	for i := range entries {
		minMaxDist := minMaxDist(x, y, z, entries[i].rect)
		if minMaxDist < minMinMaxDist {
			minMinMaxDist = minMaxDist
		}
	}
	// remove all entries with minDist > minMinMaxDist
	pruned := []d3branchT{}
	for i := range entries {
		if minDists[i] <= minMinMaxDist {
			pruned = append(pruned, entries[i])
		}
	}
	return pruned
}

// minMaxDist computes the minimum of the maximum distances from p to points
// on r.  If r is the bounding box of some geometric objects, then there is
// at least one object contained in r within minMaxDist(p, r) of p.
//
// Implemented per Definition 4 of "Nearest Neighbor Queries" by
// N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.
func minMaxDist(x, y, z float64, r d3rectT) float64 {

	p := [3]float64{x, y, z}
	rp := [3]float64{
		float64(r.min[0]), float64(r.min[1]), float64(r.min[2]),
	}
	rq := [3]float64{
		float64(r.max[0]), float64(r.max[1]), float64(r.max[2]),
	}

	// by definition, MinMaxDist(p, r) =
	// min{1<=k<=n}(|pk - rmk|^2 + sum{1<=i<=n, i != k}(|pi - rMi|^2))
	// where rmk and rMk are defined as follows:

	rm := func(k int) float64 {
		if p[k] <= (rp[k]+rq[k])/2 {
			return rp[k]
		}
		return rq[k]
	}

	rM := func(k int) float64 {
		if p[k] >= (rp[k]+rq[k])/2 {
			return rp[k]
		}
		return rq[k]
	}

	// This formula can be computed in linear time by precomputing
	// S = sum{1<=i<=n}(|pi - rMi|^2).

	S := 0.0
	for i := range p {
		d := p[i] - rM(i)
		S += d * d
	}

	// Compute MinMaxDist using the precomputed S.
	min := math.MaxFloat64
	for k := range p {
		d1 := p[k] - rM(k)
		d2 := p[k] - rm(k)
		d := S - d1*d1 + d2*d2
		if d < min {
			min = d
		}
	}

	return min
}
