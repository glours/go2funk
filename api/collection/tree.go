package collection

import (
	"cmp"
	"fmt"
	"iter"
	"strings"

	"github.com/glours/go2funk/api/control"
)

// The balance invariant, in one paragraph.
//
// Every node remembers the size of the subtree it roots. A node is balanced
// when neither child is more than balanceDelta times larger than the other, or
// when the two together hold at most one value. Insertions and deletions can
// only ever break that condition by one step, at one node, so a single rotation
// at each level on the way back up is enough to restore it.
//
// balanceRatio decides which rotation: when the inner grandchild is small
// relative to the outer one a single rotation suffices, otherwise the inner
// grandchild has to be lifted, which takes a double rotation.
//
// Picking the two constants is subtler than it looks. The scheme is Adams', but
// his own papers used delta = 4 for sets and 5 for maps, and his analysis of
// which pairs are valid turned out to be wrong: Hirai and Yamamoto (2011)
// established the real range with a Coq formalisation, producing counterexample
// trees for pairs that had been thought sound. Haskell's containers settled on
// delta = 3 afterwards, which is what these values follow.
//
// The lesson worth keeping: an invariant that a paper asserts is not the same as
// one a machine has checked, which is why the tests here replay thousands of
// random operations against the condition rather than trusting the constants.
const (
	balanceDelta = 3
	balanceRatio = 2
)

// Tree is an immutable, persistent ordered set of values.
//
// Ordering goes through cmp.Compare rather than the < operator, so a float NaN
// has a place in the order — below every other value, and equal to itself —
// instead of comparing false against everything and silently swallowing the
// values around it.
//
// Values are kept sorted, so iteration is in order, and Min, Max and Range come
// for free. Inserting or deleting rebuilds only the path from the root to the
// change — O(log n) nodes — and shares every other node with the tree it came
// from, which is what makes keeping old versions around cheap.
//
// The zero value is the empty tree. Len and IsEmpty are O(1); Insert, Delete,
// Contains, Min and Max are O(log n); anything that visits every value is O(n).
type Tree[T cmp.Ordered] struct {
	root *treeNode[T]
}

// treeNode is one node. size counts the values in the whole subtree, which is
// what both Len and the balance invariant read.
type treeNode[T cmp.Ordered] struct {
	value       T
	left, right *treeNode[T]
	size        int
}

// EmptyTree returns the empty tree. It is the same as the zero value of Tree.
func EmptyTree[T cmp.Ordered]() Tree[T] {
	return Tree[T]{}
}

// TreeOf returns a tree holding values. Duplicates collapse into one. O(n log n).
func TreeOf[T cmp.Ordered](values ...T) Tree[T] {
	tree := EmptyTree[T]()
	for _, value := range values {
		tree = tree.Insert(value)
	}
	return tree
}

// CollectTree drains seq into a new tree. O(n log n).
func CollectTree[T cmp.Ordered](seq iter.Seq[T]) Tree[T] {
	tree := EmptyTree[T]()
	for value := range seq {
		tree = tree.Insert(value)
	}
	return tree
}

// IsEmpty reports whether the tree holds no value. O(1).
func (t Tree[T]) IsEmpty() bool { return t.root == nil }

// Len returns the number of values. O(1).
func (t Tree[T]) Len() int { return treeSize(t.root) }

// Contains reports whether value is in the tree. O(log n).
func (t Tree[T]) Contains(value T) bool {
	for current := t.root; current != nil; {
		switch cmp.Compare(value, current.value) {
		case -1:
			current = current.left
		case 1:
			current = current.right
		default:
			return true
		}
	}
	return false
}

// Insert returns a tree holding value. Inserting a value that is already there
// returns the very same tree, sharing every node, rather than rebuilding the
// path to it. O(log n).
func (t Tree[T]) Insert(value T) Tree[T] {
	return Tree[T]{root: insertNode(t.root, value)}
}

// Delete returns a tree without value. Deleting a value that is not there
// returns the very same tree, sharing every node. O(log n).
func (t Tree[T]) Delete(value T) Tree[T] {
	return Tree[T]{root: deleteNode(t.root, value)}
}

// Min returns the smallest value, or None if the tree is empty. O(log n).
func (t Tree[T]) Min() control.Option[T] {
	if t.root == nil {
		return control.None[T]()
	}
	current := t.root
	for current.left != nil {
		current = current.left
	}
	return control.Some(current.value)
}

// Max returns the largest value, or None if the tree is empty. O(log n).
func (t Tree[T]) Max() control.Option[T] {
	if t.root == nil {
		return control.None[T]()
	}
	current := t.root
	for current.right != nil {
		current = current.right
	}
	return control.Some(current.value)
}

// All returns an iterator over the values, smallest first. O(n) for a full walk.
func (t Tree[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) { walkAscending(t.root, yield) }
}

// Backward returns an iterator over the values, largest first.
func (t Tree[T]) Backward() iter.Seq[T] {
	return func(yield func(T) bool) { walkDescending(t.root, yield) }
}

// Range returns an iterator over the values between from and to, both included.
// Subtrees that cannot hold a value in the range are not visited, so a narrow
// range costs O(log n + k) for k values rather than O(n).
func (t Tree[T]) Range(from, to T) iter.Seq[T] {
	return func(yield func(T) bool) { walkRange(t.root, from, to, yield) }
}

// ToSlice returns the values as a new sorted slice. O(n).
func (t Tree[T]) ToSlice() []T {
	result := make([]T, 0, t.Len())
	for value := range t.All() {
		result = append(result, value)
	}
	return result
}

// Filter returns a tree of the values matching the predicate. O(n log n).
func (t Tree[T]) Filter(predicate func(T) bool) Tree[T] {
	result := EmptyTree[T]()
	for value := range t.All() {
		if predicate(value) {
			result = result.Insert(value)
		}
	}
	return result
}

// Map returns the tree of mapper applied to every value. Values that map to the
// same result collapse into one, so the result can be smaller than the receiver:
// a set holds each value once, and which of the originals produced it makes no
// difference. O(n log n).
func (t Tree[T]) Map[U cmp.Ordered](mapper func(T) U) Tree[U] {
	result := EmptyTree[U]()
	for value := range t.All() {
		result = result.Insert(mapper(value))
	}
	return result
}

// Fold combines the values in order, starting from initial. O(n).
func (t Tree[T]) Fold[U any](initial U, combine func(U, T) U) U {
	result := initial
	for value := range t.All() {
		result = combine(result, value)
	}
	return result
}

// String renders the tree as Tree(a, b, c), in order.
func (t Tree[T]) String() string {
	var builder strings.Builder
	builder.WriteString("Tree(")
	first := true
	for value := range t.All() {
		if !first {
			builder.WriteString(", ")
		}
		fmt.Fprintf(&builder, "%v", value)
		first = false
	}
	builder.WriteString(")")
	return builder.String()
}

// treeSize reads the cached size, treating the empty subtree as zero.
func treeSize[T cmp.Ordered](node *treeNode[T]) int {
	if node == nil {
		return 0
	}
	return node.size
}

// newTreeNode builds a node and caches its size. The children must already
// satisfy the balance invariant between themselves.
func newTreeNode[T cmp.Ordered](value T, left, right *treeNode[T]) *treeNode[T] {
	return &treeNode[T]{
		value: value,
		left:  left,
		right: right,
		size:  1 + treeSize(left) + treeSize(right),
	}
}

// balanceNode rebuilds a node whose children may be one step out of balance,
// which is all an insertion or a deletion can cause at a single level.
func balanceNode[T cmp.Ordered](value T, left, right *treeNode[T]) *treeNode[T] {
	sizeLeft, sizeRight := treeSize(left), treeSize(right)

	switch {
	case sizeLeft+sizeRight <= 1:
		return newTreeNode(value, left, right)
	case sizeRight > balanceDelta*sizeLeft:
		return rotateLeft(value, left, right)
	case sizeLeft > balanceDelta*sizeRight:
		return rotateRight(value, left, right)
	default:
		return newTreeNode(value, left, right)
	}
}

// rotateLeft moves weight from an overgrown right child to the left.
func rotateLeft[T cmp.Ordered](value T, left, right *treeNode[T]) *treeNode[T] {
	if treeSize(right.left) < balanceRatio*treeSize(right.right) {
		// The outer grandchild carries the weight: lifting the right child is enough.
		return newTreeNode(right.value, newTreeNode(value, left, right.left), right.right)
	}
	// The inner grandchild carries it, so that is the one to lift.
	inner := right.left
	return newTreeNode(inner.value,
		newTreeNode(value, left, inner.left),
		newTreeNode(right.value, inner.right, right.right))
}

// rotateRight is rotateLeft with the sides exchanged.
func rotateRight[T cmp.Ordered](value T, left, right *treeNode[T]) *treeNode[T] {
	if treeSize(left.right) < balanceRatio*treeSize(left.left) {
		return newTreeNode(left.value, left.left, newTreeNode(value, left.right, right))
	}
	inner := left.right
	return newTreeNode(inner.value,
		newTreeNode(left.value, left.left, inner.left),
		newTreeNode(value, inner.right, right))
}

func insertNode[T cmp.Ordered](node *treeNode[T], value T) *treeNode[T] {
	if node == nil {
		return newTreeNode(value, nil, nil)
	}
	switch cmp.Compare(value, node.value) {
	case -1:
		grown := insertNode(node.left, value)
		if grown == node.left {
			return node // nothing changed below: share this node too
		}
		return balanceNode(node.value, grown, node.right)
	case 1:
		grown := insertNode(node.right, value)
		if grown == node.right {
			return node
		}
		return balanceNode(node.value, node.left, grown)
	default:
		// Already there: return the node untouched so the caller shares it.
		return node
	}
}

func deleteNode[T cmp.Ordered](node *treeNode[T], value T) *treeNode[T] {
	if node == nil {
		return nil
	}
	switch cmp.Compare(value, node.value) {
	case -1:
		shrunk := deleteNode(node.left, value)
		if shrunk == node.left {
			return node // the value was not there: share this node too
		}
		return balanceNode(node.value, shrunk, node.right)
	case 1:
		shrunk := deleteNode(node.right, value)
		if shrunk == node.right {
			return node
		}
		return balanceNode(node.value, node.left, shrunk)
	default:
		return joinAfterDelete(node.left, node.right)
	}
}

// joinAfterDelete rejoins the two subtrees left behind by a deleted node. The
// replacement is taken from the heavier side, so the join cannot itself unbalance
// the result by more than one step.
func joinAfterDelete[T cmp.Ordered](left, right *treeNode[T]) *treeNode[T] {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	case treeSize(left) > treeSize(right):
		value, rest := detachMax(left)
		return balanceNode(value, rest, right)
	default:
		value, rest := detachMin(right)
		return balanceNode(value, left, rest)
	}
}

func detachMin[T cmp.Ordered](node *treeNode[T]) (T, *treeNode[T]) {
	if node.left == nil {
		return node.value, node.right
	}
	value, rest := detachMin(node.left)
	return value, balanceNode(node.value, rest, node.right)
}

func detachMax[T cmp.Ordered](node *treeNode[T]) (T, *treeNode[T]) {
	if node.right == nil {
		return node.value, node.left
	}
	value, rest := detachMax(node.right)
	return value, balanceNode(node.value, node.left, rest)
}

// walkAscending reports false as soon as the consumer stops the iteration.
func walkAscending[T cmp.Ordered](node *treeNode[T], yield func(T) bool) bool {
	if node == nil {
		return true
	}
	return walkAscending(node.left, yield) && yield(node.value) && walkAscending(node.right, yield)
}

func walkDescending[T cmp.Ordered](node *treeNode[T], yield func(T) bool) bool {
	if node == nil {
		return true
	}
	return walkDescending(node.right, yield) && yield(node.value) && walkDescending(node.left, yield)
}

// walkRange skips the subtrees that cannot hold a value between from and to.
func walkRange[T cmp.Ordered](node *treeNode[T], from, to T, yield func(T) bool) bool {
	if node == nil {
		return true
	}
	if cmp.Compare(node.value, from) > 0 && !walkRange(node.left, from, to, yield) {
		return false
	}
	if cmp.Compare(node.value, from) >= 0 && cmp.Compare(node.value, to) <= 0 && !yield(node.value) {
		return false
	}
	if cmp.Compare(node.value, to) < 0 && !walkRange(node.right, from, to, yield) {
		return false
	}
	return true
}
