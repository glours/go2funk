package collection

// The trie is unexported, so its tests are internal by necessity. They pass the
// hash explicitly wherever the implementation does, which is what makes the
// collision paths reachable: forcing two different keys to share a hash is
// impossible through maphash, and trivial here.

import (
	"fmt"
	"math"
	"math/bits"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"
)

func TestChampPutAndGet(t *testing.T) {
	var root *champNode[string, int]
	var grew bool

	root, grew = champPut(root, "one", 1, hashKey("one"), 0)
	if !grew {
		t.Error("putting a new key must report growth")
	}
	root, grew = champPut(root, "two", 2, hashKey("two"), 0)
	if !grew {
		t.Error("putting a second key must report growth")
	}

	for key, want := range map[string]int{"one": 1, "two": 2} {
		if got, ok := champGet(root, key, hashKey(key), 0); !ok || got != want {
			t.Errorf("get(%q) = (%d, %v), want (%d, true)", key, got, ok, want)
		}
	}
	if _, ok := champGet(root, "three", hashKey("three"), 0); ok {
		t.Error("get of an absent key must report false")
	}
	if _, ok := champGet[string, int](nil, "one", hashKey("one"), 0); ok {
		t.Error("get on an empty trie must report false")
	}
}

func TestChampPutReplaces(t *testing.T) {
	var root *champNode[string, int]
	root, _ = champPut(root, "one", 1, hashKey("one"), 0)

	root, grew := champPut(root, "one", 11, hashKey("one"), 0)
	if grew {
		t.Error("replacing a value must not report growth")
	}
	if got, _ := champGet(root, "one", hashKey("one"), 0); got != 11 {
		t.Errorf("get = %d, want 11", got)
	}
	if got := champCount(root); got != 1 {
		t.Errorf("count = %d, want 1", got)
	}
}

func TestChampRemove(t *testing.T) {
	var root *champNode[string, int]
	for index, key := range []string{"one", "two", "three", "four"} {
		root, _ = champPut(root, key, index, hashKey(key), 0)
	}

	root, shrank := champRemove(root, "two", hashKey("two"), 0)
	if !shrank {
		t.Error("removing a present key must report shrinkage")
	}
	if _, ok := champGet(root, "two", hashKey("two"), 0); ok {
		t.Error("the key is still there after remove")
	}
	if got := champCount(root); got != 3 {
		t.Errorf("count = %d, want 3", got)
	}

	if _, shrank := champRemove(root, "nope", hashKey("nope"), 0); shrank {
		t.Error("removing an absent key must not report shrinkage")
	}
	if _, shrank := champRemove[string, int](nil, "one", hashKey("one"), 0); shrank {
		t.Error("removing from an empty trie must not report shrinkage")
	}
}

func TestChampEmptiesCompletely(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	var root *champNode[string, int]
	for index, key := range keys {
		root, _ = champPut(root, key, index, hashKey(key), 0)
	}
	for _, key := range keys {
		root, _ = champRemove(root, key, hashKey(key), 0)
	}
	if root != nil {
		t.Errorf("the trie must be nil once emptied, got %+v", root)
	}
	if got := champCount(root); got != 0 {
		t.Errorf("count = %d, want 0", got)
	}
}

// Two keys sharing a full hash must both be kept, found and removable.
func TestChampHandlesFullHashCollisions(t *testing.T) {
	const shared uint64 = 0xdeadbeefcafebabe

	var root *champNode[string, int]
	root, _ = champPut(root, "first", 1, shared, 0)
	root, grew := champPut(root, "second", 2, shared, 0)
	if !grew {
		t.Fatal("a colliding key must still be added")
	}
	if got := champCount(root); got != 2 {
		t.Fatalf("count = %d, want 2", got)
	}

	for key, want := range map[string]int{"first": 1, "second": 2} {
		if got, ok := champGet(root, key, shared, 0); !ok || got != want {
			t.Errorf("get(%q) = (%d, %v), want (%d, true)", key, got, ok, want)
		}
	}
	if _, ok := champGet(root, "third", shared, 0); ok {
		t.Error("a key that only shares the hash must not be found")
	}

	// Replacing inside a collision node keeps the count.
	root, grew = champPut(root, "first", 11, shared, 0)
	if grew || champCount(root) != 2 {
		t.Errorf("replacing in a collision node reported growth=%v count=%d", grew, champCount(root))
	}

	root, _ = champRemove(root, "first", shared, 0)
	if got := champCount(root); got != 1 {
		t.Errorf("count = %d, want 1", got)
	}
	root, _ = champRemove(root, "second", shared, 0)
	if root != nil {
		t.Error("the trie must be nil once the collision node is emptied")
	}
}

// A collision reached deep in the trie, below several levels of branching.
func TestChampCollisionUnderCommonPrefix(t *testing.T) {
	// Same low bits for many levels, so the walk descends before colliding.
	const a uint64 = 0x0000000000000001
	const b uint64 = 0x8000000000000001

	var root *champNode[string, int]
	root, _ = champPut(root, "a", 1, a, 0)
	root, _ = champPut(root, "b", 2, b, 0)
	root, _ = champPut(root, "c", 3, a, 0) // collides with a

	if got := champCount(root); got != 3 {
		t.Fatalf("count = %d, want 3", got)
	}
	for _, tc := range []struct {
		key  string
		hash uint64
		want int
	}{{"a", a, 1}, {"b", b, 2}, {"c", a, 3}} {
		if got, ok := champGet(root, tc.key, tc.hash, 0); !ok || got != tc.want {
			t.Errorf("get(%q) = (%d, %v), want (%d, true)", tc.key, got, ok, tc.want)
		}
	}
}

func TestChampIteration(t *testing.T) {
	var root *champNode[int, string]
	for value := range 100 {
		root, _ = champPut(root, value, fmt.Sprint(value), hashKey(value), 0)
	}

	seen := map[int]string{}
	champAll(root, func(key int, value string) bool {
		seen[key] = value
		return true
	})
	if len(seen) != 100 {
		t.Errorf("iteration saw %d entries, want 100", len(seen))
	}
	for key, value := range seen {
		if value != fmt.Sprint(key) {
			t.Errorf("key %d carried %q", key, value)
		}
	}

	// A consumer that stops must stop the walk.
	count := 0
	champAll(root, func(int, string) bool {
		count++
		return false
	})
	if count != 1 {
		t.Errorf("the walk yielded %d entries after a stop, want 1", count)
	}
}

// The trie must agree with Go's own map over a long random run.
func TestChampAgreesWithABuiltinMap(t *testing.T) {
	random := rand.New(rand.NewPCG(7, 11)) // fixed seed: failures reproduce
	reference := map[int]int{}
	var root *champNode[int, int]

	for step := range 20000 {
		key := random.IntN(2000)
		switch random.IntN(3) {
		case 0, 1:
			value := random.Int()
			_, existed := reference[key]
			reference[key] = value

			var grew bool
			root, grew = champPut(root, key, value, hashKey(key), 0)
			if grew == existed {
				t.Fatalf("step %d: put reported grew=%v for a key that existed=%v", step, grew, existed)
			}
		default:
			_, existed := reference[key]
			delete(reference, key)

			var shrank bool
			root, shrank = champRemove(root, key, hashKey(key), 0)
			if shrank != existed {
				t.Fatalf("step %d: remove reported shrank=%v for a key that existed=%v", step, shrank, existed)
			}
		}

		if got := champCount(root); got != len(reference) {
			t.Fatalf("step %d: count = %d, want %d", step, got, len(reference))
		}
	}

	for key, want := range reference {
		if got, ok := champGet(root, key, hashKey(key), 0); !ok || got != want {
			t.Fatalf("get(%d) = (%d, %v), want (%d, true)", key, got, ok, want)
		}
	}

	var keys []int
	champAll(root, func(key int, _ int) bool {
		keys = append(keys, key)
		return true
	})
	slices.Sort(keys)
	if len(keys) != len(reference) {
		t.Fatalf("iteration saw %d keys, want %d", len(keys), len(reference))
	}
}

// Adding an entry rebuilds only the nodes on its path.
func TestChampSharesStructure(t *testing.T) {
	var root *champNode[int, int]
	for value := range 1000 {
		root, _ = champPut(root, value*2, value, hashKey(value*2), 0)
	}

	var kept *champNode[int, int]
	allocations := testing.AllocsPerRun(50, func() {
		kept, _ = champPut(root, 1, 1, hashKey(1), 0)
	})
	if allocations > 20 {
		t.Errorf("adding one entry to a trie of 1000 allocated %.0f times, want a shallow count", allocations)
	}
	t.Logf("adding one entry to a trie of 1000 allocates %.0f times", allocations)

	if champCount(kept) != 1001 || champCount(root) != 1000 {
		t.Errorf("counts drifted: new=%d old=%d", champCount(kept), champCount(root))
	}
}

// checkChampNode verifies the shape the type comment claims:
//   - the two bitmaps never mark the same slot,
//   - each bitmap's popcount matches the length of its slice,
//   - the cached size counts the pairs actually below,
//   - below the root, no node holds fewer than two pairs.
//
// It returns the number of pairs it counted.
func checkChampNode[K comparable, V any](t *testing.T, node *champNode[K, V], isRoot bool) int {
	t.Helper()
	if node == nil {
		return 0
	}

	if node.isCollision() {
		if !isRoot && len(node.entries) < 2 {
			t.Fatalf("a collision node below the root holds %d pairs, want at least 2", len(node.entries))
		}
		for _, entry := range node.entries {
			if entry.hash != node.entries[0].hash {
				t.Fatal("a collision node holds entries with different hashes")
			}
		}
		if node.size != len(node.entries) {
			t.Fatalf("cached size %d, want %d", node.size, len(node.entries))
		}
		return node.size
	}

	if node.dataMap&node.nodeMap != 0 {
		t.Fatalf("a slot is marked in both bitmaps: data=%032b node=%032b", node.dataMap, node.nodeMap)
	}
	if got, want := len(node.entries), bits.OnesCount32(node.dataMap); got != want {
		t.Fatalf("%d entries for %d bits set in dataMap", got, want)
	}
	if got, want := len(node.nodes), bits.OnesCount32(node.nodeMap); got != want {
		t.Fatalf("%d subtrees for %d bits set in nodeMap", got, want)
	}

	pairs := len(node.entries)
	for _, subtree := range node.nodes {
		pairs += checkChampNode(t, subtree, false)
	}
	if node.size != pairs {
		t.Fatalf("cached size %d, want %d", node.size, pairs)
	}
	if !isRoot && pairs < 2 {
		t.Fatalf("a node below the root holds %d pairs, want at least 2", pairs)
	}
	return pairs
}

// The canonical shape must survive thousands of random insertions and removals.
func TestChampStaysCanonical(t *testing.T) {
	random := rand.New(rand.NewPCG(13, 17))
	reference := map[int]bool{}
	var root *champNode[int, int]

	for range 5000 {
		key := random.IntN(400)
		if random.IntN(2) == 0 {
			root, _ = champPut(root, key, key, hashKey(key), 0)
			reference[key] = true
		} else {
			root, _ = champRemove(root, key, hashKey(key), 0)
			delete(reference, key)
		}
		if got := checkChampNode(t, root, true); got != len(reference) {
			t.Fatalf("the trie holds %d pairs, want %d", got, len(reference))
		}
	}
}

// A slot can hold a key whose hash starts the same way but is not the one asked
// for. There is nowhere else to look, so the lookup stops there.
func TestChampSlotHoldsAnotherKey(t *testing.T) {
	const sameFirstLevel uint64 = 0b00001
	const alsoSameFirstLevel uint64 = 0b10000_00001

	var root *champNode[string, int]
	root, _ = champPut(root, "present", 1, sameFirstLevel, 0)

	if _, ok := champGet(root, "absent", sameFirstLevel, 0); ok {
		t.Error("a key sharing the hash but not the identity must not be found")
	}
	if _, ok := champGet(root, "absent", alsoSameFirstLevel, 0); ok {
		t.Error("a key reaching the same slot must not be found either")
	}
}

func TestChampCollisionNodeMisses(t *testing.T) {
	const shared uint64 = 0x1234567890abcdef

	var root *champNode[string, int]
	root, _ = champPut(root, "a", 1, shared, 0)
	root, _ = champPut(root, "b", 2, shared, 0)

	// A third colliding key goes in beside the other two.
	root, grew := champPut(root, "c", 3, shared, 0)
	if !grew || champCount(root) != 3 {
		t.Fatalf("adding a third colliding key: grew=%v count=%d", grew, champCount(root))
	}

	// Removing a key that is not in the collision node changes nothing.
	after, shrank := champRemove(root, "d", shared, 0)
	if shrank {
		t.Error("removing an absent key from a collision node must not shrink")
	}
	if after != root {
		t.Error("a removal that changes nothing must hand back the same node")
	}
}

func TestChampWalkStops(t *testing.T) {
	// An empty trie yields nothing and reports that it finished.
	if !champAll[int, int](nil, func(int, int) bool { return true }) {
		t.Error("walking an empty trie must report completion")
	}

	// Enough keys to guarantee subtrees, so the stop has to propagate up.
	var root *champNode[int, int]
	for key := range 500 {
		root, _ = champPut(root, key, key, hashKey(key), 0)
	}

	for _, stopAfter := range []int{1, 2, 17, 64} {
		seen := 0
		finished := champAll(root, func(int, int) bool {
			seen++
			return seen < stopAfter
		})
		if finished {
			t.Errorf("stopping after %d reported completion", stopAfter)
		}
		if seen != stopAfter {
			t.Errorf("the walk yielded %d pairs, want %d", seen, stopAfter)
		}
	}
}

// A NaN key is unreachable here, as it is in a Go map, and findable in Tree, as
// it is with slices.BinarySearch. This pins both sides of that split so neither
// can drift into the other by accident.
func TestChampMatchesGoMapsOnNaN(t *testing.T) {
	nan := math.NaN()

	// Two calls, same argument, different results: that is the whole premise.
	first, second := hashKey(nan), hashKey(nan)
	if first == second {
		t.Fatal("hashing a NaN twice gave the same value: the premise no longer holds")
	}

	var root *champNode[float64, int]
	for value := range 5 {
		var grew bool
		root, grew = champPut(root, nan, value, hashKey(nan), 0)
		if !grew {
			t.Error("every NaN put lands as a new entry, as it does in a Go map")
		}
	}

	reference := map[float64]int{}
	for value := range 5 {
		reference[nan] = value
	}
	if got := champCount(root); got != len(reference) {
		t.Errorf("the trie holds %d NaN entries, the builtin map holds %d", got, len(reference))
	}

	if _, ok := champGet(root, nan, hashKey(nan), 0); ok {
		t.Error("a NaN entry must not be readable, as in a Go map")
	}
	if _, shrank := champRemove(root, nan, hashKey(nan), 0); shrank {
		t.Error("a NaN entry must not be removable, as in a Go map")
	}

	// The ordered side of the split: Tree holds the NaN once and finds it,
	// exactly as slices.Sort and slices.BinarySearch do.
	tree := TreeOf(nan, nan, nan)
	if got := tree.Len(); got != 1 {
		t.Errorf("Tree holds %d NaN values, want 1: it orders through cmp.Compare", got)
	}
	if !tree.Contains(nan) {
		t.Error("Tree must find a NaN, the way slices.BinarySearch does")
	}

	// The two reference points, so the claim is not just prose.
	sorted := []float64{3, nan, 1}
	slices.Sort(sorted)
	if _, found := slices.BinarySearch(sorted, nan); !found {
		t.Error("slices.BinarySearch no longer finds a NaN: the premise changed")
	}
	if slices.Contains(sorted, nan) {
		t.Error("slices.Contains now finds a NaN: the premise changed")
	}
}

// The last level is the only one that branches on fewer than champBits of hash:
// bits 60 to 63 are all that remain. Nothing above exercises it, because every
// other collision test uses hashes that are equal outright and therefore descend
// the chain without ever branching there.
//
// These sixteen hashes agree on bits 0 to 59 and differ only in the top four, so
// they must all land side by side in one node at shift 60 — not in a collision
// node, which is what an off-by-one in either champMaxShift or the comparison
// against it would produce.
func TestChampBranchesOnTheLastHashBits(t *testing.T) {
	const shared = 0x0fedcba987654321 & (1<<60 - 1)

	hashOf := func(index int) uint64 { return uint64(index)<<60 | shared }

	var root *champNode[int, int]
	for index := range 16 {
		var grew bool
		root, grew = champPut(root, index, index, hashOf(index), 0)
		if !grew {
			t.Fatalf("key %d was treated as already present", index)
		}
	}

	if got := champCount(root); got != 16 {
		t.Fatalf("count = %d, want 16", got)
	}
	for index := range 16 {
		if got, ok := champGet(root, index, hashOf(index), 0); !ok || got != index {
			t.Errorf("get(%d) = (%d, %v), want (%d, true)", index, got, ok, index)
		}
	}
	checkChampNode(t, root, true)

	// They differ, so they must have branched rather than piled into a
	// collision node. Walking down finds the node where they part.
	node, shift := root, uint(0)
	for len(node.nodes) == 1 && len(node.entries) == 0 {
		node = node.nodes[0]
		shift += champBits
	}
	if node.isCollision() {
		t.Fatalf("keys with different hashes ended up in a collision node at shift %d", shift)
	}
	if shift != champMaxShift {
		t.Errorf("they parted at shift %d, want %d: the last level is the one with four bits left",
			shift, champMaxShift)
	}
	if got := len(node.entries); got != 16 {
		t.Errorf("the last level holds %d entries, want 16", got)
	}

	// Removing them one by one must drain the trie cleanly.
	for index := range 16 {
		var shrank bool
		root, shrank = champRemove(root, index, hashOf(index), 0)
		if !shrank {
			t.Fatalf("removing key %d reported no shrinkage", index)
		}
	}
	if root != nil {
		t.Error("the trie must be nil once drained")
	}
}

// Two keys with the same hash cannot be separated by any prefix of it, so the
// trie stretches one node per level down to the collision node that holds them.
// The canonical invariant bounds how many pairs a node carries, not the shape
// above them, and this pins that distinction so the type comment stays honest.
func TestChampChainsDownToACollisionNode(t *testing.T) {
	const shared uint64 = 0xabcdef0123456789

	var root *champNode[string, int]
	root, _ = champPut(root, "a", 1, shared, 0)
	root, _ = champPut(root, "b", 2, shared, 0)

	depth, singleChild := 1, 0
	for node := root; ; depth++ {
		if node.isCollision() {
			break
		}
		if len(node.entries) == 0 && len(node.nodes) == 1 {
			singleChild++
		}
		if len(node.nodes) != 1 {
			t.Fatalf("at depth %d the chain forked, which two equal hashes cannot do", depth)
		}
		node = node.nodes[0]
	}

	if want := 13; singleChild != want {
		t.Errorf("the chain holds %d single-child nodes, want %d — one per level of hash bits",
			singleChild, want)
	}
	checkChampNode(t, root, true)
	if got := champCount(root); got != 2 {
		t.Errorf("count = %d, want 2", got)
	}
}

// Mapping the values keeps the trie's shape: same bitmaps, same hashes, so the
// result needs no rehash and stays canonical. A collision node is included,
// since only a test that forces the hash can build one.
func TestChampMapValuesKeepsTheShape(t *testing.T) {
	const shared uint64 = 0xdeadbeefcafebabe

	var root *champNode[string, int]
	for value := range 200 {
		key := fmt.Sprint(value)
		root, _ = champPut(root, key, value, hashKey(key), 0)
	}
	root, _ = champPut(root, "first", 1000, shared, 0)
	root, _ = champPut(root, "second", 2000, shared, 0)

	mapped := champMapValues(root, func(value int) string { return fmt.Sprint(value * 2) })

	checkChampNode(t, mapped, true)
	if champCount(mapped) != champCount(root) {
		t.Fatalf("count = %d, want %d", champCount(mapped), champCount(root))
	}
	for value := range 200 {
		key := fmt.Sprint(value)
		if got, ok := champGet(mapped, key, hashKey(key), 0); !ok || got != fmt.Sprint(value*2) {
			t.Fatalf("get(%q) = (%q, %v), want (%q, true)", key, got, ok, fmt.Sprint(value*2))
		}
	}

	// The colliding keys are still reachable, and still removable, by the hash
	// they were stored under: mapping carried it over rather than dropping it.
	for key, want := range map[string]string{"first": "2000", "second": "4000"} {
		if got, ok := champGet(mapped, key, shared, 0); !ok || got != want {
			t.Errorf("get(%q) = (%q, %v), want (%q, true)", key, got, ok, want)
		}
	}
	shrunk, removed := champRemove(mapped, "first", shared, 0)
	if !removed {
		t.Fatal("a colliding key of the mapped trie could not be removed")
	}
	checkChampNode(t, shrunk, true)

	if got, _ := champGet(root, "first", shared, 0); got != 1000 {
		t.Errorf("the original trie was modified: get(first) = %d, want 1000", got)
	}
	if champMapValues[string, int, string](nil, strconv.Itoa) != nil {
		t.Error("mapping the empty trie gives the empty trie")
	}
}
