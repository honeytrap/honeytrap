package rtree

// Item is an rtree item
type Item interface {
	Rect() (minX, minY, minZ, maxX, maxY, maxZ float64)
}

// Rect is a rectangle
type Rect struct {
	MinX, MinY, MinZ, MaxX, MaxY, MaxZ float64
}

// Rect returns the rectangle
func (item *Rect) Rect() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	return item.MinX, item.MinY, item.MinZ, item.MaxX, item.MaxY, item.MaxZ
}

// RTree is an implementation of an rtree
type RTree struct {
	tr *d3RTree
}

// New creates a new RTree
func New() *RTree {
	return &RTree{
		tr: d3New(),
	}
}

// Insert inserts item into rtree
func (tr *RTree) Insert(item Item) {
	minX, minY, minZ, maxX, maxY, maxZ := item.Rect()
	tr.tr.Insert([3]float64{minX, minY, minZ}, [3]float64{maxX, maxY, maxZ}, item)
}

// Remove removes item from rtree
func (tr *RTree) Remove(item Item) {
	minX, minY, minZ, maxX, maxY, maxZ := item.Rect()
	tr.tr.Remove([3]float64{minX, minY, minZ}, [3]float64{maxX, maxY, maxZ}, item)
}

// Search finds all items in bounding box.
func (tr *RTree) Search(minX, minY, minZ, maxX, maxY, maxZ float64, iterator func(item Item) bool) {
	tr.tr.Search([3]float64{minX, minY, minZ}, [3]float64{maxX, maxY, maxZ}, func(data interface{}) bool {
		return iterator(data.(Item))
	})
}

// Count return the number of items in rtree.
func (tr *RTree) Count() int {
	return tr.tr.Count()
}

// RemoveAll removes all items from rtree.
func (tr *RTree) RemoveAll() {
	tr.tr.RemoveAll()
}

func (tr *RTree) Bounds() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	var rect d3rectT
	if tr.tr.root != nil {
		if tr.tr.root.count > 0 {
			rect = tr.tr.root.branch[0].rect
			for i := 1; i < tr.tr.root.count; i++ {
				rect2 := tr.tr.root.branch[i].rect
				rect = d3combineRect(&rect, &rect2)
			}
		}
	}
	minX, minY, minZ = float64(rect.min[0]), float64(rect.min[1]), float64(rect.min[2])
	maxX, maxY, maxZ = float64(rect.max[0]), float64(rect.max[1]), float64(rect.max[2])
	return
}
