package index

import (
	"math"

	"github.com/tidwall/tile38/index/rtree"
)

// Item represents an index item.
type Item interface {
	Point() (x, y, z float64)
	Rect() (minX, minY, minZ, maxX, maxY, maxZ float64)
}

// FlexItem can represent a point or a rectangle
type FlexItem struct {
	MinX, MinY, MinZ, MaxX, MaxY, MaxZ float64
}

// Rect returns the rectangle
func (item *FlexItem) Rect() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	return item.MinX, item.MinY, item.MinZ, item.MaxX, item.MaxY, item.MaxZ
}

// Point returns the point
func (item *FlexItem) Point() (x, y, z float64) {
	return item.MinX, item.MinY, item.MinZ
}

// Index is a geospatial index
type Index struct {
	r    *rtree.RTree
	nr   map[*rtree.Rect]Item   // normalized points
	nrr  map[Item][]*rtree.Rect // normalized points
	mulm map[Item]bool          // store items that contain multiple rects
}

// New create a new index
func New() *Index {
	return &Index{
		r:    rtree.New(),
		mulm: make(map[Item]bool),
		nr:   make(map[*rtree.Rect]Item),
		nrr:  make(map[Item][]*rtree.Rect),
	}
}

// Insert inserts an item into the index
func (ix *Index) Insert(item Item) {
	minX, minY, minZ, maxX, maxY, maxZ := item.Rect()
	if minX == maxX && minY == maxY {
		x, y, normd := normPoint(minY, minX)
		if normd {
			nitem := &rtree.Rect{MinX: x, MinY: y, MinZ: minZ, MaxX: x, MaxY: y, MaxZ: maxZ}
			ix.nr[nitem] = item
			ix.nrr[item] = []*rtree.Rect{nitem}
			ix.r.Insert(nitem)
		} else {
			ix.r.Insert(item)
		}
	} else {
		mins, maxs, normd := normRect(minY, minX, maxY, maxX)
		if normd {
			var nitems []*rtree.Rect
			for i := range mins {
				minX, minY, maxX, maxY := mins[i][0], mins[i][1], maxs[i][0], maxs[i][1]
				nitem := &rtree.Rect{MinX: minX, MinY: minY, MinZ: minZ, MaxX: maxX, MaxY: maxY, MaxZ: maxZ}
				ix.nr[nitem] = item
				nitems = append(nitems, nitem)
				ix.r.Insert(nitem)
			}
			ix.nrr[item] = nitems
			if len(mins) > 1 {
				ix.mulm[item] = true
			}
		} else {
			ix.r.Insert(item)
		}
	}
	return
}

// Remove removed an item from the index
func (ix *Index) Remove(item Item) {
	if nitems, ok := ix.nrr[item]; ok {
		for _, nitem := range nitems {
			ix.r.Remove(nitem)
			delete(ix.nr, nitem)
		}
		delete(ix.nrr, item)
	} else {
		ix.r.Remove(item)
	}
}

// Count counts all items in the index.
func (ix *Index) Count() int {
	count := 0
	ix.Search(0, -90, -180, 90, 180, math.Inf(-1), math.Inf(+1), func(item Item) bool {
		count++
		return true
	})
	return count
}

// Bounds returns the minimum bounding rectangle of all items in the index.
func (ix *Index) Bounds() (MinX, MinY, MinZ, MaxX, MaxY, MaxZ float64) {
	return ix.r.Bounds()
}

// RemoveAll removes all items from the index.
func (ix *Index) RemoveAll() {
	ix.r.RemoveAll()
}

func (ix *Index) getRTreeItem(item rtree.Item) Item {
	switch item := item.(type) {
	case Item:
		return item
	case *rtree.Rect:
		return ix.nr[item]
	}
	return nil
}

func (ix *Index) NearestNeighbors(k int, lat, lon float64, iterator func(item Item) bool) {
	x, y, _ := normPoint(lat, lon)
	items := ix.r.NearestNeighbors(k, x, y, 0)
	for _, item := range items {
		iitm := ix.getRTreeItem(item)
		if item == nil {
			continue
		}
		if !iterator(iitm) {
			break
		}
	}
}

// Search returns all items that intersect the bounding box.
func (ix *Index) Search(cursor uint64, swLat, swLon, neLat, neLon, minZ, maxZ float64, iterator func(item Item) bool) (ncursor uint64) {
	var idx uint64
	var active = true
	var idm = make(map[Item]bool)
	mins, maxs, _ := normRect(swLat, swLon, neLat, neLon)
	// Points
	if len(mins) == 1 {
		// There is only one rectangle.
		// It's possible that a r rect may span multiple entries. Check mulm map for spanning rects.
		if active {
			ix.r.Search(mins[0][0], mins[0][1], minZ, maxs[0][0], maxs[0][1], maxZ, func(item rtree.Item) bool {
				if idx >= cursor {
					iitm := ix.getRTreeItem(item)
					if iitm != nil {
						if ix.mulm[iitm] {
							if !idm[iitm] {
								idm[iitm] = true
								active = iterator(iitm)
							}
						} else {
							active = iterator(iitm)
						}
					}
				}
				idx++
				return active
			})
		}
	} else {
		// There are multiple rectangles. Duplicates might occur.
		for i := range mins {
			if active {
				ix.r.Search(mins[i][0], mins[i][1], minZ, maxs[i][0], maxs[i][1], maxZ, func(item rtree.Item) bool {
					if idx >= cursor {
						iitm := ix.getRTreeItem(item)
						if iitm != nil {
							if ix.mulm[iitm] {
								if !idm[iitm] {
									idm[iitm] = true
									active = iterator(iitm)
								}
							} else {
								active = iterator(iitm)
							}
						}
					}
					idx++
					return active
				})
			}
		}
	}
	return idx
}
