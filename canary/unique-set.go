package canary

type equalFn func(interface{}, interface{}) bool

func NewUniqueSet(fn equalFn) *UniqueSet {
	return &UniqueSet{
		uniqueFunc: fn,
	}
}

type UniqueSet struct {
	items []interface{}

	uniqueFunc equalFn
}

func (us *UniqueSet) Count() int {
	return len(us.items)
}

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

func (us *UniqueSet) Each(fn func(int, interface{})) {
	items := us.items[:]

	for i, item := range items {
		fn(i, item)
	}
}

func (us *UniqueSet) Find(fn func(interface{}) bool) interface{} {
	for _, item := range us.items {
		if fn(item) {
			return item
		}
	}

	return nil
}
