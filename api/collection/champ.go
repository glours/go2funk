package collection

import (
	"hash/maphash"
	"math/bits"
	"slices"
)

// hashSeed is drawn once per process. Every version of every trie shares it,
// which is what lets the zero value of a map be the empty one: there is no seed
// to carry around. Being random per process is enough to keep an attacker from
// precomputing colliding keys.
var hashSeed = maphash.MakeSeed()

// hashKey hashes any comparable value without a dependency. Structs, arrays and
// pointers all work, which is the whole reason this trie is keyed by comparable
// rather than by cmp.Ordered the way Tree is.
//
// A float NaN is not equal to itself, so maphash gives it a fresh hash every
// call and an entry keyed by one can never be read back or removed. That is what
// Go's own map does too: m[math.NaN()] = v grows it every time and none of the
// entries is reachable.
//
// Tree, in this same package, finds a NaN, and the two are not in disagreement —
// they follow the split the standard library already draws. Lookup by equality
// does not find a NaN; lookup by order does:
//
//	slices.Contains(values, math.NaN())     // false
//	slices.Index(values, math.NaN())        // -1
//	map[float64]V                           // the entry is unreachable
//
//	slices.Sort(values)                     // [NaN 1 2 3] — it has a place
//	slices.BinarySearch(values, math.NaN()) // 0, true — it is found
//	slices.Min, slices.Max                  // NaN
//
// This trie hashes and compares keys, so it lands on the first side, like the
// built-in map. Tree orders through cmp.Compare, so it lands on the second, like
// slices.BinarySearch. That is the reason cmp.Compare exists: to give a total
// order where == cannot.
func hashKey[K comparable](key K) uint64 {
	return maphash.Comparable(hashSeed, key)
}

// How the trie branches.
//
// A key's 64-bit hash is read champBits at a time, so each level picks one of
// champWidth slots. Twelve levels consume 60 bits and the thirteenth the last
// four; two keys that still collide there have the same hash outright and land
// in a collision node.
const (
	champBits     = 5
	champWidth    = 1 << champBits // 32 slots per level
	champMask     = champWidth - 1
	champMaxShift = 60 // the last shift that still has hash bits to read
)

// champEntry is one key-value pair. The hash is kept alongside so that pushing
// an entry down a level never needs to hash it again, and so that the functions
// below take the hash as an argument rather than computing it — which is what
// makes the collision paths reachable from a test.
type champEntry[K comparable, V any] struct {
	key   K
	value V
	hash  uint64
}

// champNode is a CHAMP node: two bitmaps rather than one.
//
// dataMap marks the slots holding an entry, nodeMap the slots holding a subtree,
// and the two can never mark the same slot. Each bitmap indexes its own compact
// slice: the position of a slot's payload is the number of lower bits set in the
// matching bitmap, which is one popcount. Keeping entries and subtrees apart is
// what removes the need for an interface or a type switch on the way down, and
// it lets iteration yield a node's entries without touching its subtrees.
//
// A collision node is the one exception: both bitmaps are zero and entries holds
// every pair sharing a full hash, scanned linearly. Nothing else can have
// entries with an empty dataMap, so the shape is unambiguous.
//
// The trie is kept canonical: below the root, every node holds at least two
// pairs. Subtrees are only ever born from a collision between two keys, and a
// removal that would leave one pair behind pulls it back into the parent. That
// is what stops a removal from leaving hollow nodes behind, and it is checked by
// a test rather than assumed.
//
// It bounds how many pairs a node carries, not the shape above them. Two keys
// whose hashes are equal outright still stretch a chain of single-child nodes
// down to the collision node that finally holds them — thirteen of them, one per
// level of hash bits — because nothing shorter can tell the two keys apart. That
// chain is inherent to the structure, not a lapse in the invariant.
type champNode[K comparable, V any] struct {
	dataMap uint32
	nodeMap uint32
	entries []champEntry[K, V]
	nodes   []*champNode[K, V]
	size    int
}

// champIndex reads the champBits of the hash that this level branches on.
func champIndex(hash uint64, shift uint) uint32 {
	return uint32(hash>>shift) & champMask
}

// champBit turns that index into the single bit the bitmaps use for the slot.
func champBit(hash uint64, shift uint) uint32 {
	return 1 << champIndex(hash, shift)
}

// champPos gives a slot's position in its compact slice: how many slots below it
// are occupied. This is the popcount that replaces a 32-wide array per node.
func champPos(bitmap, bit uint32) int {
	return bits.OnesCount32(bitmap & (bit - 1))
}

func (n *champNode[K, V]) isCollision() bool {
	return n.dataMap == 0 && n.nodeMap == 0 && len(n.entries) > 0
}

// newChampNode builds a node and caches the number of pairs beneath it.
func newChampNode[K comparable, V any](
	dataMap, nodeMap uint32,
	entries []champEntry[K, V],
	nodes []*champNode[K, V],
) *champNode[K, V] {
	size := len(entries)
	for _, sub := range nodes {
		size += sub.size
	}
	return &champNode[K, V]{dataMap: dataMap, nodeMap: nodeMap, entries: entries, nodes: nodes, size: size}
}

func newChampCollision[K comparable, V any](entries []champEntry[K, V]) *champNode[K, V] {
	return &champNode[K, V]{entries: entries, size: len(entries)}
}

// champCount returns the number of pairs. O(1).
func champCount[K comparable, V any](node *champNode[K, V]) int {
	if node == nil {
		return 0
	}
	return node.size
}

// champGet walks down one level per champBits of hash. O(log32 n).
func champGet[K comparable, V any](node *champNode[K, V], key K, hash uint64, shift uint) (V, bool) {
	var zero V
	for node != nil {
		if node.isCollision() {
			for _, entry := range node.entries {
				if entry.key == key {
					return entry.value, true
				}
			}
			return zero, false
		}

		bit := champBit(hash, shift)
		switch {
		case node.dataMap&bit != 0:
			entry := node.entries[champPos(node.dataMap, bit)]
			if entry.key == key {
				return entry.value, true
			}
			return zero, false // the slot holds a different key: nowhere else to look
		case node.nodeMap&bit != 0:
			node = node.nodes[champPos(node.nodeMap, bit)]
			shift += champBits
		default:
			return zero, false
		}
	}
	return zero, false
}

// champPut returns the trie holding key, and whether it grew by one pair.
// Only the nodes on the path are rebuilt; everything else is shared.
func champPut[K comparable, V any](
	node *champNode[K, V], key K, value V, hash uint64, shift uint,
) (*champNode[K, V], bool) {
	added := champEntry[K, V]{key: key, value: value, hash: hash}

	if node == nil {
		return newChampNode(champBit(hash, shift), 0, []champEntry[K, V]{added}, nil), true
	}

	if node.isCollision() {
		for index, entry := range node.entries {
			if entry.key == key {
				entries := slices.Clone(node.entries)
				entries[index] = added
				return newChampCollision(entries), false
			}
		}
		return newChampCollision(append(slices.Clone(node.entries), added)), true
	}

	bit := champBit(hash, shift)
	switch {
	case node.dataMap&bit != 0:
		position := champPos(node.dataMap, bit)
		existing := node.entries[position]

		if existing.key == key {
			entries := slices.Clone(node.entries)
			entries[position] = added
			return newChampNode(node.dataMap, node.nodeMap, entries, node.nodes), false
		}

		// Two keys want the same slot: push both into a subtree one level down.
		subtree := champMerge(existing, added, shift+champBits)
		entries := slices.Delete(slices.Clone(node.entries), position, position+1)
		nodes := slices.Insert(slices.Clone(node.nodes), champPos(node.nodeMap, bit), subtree)
		return newChampNode(node.dataMap^bit, node.nodeMap|bit, entries, nodes), true

	case node.nodeMap&bit != 0:
		position := champPos(node.nodeMap, bit)
		subtree, grew := champPut(node.nodes[position], key, value, hash, shift+champBits)
		nodes := slices.Clone(node.nodes)
		nodes[position] = subtree
		return newChampNode(node.dataMap, node.nodeMap, node.entries, nodes), grew

	default:
		// The slot is free, so the entry lives here rather than in a subtree.
		entries := slices.Insert(slices.Clone(node.entries), champPos(node.dataMap, bit), added)
		return newChampNode(node.dataMap|bit, node.nodeMap, entries, node.nodes), true
	}
}

// champMerge builds the subtree holding two entries that collided at the level
// above. It keeps descending while their hashes agree, and gives up into a
// collision node once there are no hash bits left to tell them apart.
func champMerge[K comparable, V any](first, second champEntry[K, V], shift uint) *champNode[K, V] {
	if shift > champMaxShift {
		return newChampCollision([]champEntry[K, V]{first, second})
	}

	firstBit, secondBit := champBit(first.hash, shift), champBit(second.hash, shift)
	if firstBit == secondBit {
		subtree := champMerge(first, second, shift+champBits)
		return newChampNode(0, firstBit, nil, []*champNode[K, V]{subtree})
	}

	entries := []champEntry[K, V]{first, second}
	if secondBit < firstBit {
		entries[0], entries[1] = second, first
	}
	return newChampNode(firstBit|secondBit, 0, entries, nil)
}

// champRemove returns the trie without key, and whether it shrank by one pair.
func champRemove[K comparable, V any](
	node *champNode[K, V], key K, hash uint64, shift uint,
) (*champNode[K, V], bool) {
	if node == nil {
		return nil, false
	}

	if node.isCollision() {
		for index, entry := range node.entries {
			if entry.key == key {
				// The canonical invariant keeps a collision node at two pairs or
				// more, so deleting one always leaves at least one behind and this
				// never yields an empty node.
				return newChampCollision(slices.Delete(slices.Clone(node.entries), index, index+1)), true
			}
		}
		return node, false
	}

	bit := champBit(hash, shift)
	switch {
	case node.dataMap&bit != 0:
		position := champPos(node.dataMap, bit)
		if node.entries[position].key != key {
			return node, false
		}
		if len(node.entries) == 1 && len(node.nodes) == 0 {
			return nil, true // the node held nothing else
		}
		entries := slices.Delete(slices.Clone(node.entries), position, position+1)
		return newChampNode(node.dataMap^bit, node.nodeMap, entries, node.nodes), true

	case node.nodeMap&bit != 0:
		position := champPos(node.nodeMap, bit)
		subtree, shrank := champRemove(node.nodes[position], key, hash, shift+champBits)
		if !shrank {
			return node, false
		}

		// A subtree down to a single pair is pulled back up. That is what keeps
		// the trie canonical: below the root, every node holds at least two
		// pairs, so a subtree can never empty out. subtree is therefore never
		// nil here, which is the invariant checkChampNode verifies rather than
		// this code defending against.
		if lone, ok := champLoneEntry(subtree); ok {
			nodes := slices.Delete(slices.Clone(node.nodes), position, position+1)
			entries := slices.Insert(slices.Clone(node.entries), champPos(node.dataMap, bit), lone)
			return newChampNode(node.dataMap|bit, node.nodeMap^bit, entries, nodes), true
		}

		nodes := slices.Clone(node.nodes)
		nodes[position] = subtree
		return newChampNode(node.dataMap, node.nodeMap, node.entries, nodes), true

	default:
		return node, false
	}
}

// champLoneEntry reports the single pair of a node that holds nothing else.
func champLoneEntry[K comparable, V any](node *champNode[K, V]) (champEntry[K, V], bool) {
	if len(node.entries) == 1 && len(node.nodes) == 0 {
		return node.entries[0], true
	}
	var zero champEntry[K, V]
	return zero, false
}

// champAll yields every pair, reporting false as soon as the consumer stops.
// A node's own entries come first, then its subtrees — no type test on the way,
// which is the other thing the two bitmaps buy.
func champAll[K comparable, V any](node *champNode[K, V], yield func(K, V) bool) bool {
	if node == nil {
		return true
	}
	for _, entry := range node.entries {
		if !yield(entry.key, entry.value) {
			return false
		}
	}
	for _, subtree := range node.nodes {
		if !champAll(subtree, yield) {
			return false
		}
	}
	return true
}

// champMapValues rebuilds the trie with every value mapped. Keys and their
// hashes are untouched, so both bitmaps carry over as they are: no key is hashed
// again and no node changes shape, which keeps the result canonical. O(n).
func champMapValues[K comparable, V, U any](node *champNode[K, V], mapper func(V) U) *champNode[K, U] {
	if node == nil {
		return nil
	}
	entries := make([]champEntry[K, U], len(node.entries))
	for index, entry := range node.entries {
		entries[index] = champEntry[K, U]{key: entry.key, value: mapper(entry.value), hash: entry.hash}
	}
	nodes := make([]*champNode[K, U], len(node.nodes))
	for index, subtree := range node.nodes {
		nodes[index] = champMapValues(subtree, mapper)
	}
	return &champNode[K, U]{dataMap: node.dataMap, nodeMap: node.nodeMap, entries: entries, nodes: nodes, size: node.size}
}
