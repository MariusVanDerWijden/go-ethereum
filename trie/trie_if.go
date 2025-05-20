package trie

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

type TrieIF interface {
	MustNodeIterator(start []byte) NodeIterator
	MustGet(key []byte) []byte
	MustGetNode(path []byte) ([]byte, int)
	MustUpdate(key, value []byte)
	MustDelete(key []byte)

	NodeIterator(start []byte) (NodeIterator, error)
	Get(key []byte) ([]byte, error)
	GetNode(path []byte) ([]byte, int, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	Hash() common.Hash
	Commit(collectLeaf bool) (common.Hash, *trienode.NodeSet)
	Witness() map[string]struct{}
	Reset()
	Copy() *Trie
}

var _ = TrieIF(new(Trie))
