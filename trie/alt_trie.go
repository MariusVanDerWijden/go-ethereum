package trie

type altTrie struct {
	root *stNode
	h    *hasher
	kBuf []byte
	pBuf []byte
}

func (t *altTrie) Update(key, value []byte) error {
	t.grow(key)
	k := writeHexKey(t.kBuf, key)
	if len(value) == 0 {
		t.delete(t.root, k)
	}
	t.insert(t.root, k, value, t.pBuf[:0])
	return nil
}

func (t *altTrie) insert(st *stNode, key, value []byte, path []byte) (bool, error) {
	switch st.typ {
	case branchNode:
		idx := int(key[0])
		// Add new child
		if st.children[idx] == nil {
			st.children[idx] = newLeaf(key[1:], value)
		} else {
			return t.insert(st.children[idx], key[1:], value, append(path, key[0]))
		}
	case extNode:
		diffidx := st.getDiffIndex(key)
		// Check if chunks are identical. If so, recurse into
		// the child node. Otherwise, the key has to be split
		// into 1) an optional common prefix, 2) the fullnode
		// representing the two differing path, and 3) a leaf
		// for each of the differentiated subtrees.
		if diffidx == len(st.key) {
			// Ext key and key segment are identical, recurse into
			// the child node.
			return t.insert(st.children[0], key[diffidx:], value, append(path, key[:diffidx]...))
		}
		// Save the original part. Depending if the break is
		// at the extension's last byte or not, create an
		// intermediate extension or use the extension's child
		// node directly.
		var n *stNode
		if diffidx < len(st.key)-1 {
			// Break on the non-last byte, insert an intermediate
			// extension. The path prefix of the newly-inserted
			// extension should also contain the different byte.
			n = newExt(st.key[diffidx+1:], st.children[0])
			t.hash(n, append(path, st.key[:diffidx+1]...))
		} else {
			// Break on the last byte, no need to insert
			// an extension node: reuse the current node.
			// The path prefix of the original part should
			// still be same.
			n = st.children[0]
			t.hash(n, append(path, st.key...))
		}
		var p *stNode
		if diffidx == 0 {
			// the break is on the first byte, so
			// the current node is converted into
			// a branch node.
			st.children[0] = nil
			p = st
			st.typ = branchNode
		} else {
			// the common prefix is at least one byte
			// long, insert a new intermediate branch
			// node.
			st.children[0] = stPool.Get().(*stNode)
			st.children[0].typ = branchNode
			p = st.children[0]
		}
		// Create a leaf for the inserted part
		o := newLeaf(key[diffidx+1:], value)

		// Insert both child leaves where they belong:
		origIdx := st.key[diffidx]
		newIdx := key[diffidx]
		p.children[origIdx] = n
		p.children[newIdx] = o
		st.key = st.key[:diffidx]
	}
	return false, nil
}

func (t *altTrie) delete(st *stNode, key []byte)

func (t *altTrie) grow(key []byte) {
	if cap(t.kBuf) < 2*len(key) {
		t.kBuf = make([]byte, 2*len(key))
	}
	if cap(t.pBuf) < 2*len(key) {
		t.pBuf = make([]byte, 2*len(key))
	}
}
