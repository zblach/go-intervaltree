package intervaltree

import (
	"sync"
)

// imbalanceThreshold is the threshold for when to rebalance the tree.
const imbalanceThreshold = 1

// Interval represents an interval with a key of type T and values of type []V.
type Interval[T comparable, V any] struct {
	Key    T
	Values []V
}

// IntervalNode represents a node in the interval tree.
type intervalNode[T comparable, V any] struct {
	interval Interval[T, V]
	left     *intervalNode[T, V]
	right    *intervalNode[T, V]
	height   uint
}

// CollisionHandlers are functions that handlescollisions when inserting an interval into the interval tree.
type CollisionHandler[T comparable, V any] func(existing Interval[T, V], new V) Interval[T, V]

// Replace is a collision handler that replaces the existing interval with the new interval.
func Replace[T comparable, V any](existing Interval[T, V], new V) Interval[T, V] {
	return Interval[T, V]{Key: existing.Key, Values: []V{new}}
}

// Append is a collision handler that appends the new interval to the existing interval.
func Append[T comparable, V any](existing Interval[T, V], new V) Interval[T, V] {
	existing.Values = append(existing.Values, new)
	return existing
}

// IntervalTree represents an interval tree.
type IntervalTree[T comparable, V any] struct {
	mutex            sync.RWMutex
	root             *intervalNode[T, V]
	lessFunc         func(a, b T) bool
	collisionHandler CollisionHandler[T, V]
}

// NewIntervalTree creates a new instance of IntervalTree with the specified less function and collision handling strategy.
func New[T comparable, V any](lessFunc func(a, b T) bool, collisionHandler CollisionHandler[T, V]) *IntervalTree[T, V] {
	return &IntervalTree[T, V]{
		lessFunc:         lessFunc,
		collisionHandler: collisionHandler,
	}
}

// Unique creates a new instance of IntervalTree with the specified less function and Replacement collision handling strategy.
func Unique[T comparable, V any](lessFunc func(a, b T) bool) *IntervalTree[T, V] {
	return New[T, V](lessFunc, Replace[T, V])
}

// Duplicates creates a new instance of IntervalTree with the specified less function and Append collision handling strategy.
func Duplicates[T comparable, V any](lessFunc func(a, b T) bool) *IntervalTree[T, V] {
	return New[T, V](lessFunc, Append[T, V])
}

// Insert inserts a new value into the interval tree
func (tree *IntervalTree[T, V]) Insert(key T, value V) {
	tree.mutex.Lock()
	defer tree.mutex.Unlock()

	tree.root = tree.insertNode(tree.root, key, value)
}

// insertNode recursively inserts a new value into the interval tree.
func (tree *IntervalTree[T, V]) insertNode(node *intervalNode[T, V], key T, value V) *intervalNode[T, V] {
	if node == nil {
		return &intervalNode[T, V]{
			interval: Interval[T, V]{Key: key, Values: []V{value}},
			left:     nil,
			right:    nil,
			height:   1,
		}
	}

	switch {
	case tree.lessFunc(key, node.interval.Key):
		node.left = tree.insertNode(node.left, key, value)
	case tree.lessFunc(node.interval.Key, key):
		node.right = tree.insertNode(node.right, key, value)
	default:
		// Handle interval collision
		node.interval = tree.collisionHandler(node.interval, value)
	}

	node.height = maxUint(getHeight(node.left), getHeight(node.right)) + 1

	// Rebalance the tree
	switch balanceFactor := getBalance(node); {
	// Left Left Case
	case balanceFactor > imbalanceThreshold && tree.lessFunc(key, node.left.interval.Key):
		return tree.rightRotate(node)
	// Right Right Case
	case balanceFactor < -imbalanceThreshold && tree.lessFunc(node.right.interval.Key, key):
		return tree.leftRotate(node)
	// Left Right Case
	case balanceFactor > imbalanceThreshold && tree.lessFunc(node.left.interval.Key, key):
		node.left = tree.leftRotate(node.left)
		return tree.rightRotate(node)
	// Right Left Case
	case balanceFactor < -imbalanceThreshold && tree.lessFunc(key, node.right.interval.Key):
		node.right = tree.rightRotate(node.right)
		return tree.leftRotate(node)
	}

	return node
}

// Delete deletes an entry from the interval tree.
func (tree *IntervalTree[T, V]) Delete(key T) {
	tree.mutex.Lock()
	defer tree.mutex.Unlock()

	tree.root = tree.deleteNode(tree.root, key)
}

// deleteNode recursively deletes a value from the interval tree.
func (tree *IntervalTree[T, V]) deleteNode(node *intervalNode[T, V], key T) *intervalNode[T, V] {
	if node == nil {
		return node
	}

	switch {
	case tree.lessFunc(key, node.interval.Key):
		node.left = tree.deleteNode(node.left, key)
	case tree.lessFunc(node.interval.Key, key):
		node.right = tree.deleteNode(node.right, key)
	default:
		// node is the node to be deleted
		if node.left == nil && node.right == nil {
			node = nil
		} else if node.left == nil {
			node = node.right
		} else if node.right == nil {
			node = node.left
		} else {
			// node has two children, get the in-order successor
			successor := tree.getMinValueNode(node.right)
			node.interval = successor.interval
			node.right = tree.deleteNode(node.right, successor.interval.Key)
		}
	}

	if node == nil {
		return node
	}

	node.height = maxUint(getHeight(node.left), getHeight(node.right)) + 1

	// Rebalance the tree
	switch balanceFactor := getBalance(node); {
	// Left Left Case
	case balanceFactor > imbalanceThreshold && getBalance(node.left) >= 0:
		return tree.rightRotate(node)
	// Right Right Case
	case balanceFactor < -imbalanceThreshold && getBalance(node.right) <= 0:
		return tree.leftRotate(node)
	// Left Right Case
	case balanceFactor > imbalanceThreshold && getBalance(node.left) < 0:
		node.left = tree.leftRotate(node.left)
		return tree.rightRotate(node)
	// Right Left Case
	case balanceFactor < -imbalanceThreshold && getBalance(node.right) > 0:
		node.right = tree.rightRotate(node.right)
		return tree.leftRotate(node)
	}

	return node
}

// getMinValueNode returns the node with the minimum value in the tree.
func (tree *IntervalTree[T, V]) getMinValueNode(node *intervalNode[T, V]) *intervalNode[T, V] {
	for node.left != nil {
		node = node.left
	}
	return node
}

// Entry represents a (T, V) tuple in the interval tree.
type Entry[T comparable, V any] struct {
	Key   T
	Value V
}

// Search searches for values that exist within the given range.
func (tree *IntervalTree[T, V]) Search(start, end T) []Entry[T, V] {
	tree.mutex.RLock()
	defer tree.mutex.RUnlock()

	a := start
	b := end
	// swap start/end to ensure that there's always a positive range
	if tree.lessFunc(end, start) {
		a = end
		b = start
	}

	results := make([]Entry[T, V], 0)
	tree.searchNodes(tree.root, a, b, &results)
	return results
}

// searchNodes searches for intervals that overlap with the given range and appends the overlapping intervals to the results slice.
func (tree *IntervalTree[T, V]) searchNodes(node *intervalNode[T, V], start, end T, results *[]Entry[T, V]) {
	switch {
	case node == nil:
		return
	case tree.lessFunc(end, node.interval.Key):
		tree.searchNodes(node.left, start, end, results)
	case tree.lessFunc(node.interval.Key, start):
		tree.searchNodes(node.right, start, end, results)
	default:
		tree.searchNodes(node.left, start, end, results)
		// Interval overlaps, flatten and append the (T, V) tuple to the results.
		for _, v := range node.interval.Values {
			*results = append(*results, Entry[T, V]{node.interval.Key, v})
		}
		tree.searchNodes(node.right, start, end, results)
	}
}

// rightRotate performs a right rotation on the given node and returns the new root.
func (tree *IntervalTree[T, V]) rightRotate(node *intervalNode[T, V]) *intervalNode[T, V] {
	l := node.left
	lr := l.right

	l.right = node
	node.left = lr

	node.height = maxUint(getHeight(node.left), getHeight(node.right)) + 1
	l.height = maxUint(getHeight(l.left), getHeight(l.right)) + 1

	return l
}

// leftRotate performs a left rotation on the given node and returns the new root.
func (tree *IntervalTree[T, V]) leftRotate(node *intervalNode[T, V]) *intervalNode[T, V] {
	r := node.right
	rl := r.left

	r.left = node
	node.right = rl

	node.height = maxUint(getHeight(node.left), getHeight(node.right)) + 1
	r.height = maxUint(getHeight(r.left), getHeight(r.right)) + 1

	return r
}

// getHeight returns the height of the given node.
func getHeight[T comparable, V any](node *intervalNode[T, V]) uint {
	if node == nil {
		return 0
	}
	return node.height
}

// getBalance returns the balance factor of the given node.
func getBalance[T comparable, V any](node *intervalNode[T, V]) int {
	if node == nil {
		return 0
	}
	return int(getHeight(node.left)) - int(getHeight(node.right))
}

func maxUint(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}
