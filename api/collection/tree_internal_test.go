package collection

// The balance invariant lives in the node layout, which no exported API lets a
// test reach. Checking it means walking the nodes, so this file is internal on
// purpose — it is the executable statement of what tree.go claims.

import (
	"cmp"
	"math/rand/v2"
	"testing"
)

// checkNode verifies, for every node of the subtree, that
//   - the cached size matches the values actually below it,
//   - values are ordered, left < node < right,
//   - neither child is more than balanceDelta times larger than the other.
//
// It returns the size it counted so the caller can compare it with the cache.
func checkNode[T cmp.Ordered](t *testing.T, node *treeNode[T], low, high *T) int {
	t.Helper()
	if node == nil {
		return 0
	}

	if low != nil && cmp.Compare(node.value, *low) <= 0 {
		t.Fatalf("value %v is not greater than its left bound %v", node.value, *low)
	}
	if high != nil && cmp.Compare(node.value, *high) >= 0 {
		t.Fatalf("value %v is not smaller than its right bound %v", node.value, *high)
	}

	sizeLeft := checkNode(t, node.left, low, &node.value)
	sizeRight := checkNode(t, node.right, &node.value, high)

	if got, want := node.size, 1+sizeLeft+sizeRight; got != want {
		t.Fatalf("cached size of %v is %d, want %d", node.value, got, want)
	}
	if sizeLeft+sizeRight > 1 {
		if sizeLeft > balanceDelta*sizeRight || sizeRight > balanceDelta*sizeLeft {
			t.Fatalf("node %v is out of balance: left=%d right=%d, delta=%d",
				node.value, sizeLeft, sizeRight, balanceDelta)
		}
	}
	return node.size
}

func checkTree[T cmp.Ordered](t *testing.T, tree Tree[T], wantSize int) {
	t.Helper()
	if got := checkNode(t, tree.root, nil, nil); got != wantSize {
		t.Fatalf("the tree holds %d values, want %d", got, wantSize)
	}
}

// Thousands of random insertions and deletions must never break the invariant.
func TestTreeKeepsItsInvariantUnderRandomOperations(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2)) // fixed seed: a failure is reproducible
	tree := EmptyTree[int]()
	present := map[int]bool{}

	for range 5000 {
		value := random.IntN(500)
		if random.IntN(2) == 0 {
			tree = tree.Insert(value)
			present[value] = true
		} else {
			tree = tree.Delete(value)
			delete(present, value)
		}
		checkTree(t, tree, len(present))
	}

	for value := range present {
		if !tree.Contains(value) {
			t.Fatalf("%d should be present", value)
		}
	}
}

// Sorted input is the worst case for a naive binary search tree: it degenerates
// into a linked list. A balanced one stays logarithmic.
func TestTreeStaysBalancedOnSortedInput(t *testing.T) {
	const count = 4096

	tree := EmptyTree[int]()
	for value := range count {
		tree = tree.Insert(value)
	}
	checkTree(t, tree, count)

	// log2(4096) is 12; the weight-balanced bound leaves some slack above it.
	if got, want := depth(tree.root), 30; got > want {
		t.Errorf("depth = %d after %d sorted insertions, want at most %d", got, count, want)
	}
}

func depth[T cmp.Ordered](node *treeNode[T]) int {
	if node == nil {
		return 0
	}
	return 1 + max(depth(node.left), depth(node.right))
}

// Inserting shares every node off the path from the root to the new value, so
// the cost is logarithmic in the size of the tree rather than proportional to it.
func TestTreeSharesStructureOnInsert(t *testing.T) {
	const count = 1000

	tree := EmptyTree[int]()
	for value := range count {
		tree = tree.Insert(value * 2) // leave gaps so the inserts below are new
	}

	// The result has to be retained: discarding it lets escape analysis elide
	// allocations and the measurement silently under-counts. Prepend on a List
	// reads as zero allocations that way, and as one when the result is kept.
	var retained Tree[int]
	allocations := testing.AllocsPerRun(100, func() { retained = tree.Insert(1) })
	if retained.Len() != count+1 {
		t.Fatalf("the retained tree holds %d values, want %d", retained.Len(), count+1)
	}
	if allocations > 30 {
		t.Errorf("inserting into a tree of %d values allocated %.0f nodes, want a logarithmic count",
			count, allocations)
	}
	t.Logf("inserting into a tree of %d values allocates %.0f nodes", count, allocations)

	// The tree it came from is untouched.
	checkTree(t, tree, count)
}

// The diagram in doc.go claims that inserting 4 into a tree holding
// 1, 2, 3, 5, 6, 7 rebuilds exactly the path 3 → 6 → 5 and shares the rest.
// Pointer identity is the only way to check that, and it is what makes the
// diagram a statement about the code rather than a drawing.
func TestTreeSharesExactlyWhatTheDiagramClaims(t *testing.T) {
	before := TreeOf(3, 1, 6, 2, 5, 7)
	checkTree(t, before, 6)

	// The shape the diagram draws, verified before relying on it.
	if before.root.value != 3 || before.root.left.value != 1 || before.root.right.value != 6 {
		t.Fatalf("the tree is not shaped as the diagram says: root=%v left=%v right=%v",
			before.root.value, before.root.left.value, before.root.right.value)
	}

	after := before.Insert(4)
	checkTree(t, after, 7)

	// 4 sits to the left of 5, under 6 — not on the 1 side.
	if got := after.root.right.left.left; got == nil || got.value != 4 {
		t.Fatalf("4 is not where the diagram puts it; found %v", got)
	}

	// Rebuilt: every node on the path from the root down to 4.
	for _, rebuilt := range []struct {
		name     string
		was, now *treeNode[int]
	}{
		{"root 3", before.root, after.root},
		{"node 6", before.root.right, after.root.right},
		{"node 5", before.root.right.left, after.root.right.left},
	} {
		if rebuilt.was == rebuilt.now {
			t.Errorf("%s should have been rebuilt, but it is the same node", rebuilt.name)
		}
	}

	// Shared: everything off that path, kept as the very same nodes.
	for _, shared := range []struct {
		name     string
		was, now *treeNode[int]
	}{
		{"subtree under 1", before.root.left, after.root.left},
		{"node 2", before.root.left.right, after.root.left.right},
		{"node 7", before.root.right.right, after.root.right.right},
	} {
		if shared.was != shared.now {
			t.Errorf("%s should have been shared, but it was copied", shared.name)
		}
	}

	// The other half of the diagram's claim: four nodes allocated, the three
	// rebuilt plus the new one. Measured, not asserted in prose.
	var retained Tree[int]
	allocations := testing.AllocsPerRun(50, func() { retained = before.Insert(4) })
	if allocations != 4 {
		t.Errorf("inserting 4 allocated %.0f nodes, want the 4 the diagram shows", allocations)
	}
	if retained.Len() != 7 {
		t.Fatalf("the retained tree holds %d values, want 7", retained.Len())
	}

	if got := before.ToSlice(); len(got) != 6 {
		t.Errorf("the original tree changed: %v", got)
	}
}
