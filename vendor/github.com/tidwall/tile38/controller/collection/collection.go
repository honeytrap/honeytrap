package collection

import (
	"math"

	"github.com/tidwall/btree"
	"github.com/tidwall/tile38/geojson"
	"github.com/tidwall/tile38/index"
)

const (
	idOrdered    = 0
	valueOrdered = 1
)

type itemT struct {
	id     string
	object geojson.Object
}

func (i *itemT) Less(item btree.Item, ctx interface{}) bool {
	switch ctx {
	default:
		return false
	case idOrdered:
		return i.id < item.(*itemT).id
	case valueOrdered:
		i1, i2 := i.object.String(), item.(*itemT).object.String()
		if i1 < i2 {
			return true
		}
		if i1 > i2 {
			return false
		}
		// the values match so we will compare the ids, which are always unique.
		return i.id < item.(*itemT).id
	}
}

func (i *itemT) Rect() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	bbox := i.object.CalculatedBBox()
	return bbox.Min.X, bbox.Min.Y, bbox.Min.Z, bbox.Max.X, bbox.Max.Y, bbox.Max.Z
}

func (i *itemT) Point() (x, y, z float64) {
	x, y, z, _, _, _ = i.Rect()
	return
}

// Collection represents a collection of geojson objects.
type Collection struct {
	items       *btree.BTree // items sorted by keys
	values      *btree.BTree // items sorted by value+key
	index       *index.Index // items geospatially indexed
	fieldMap    map[string]int
	fieldValues map[string][]float64
	weight      int
	points      int
	objects     int // geometry count
	nobjects    int // non-geometry count
}

var counter uint64

// New creates an empty collection
func New() *Collection {
	col := &Collection{
		index:    index.New(),
		items:    btree.New(128, idOrdered),
		values:   btree.New(128, valueOrdered),
		fieldMap: make(map[string]int),
	}
	return col
}

func (c *Collection) setFieldValues(id string, values []float64) {
	if c.fieldValues == nil {
		c.fieldValues = make(map[string][]float64)
	}
	c.fieldValues[id] = values
}

func (c *Collection) getFieldValues(id string) (values []float64) {
	if c.fieldValues == nil {
		return nil
	}
	return c.fieldValues[id]
}
func (c *Collection) deleteFieldValues(id string) {
	if c.fieldValues != nil {
		delete(c.fieldValues, id)
	}
}

// Count returns the number of objects in collection.
func (c *Collection) Count() int {
	return c.objects + c.nobjects
}

// StringCount returns the number of string values.
func (c *Collection) StringCount() int {
	return c.nobjects
}

// PointCount returns the number of points (lat/lon coordinates) in collection.
func (c *Collection) PointCount() int {
	return c.points
}

// TotalWeight calculates the in-memory cost of the collection in bytes.
func (c *Collection) TotalWeight() int {
	return c.weight
}

// Bounds returns the bounds of all the items in the collection.
func (c *Collection) Bounds() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	return c.index.Bounds()
}

// ReplaceOrInsert adds or replaces an object in the collection and returns the fields array.
// If an item with the same id is already in the collection then the new item will adopt the old item's fields.
// The fields argument is optional.
// The return values are the old object, the old fields, and the new fields
func (c *Collection) ReplaceOrInsert(id string, obj geojson.Object, fields []string, values []float64) (oldObject geojson.Object, oldFields []float64, newFields []float64) {
	var oldItem *itemT
	var newItem *itemT = &itemT{id: id, object: obj}
	// add the new item to main btree and remove the old one if needed
	oldItemPtr := c.items.ReplaceOrInsert(newItem)
	if oldItemPtr != nil {
		// the old item was removed, now let's remove from the rtree
		// or strings tree.
		oldItem = oldItemPtr.(*itemT)
		if obj.IsGeometry() {
			// geometry
			c.index.Remove(oldItem)
			c.objects--
		} else {
			// string
			c.values.Delete(oldItem)
			c.nobjects--
		}
		// decrement the point count
		c.points -= oldItem.object.PositionCount()

		// decrement the weights
		c.weight -= len(c.getFieldValues(id)) * 8
		c.weight -= oldItem.object.Weight() + len(oldItem.id)

		// references
		oldObject = oldItem.object
		oldFields = c.getFieldValues(id)
		newFields = oldFields
	}
	// insert the new item into the rtree or strings tree.
	if obj.IsGeometry() {
		c.index.Insert(newItem)
		c.objects++
	} else {
		c.values.ReplaceOrInsert(newItem)
		c.nobjects++
	}
	// increment the point count
	c.points += obj.PositionCount()

	// add the new weights
	c.weight += len(newFields) * 8
	c.weight += obj.Weight() + len(id)
	if fields == nil {
		if len(values) > 0 {
			// directly set the field values, update weight
			c.weight -= len(newFields) * 8
			newFields = values
			c.setFieldValues(id, newFields)
			c.weight += len(newFields) * 8
		}
	} else {
		//if len(fields) == 0 {
		//	panic("if fields is empty, make it nil")
		//}
		// map field name to value
		for i, field := range fields {
			c.setField(newItem, field, values[i])
		}
		newFields = c.getFieldValues(id)
	}
	return oldObject, oldFields, newFields
}

// Remove removes an object and returns it.
// If the object does not exist then the 'ok' return value will be false.
func (c *Collection) Remove(id string) (obj geojson.Object, fields []float64, ok bool) {
	i := c.items.Delete(&itemT{id: id})
	if i == nil {
		return nil, nil, false
	}
	item := i.(*itemT)
	if item.object.IsGeometry() {
		c.index.Remove(item)
		c.objects--
	} else {
		c.values.Delete(item)
		c.nobjects--
	}
	fields = c.getFieldValues(id)
	c.deleteFieldValues(id)
	c.weight -= len(fields) * 8
	c.weight -= item.object.Weight() + len(item.id)
	c.points -= item.object.PositionCount()
	return item.object, fields, true
}

// Get returns an object.
// If the object does not exist then the 'ok' return value will be false.
func (c *Collection) Get(id string) (obj geojson.Object, fields []float64, ok bool) {
	i := c.items.Get(&itemT{id: id})
	if i == nil {
		return nil, nil, false
	}
	item := i.(*itemT)
	return item.object, c.getFieldValues(id), true
}

// SetField set a field value for an object and returns that object.
// If the object does not exist then the 'ok' return value will be false.
func (c *Collection) SetField(id, field string, value float64) (obj geojson.Object, fields []float64, updated bool, ok bool) {
	i := c.items.Get(&itemT{id: id})
	if i == nil {
		ok = false
		return
	}
	item := i.(*itemT)
	updated = c.setField(item, field, value)
	return item.object, c.getFieldValues(id), updated, true
}

func (c *Collection) setField(item *itemT, field string, value float64) (updated bool) {
	idx, ok := c.fieldMap[field]
	if !ok {
		idx = len(c.fieldMap)
		c.fieldMap[field] = idx
	}
	fields := c.getFieldValues(item.id)
	c.weight -= len(fields) * 8
	for idx >= len(fields) {
		fields = append(fields, 0)
	}
	c.weight += len(fields) * 8
	ovalue := fields[idx]
	fields[idx] = value
	c.setFieldValues(item.id, fields)
	return ovalue != value
}

// FieldMap return a maps of the field names.
func (c *Collection) FieldMap() map[string]int {
	return c.fieldMap
}

// FieldArr return an array representation of the field names.
func (c *Collection) FieldArr() []string {
	arr := make([]string, len(c.fieldMap))
	for field, i := range c.fieldMap {
		arr[i] = field
	}
	return arr
}

// Scan iterates though the collection ids. A cursor can be used for paging.
func (c *Collection) Scan(cursor uint64, desc bool,
	iterator func(id string, obj geojson.Object, fields []float64) bool,
) (ncursor uint64) {
	var i uint64
	var active = true
	iter := func(item btree.Item) bool {
		if i >= cursor {
			iitm := item.(*itemT)
			active = iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id))
		}
		i++
		return active
	}
	if desc {
		c.items.Descend(iter)
	} else {
		c.items.Ascend(iter)
	}
	return i
}

// ScanGreaterOrEqual iterates though the collection starting with specified id. A cursor can be used for paging.
func (c *Collection) ScanRange(cursor uint64, start, end string, desc bool,
	iterator func(id string, obj geojson.Object, fields []float64) bool,
) (ncursor uint64) {
	var i uint64
	var active = true
	iter := func(item btree.Item) bool {
		if i >= cursor {
			iitm := item.(*itemT)
			active = iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id))
		}
		i++
		return active
	}

	if desc {
		c.items.DescendRange(&itemT{id: start}, &itemT{id: end}, iter)
	} else {
		c.items.AscendRange(&itemT{id: start}, &itemT{id: end}, iter)
	}
	return i
}

// SearchValues iterates though the collection values. A cursor can be used for paging.
func (c *Collection) SearchValues(cursor uint64, desc bool,
	iterator func(id string, obj geojson.Object, fields []float64) bool,
) (ncursor uint64) {
	var i uint64
	var active = true
	iter := func(item btree.Item) bool {
		if i >= cursor {
			iitm := item.(*itemT)
			active = iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id))
		}
		i++
		return active
	}
	if desc {
		c.values.Descend(iter)
	} else {
		c.values.Ascend(iter)
	}
	return i
}

// SearchValuesRange iterates though the collection values. A cursor can be used for paging.
func (c *Collection) SearchValuesRange(cursor uint64, start, end string, desc bool,
	iterator func(id string, obj geojson.Object, fields []float64) bool,
) (ncursor uint64) {
	var i uint64
	var active = true
	iter := func(item btree.Item) bool {
		if i >= cursor {
			iitm := item.(*itemT)
			active = iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id))
		}
		i++
		return active
	}
	if desc {
		c.values.DescendRange(&itemT{object: geojson.String(start)}, &itemT{object: geojson.String(end)}, iter)
	} else {
		c.values.AscendRange(&itemT{object: geojson.String(start)}, &itemT{object: geojson.String(end)}, iter)
	}
	return i
}

// ScanGreaterOrEqual iterates though the collection starting with specified id. A cursor can be used for paging.
func (c *Collection) ScanGreaterOrEqual(id string, cursor uint64, desc bool,
	iterator func(id string, obj geojson.Object, fields []float64) bool,
) (ncursor uint64) {
	var i uint64
	var active = true
	iter := func(item btree.Item) bool {
		if i >= cursor {
			iitm := item.(*itemT)
			active = iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id))
		}
		i++
		return active
	}
	if desc {
		c.items.DescendLessOrEqual(&itemT{id: id}, iter)
	} else {
		c.items.AscendGreaterOrEqual(&itemT{id: id}, iter)
	}
	return i
}

func (c *Collection) geoSearch(cursor uint64, bbox geojson.BBox, iterator func(id string, obj geojson.Object, fields []float64) bool) (ncursor uint64) {
	return c.index.Search(cursor, bbox.Min.Y, bbox.Min.X, bbox.Max.Y, bbox.Max.X, bbox.Min.Z, bbox.Max.Z, func(item index.Item) bool {
		var iitm *itemT
		iitm, ok := item.(*itemT)
		if !ok {
			return true // just ignore
		}
		if !iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id)) {
			return false
		}
		return true
	})
}

// Nearby returns all object that are nearby a point.
func (c *Collection) Nearby(cursor uint64, sparse uint8, lat, lon, meters, minZ, maxZ float64, iterator func(id string, obj geojson.Object, fields []float64) bool) (ncursor uint64) {
	center := geojson.Position{X: lon, Y: lat, Z: 0}
	bbox := geojson.BBoxesFromCenter(lat, lon, meters)
	bboxes := bbox.Sparse(sparse)
	if sparse > 0 {
		for _, bbox := range bboxes {
			bbox.Min.Z, bbox.Max.Z = minZ, maxZ
			c.geoSearch(cursor, bbox, func(id string, obj geojson.Object, fields []float64) bool {
				if obj.Nearby(center, meters) {
					if iterator(id, obj, fields) {
						return false
					}
				}
				return true
			})
		}
		return 0
	}
	bbox.Min.Z, bbox.Max.Z = minZ, maxZ
	return c.geoSearch(cursor, bbox, func(id string, obj geojson.Object, fields []float64) bool {
		if obj.Nearby(center, meters) {
			return iterator(id, obj, fields)
		}
		return true
	})
}

// Within returns all object that are fully contained within an object or bounding box. Set obj to nil in order to use the bounding box.
func (c *Collection) Within(cursor uint64, sparse uint8, obj geojson.Object, minLat, minLon, maxLat, maxLon, minZ, maxZ float64, iterator func(id string, obj geojson.Object, fields []float64) bool) (ncursor uint64) {
	var bbox geojson.BBox
	if obj != nil {
		bbox = obj.CalculatedBBox()
		if minZ == math.Inf(-1) && maxZ == math.Inf(+1) {
			if bbox.Min.Z == 0 && bbox.Max.Z == 0 {
				bbox.Min.Z = minZ
				bbox.Max.Z = maxZ
			}
		}
	} else {
		bbox = geojson.BBox{Min: geojson.Position{X: minLon, Y: minLat, Z: minZ}, Max: geojson.Position{X: maxLon, Y: maxLat, Z: maxZ}}
	}
	bboxes := bbox.Sparse(sparse)
	if sparse > 0 {
		for _, bbox := range bboxes {
			if obj != nil {
				c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
					if o.Within(obj) {
						if iterator(id, o, fields) {
							return false
						}
					}
					return true
				})
			}
			c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
				if o.WithinBBox(bbox) {
					if iterator(id, o, fields) {
						return false
					}
				}
				return true
			})
		}
		return 0
	}
	if obj != nil {
		return c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
			if o.Within(obj) {
				return iterator(id, o, fields)
			}
			return true
		})
	}
	return c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
		if o.WithinBBox(bbox) {
			return iterator(id, o, fields)
		}
		return true
	})
}

// Intersects returns all object that are intersect an object or bounding box. Set obj to nil in order to use the bounding box.
func (c *Collection) Intersects(cursor uint64, sparse uint8, obj geojson.Object, minLat, minLon, maxLat, maxLon, minZ, maxZ float64, iterator func(id string, obj geojson.Object, fields []float64) bool) (ncursor uint64) {
	var bbox geojson.BBox
	if obj != nil {
		bbox = obj.CalculatedBBox()
		if minZ == math.Inf(-1) && maxZ == math.Inf(+1) {
			if bbox.Min.Z == 0 && bbox.Max.Z == 0 {
				bbox.Min.Z = minZ
				bbox.Max.Z = maxZ
			}
		}
	} else {
		bbox = geojson.BBox{Min: geojson.Position{X: minLon, Y: minLat, Z: minZ}, Max: geojson.Position{X: maxLon, Y: maxLat, Z: maxZ}}
	}
	var bboxes []geojson.BBox
	if sparse > 0 {
		split := 1 << sparse
		xpart := (bbox.Max.X - bbox.Min.X) / float64(split)
		ypart := (bbox.Max.Y - bbox.Min.Y) / float64(split)
		for y := bbox.Min.Y; y < bbox.Max.Y; y += ypart {
			for x := bbox.Min.X; x < bbox.Max.X; x += xpart {
				bboxes = append(bboxes, geojson.BBox{
					Min: geojson.Position{X: x, Y: y, Z: minZ},
					Max: geojson.Position{X: x + xpart, Y: y + ypart, Z: maxZ},
				})
			}
		}
		for _, bbox := range bboxes {
			if obj != nil {
				c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
					if o.Intersects(obj) {
						if iterator(id, o, fields) {
							return false
						}
					}
					return true
				})
			}
			c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
				if o.IntersectsBBox(bbox) {
					if iterator(id, o, fields) {
						return false
					}
				}
				return true
			})
		}
		return 0
	}
	if obj != nil {
		return c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
			if o.Intersects(obj) {
				return iterator(id, o, fields)
			}
			return true
		})
	}
	return c.geoSearch(cursor, bbox, func(id string, o geojson.Object, fields []float64) bool {
		if o.IntersectsBBox(bbox) {
			return iterator(id, o, fields)
		}
		return true
	})
}

func (c *Collection) NearestNeighbors(k int, lat, lon float64, iterator func(id string, obj geojson.Object, fields []float64) bool) {
	c.index.NearestNeighbors(k, lat, lon, func(item index.Item) bool {
		var iitm *itemT
		iitm, ok := item.(*itemT)
		if !ok {
			return true // just ignore
		}
		if !iterator(iitm.id, iitm.object, c.getFieldValues(iitm.id)) {
			return false
		}
		return true
	})
}
