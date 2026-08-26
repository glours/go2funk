// Package collection provides purely functional collections.
//
// The collections here are persistent: every operation returns a new value and
// nothing is ever mutated in place, so sharing a collection between goroutines
// or keeping an old version around is always safe.
//
// Every type integrates with the Go iterator protocol through All, which
// returns an iter.Seq usable directly in a for range loop.
//
// # What persistence costs, and why it does not cost much
//
// Returning a new collection instead of mutating one sounds like it should copy
// everything. It does not: the new value shares almost all of its nodes with
// the one it came from, and only rebuilds what actually had to change.
//
// Take a tree holding 1, 2, 3, 5, 6, 7 and insert 4. It belongs to the left of
// 5, so the path walked is 3 → 6 → 5: those three nodes are rebuilt, one
// more is allocated for 4 itself, and everything off that path is shared.
//
//	before                         after
//
//	    (3)                            (3)*         * rebuilt: on the path
//	   /   \                          /   \
//	 (1)   (6)                      (1)   (6)*      (1) and (2) are shared,
//	   \   /  \                      \   /  \       so is (7): the insertion
//	   (2)(5) (7)                    (2)(5)* (7)     never went near them
//	                                    /
//	                                  (4)+          + newly allocated
//
// The old tree is still there, still correct, still usable. Nothing was
// invalidated, because nothing was written to.
//
// Four nodes were allocated, three were shared. That is why the cost of an
// insertion is the depth of the tree rather than its size, and the tests
// measure it rather than assert it: inserting into a tree of a thousand values
// allocates ten nodes, not a thousand, and the sharing above is checked by
// pointer identity.
//
// The same idea drives List, where Prepend allocates one node and shares the
// entire tail.
package collection
