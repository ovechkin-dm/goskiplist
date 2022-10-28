package skiplist

import (
	"math/rand"
)

type node[K, V any] struct {
	right *node[K, V]
	down  *node[K, V]
	key   *K
	value *V
}

type Comparator[T any] func(T, T) int

type Pair[K any, V any] struct {
	Key   K
	Value V
}

type Map[K any, V any] interface {
	Get(k K) (V, bool)
	Add(k K, v V)
	Remove(k K) (V, bool)
	ForEach(func(K, V) bool)
	Lt(k K) (V, bool)
	Gt(k K) (V, bool)
	LtEq(k K) (V, bool)
	GtEq(k K) (V, bool)
	ForEachRange(K, K, func(K, V) bool)
	Size() int
}

type mapImpl[K any, V any] struct {
	cmp        Comparator[K]
	head       *node[K, V]
	zeroLevel  *node[K, V]
	levels     int
	emptyValue V
	size       int
}

func (m *mapImpl[K, V]) Get(k K) (V, bool) {
	lt := m.getLt(k)
	if lt == nil {
		return m.emptyValue, false
	}
	if lt.right == nil {
		return m.emptyValue, false
	}
	if lt.right.key == nil {
		return m.emptyValue, false
	}
	eq := m.cmp(k, *lt.right.key)
	if eq == 0 {
		return *lt.right.value, true
	}
	return m.emptyValue, false
}

func (m *mapImpl[K, V]) Lt(k K) (V, bool) {
	lt := m.getLt(k)
	if lt == nil {
		return m.emptyValue, false
	}
	if lt.key == nil {
		return m.emptyValue, false
	}
	return *lt.value, true
}

func (m *mapImpl[K, V]) Gt(k K) (V, bool) {
	lt := m.getLt(k)
	if lt == nil {
		return m.emptyValue, false
	}
	if lt.right == nil {
		return m.emptyValue, false
	}
	if lt.right.key == nil {
		return m.emptyValue, false
	}
	eq := m.cmp(k, *lt.right.key)
	if eq == 0 {
		lt = lt.right
		if lt.right == nil || lt.right.key == nil {
			return m.emptyValue, false
		}
	}
	return *lt.right.value, true
}

func (m *mapImpl[K, V]) LtEq(k K) (V, bool) {
	lt := m.getLt(k)
	if lt == nil {
		return m.emptyValue, false
	}
	if lt.right != nil && lt.right.key != nil && m.cmp(k, *lt.right.key) == 0 {
		return *lt.right.value, true
	}
	if lt == nil {
		return m.emptyValue, false
	}
	if lt.key == nil {
		return m.emptyValue, false
	}
	return *lt.value, true
}

func (m *mapImpl[K, V]) GtEq(k K) (V, bool) {
	lt := m.getLt(k)
	if lt == nil {
		return m.emptyValue, false
	}
	if lt.right != nil && lt.right.key != nil && m.cmp(k, *lt.right.key) == 0 {
		return *lt.right.value, true
	}
	if lt.right == nil {
		return m.emptyValue, false
	}
	if lt.right.key == nil {
		return m.emptyValue, false
	}
	eq := m.cmp(k, *lt.right.key)
	if eq == 0 {
		lt = lt.right
		if lt.right == nil || lt.right.key == nil {
			return m.emptyValue, false
		}
	}
	return *lt.right.value, true
}

func (m *mapImpl[K, V]) Add(k K, v V) {
	nodeLevel := randomLevel()
	if m.levels < nodeLevel {
		m.addLevels(nodeLevel)
	}
	head := m.head
	cur := head
	curLevel := m.levels
	var up *node[K, V]
	exists := false
	for cur != nil {
		n := &node[K, V]{
			right: nil,
			down:  nil,
			key:   &k,
			value: &v,
		}
		if cur.right == nil {
			if curLevel > nodeLevel {
				cur = cur.down
				curLevel -= 1
				continue
			}
			cur.right = n
			if up != nil {
				up.down = n
			}
			up = n
			cur = cur.down
			continue
		}
		cmp := m.cmp(k, *cur.right.key)
		if cmp < 0 {
			if curLevel > nodeLevel {
				cur = cur.down
				curLevel -= 1
				continue
			}
			if up != nil {
				up.down = n
			}
			n.right = cur.right
			cur.right = n
			cur = cur.down
			up = n
		} else if cmp > 0 {
			cur = cur.right
		} else {
			exists = true
			if curLevel > nodeLevel {
				cur.right = cur.right.right
				curLevel -= 1
			} else {
				if up != nil {
					up.down = cur.right
				}
				cur.right.value = &v
				up = cur.right
				cur = cur.down
			}
		}
	}
	if !exists {
		m.size += 1
	}
	m.shrink()
}

func (m *mapImpl[K, V]) getLt(k K) *node[K, V] {
	cur := m.head
	for cur.down != nil {
		if cur.right == nil {
			cur = cur.down
			continue
		}
		cmp := m.cmp(k, *cur.right.key)
		if cmp <= 0 {
			cur = cur.down
		} else {
			cur = cur.right
		}
	}
	for cur.right != nil {
		cmp := m.cmp(k, *cur.right.key)
		if cmp <= 0 {
			return cur
		} else if cmp > 0 {
			cur = cur.right
		}
	}
	return cur
}

func (m *mapImpl[K, V]) Remove(k K) (V, bool) {
	cur := m.head
	lvl := m.levels
	found := false
	foundValue := m.emptyValue
	for lvl >= 0 {
		if cur.right == nil {
			cur = cur.down
			lvl -= 1
		} else {
			cmp := m.cmp(k, *cur.right.key)
			if cmp > 0 {
				cur = cur.right
			} else if cmp < 0 {
				cur = cur.down
				lvl -= 1
			} else {
				foundValue = *cur.right.value
				found = true
				cur.right = cur.right.right
				cur = cur.down
				lvl -= 1
			}
		}
	}
	if found {
		m.size -= 1
	}
	m.shrink()
	return foundValue, found
}

func (m *mapImpl[K, V]) ForEachRange(start K, end K, f func(K, V) bool) {
	cur := m.getLt(start)
	if cur == nil {
		return
	}
	cur = cur.right
	for cur == nil {
		return
	}
	for cur != nil {
		cmp := m.cmp(*cur.key, end)
		if cmp >= 0 {
			return
		}
		mv := f(*cur.key, *cur.value)
		if !mv {
			return
		}
		cur = cur.right
	}
}

func (m *mapImpl[K, V]) ForEach(f func(k K, v V) bool) {
	cur := m.zeroLevel
	for cur != nil {
		if cur.key != nil {
			mv := f(*cur.key, *cur.value)
			if !mv {
				return
			}
		}
		cur = cur.right
	}
}

func (m *mapImpl[K, V]) Size() int {
	return m.size
}

func NewMap[K any, V any](cmp Comparator[K]) Map[K, V] {
	head := &node[K, V]{
		right: nil,
		down:  nil,
		key:   nil,
		value: nil,
	}
	return &mapImpl[K, V]{
		cmp:       cmp,
		head:      head,
		zeroLevel: head,
		levels:    0,
	}
}

func (m *mapImpl[K, V]) shrink() {
	for m.head.down != nil {
		if m.head.right == nil {
			m.head = m.head.down
			m.levels -= 1
		} else {
			break
		}
	}
}

func (m *mapImpl[K, V]) addLevels(levels int) {
	numLevels := levels - m.levels
	for i := 0; i < numLevels; i++ {
		prev := m.head
		cur := new(node[K, V])
		cur.down = prev
		m.head = cur
	}
	m.levels = levels
}

func randomLevel() int {
	coin := rand.Int63n(1 << 62)
	for i := 61; i >= 0; i-- {
		if coin > 1<<i {
			return 61 - i
		}
	}
	return 0
}
