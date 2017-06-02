package canary

type equalFn func(interface{}, interface{}) bool

// NewUniqueSet returns a new instance of UniqueSet.
func NewUniqueSet(fn equalFn) *UniqueSet {
	return &UniqueSet{
		uniqueFunc: fn,
	}
}

// UniqueSet defines a type to create a unique set of values.
type UniqueSet struct {
	items []interface{}

	uniqueFunc equalFn
}

// Count returns count of all elements.
func (us *UniqueSet) Count() int {
	return len(us.items)
}

// Add adds the given item into the set if it not yet included.
func (us *UniqueSet) Add(item interface{}) interface{} {
	for _, item2 := range us.items {
		if !us.uniqueFunc(item, item2) {
			continue
		}

		return item2
	}

	us.items = append(us.items, item)
	return item
}

// Remove removes the item from the interal set.
func (us *UniqueSet) Remove(item interface{}) {
	for i, item2 := range us.items {
		if item != item2 {
			continue
		}

		us.items[i] = nil

		us.items = append(us.items[:i], us.items[i+1:]...)
		return
	}
}

// Each runs the function against all items in set.
func (us *UniqueSet) Each(fn func(int, interface{})) {
	items := us.items[:]

	for i, item := range items {
		fn(i, item)
	}
}

// Find runs the function against all items in set.
func (us *UniqueSet) Find(fn func(interface{}) bool) interface{} {
	for _, item := range us.items {
		if fn(item) {
			return item
		}
	}

	return nil
}
