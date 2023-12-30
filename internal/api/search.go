package api

import "bactic/internal"

type TrieNode struct {
	id       uint32
	children map[rune]*TrieNode
}

type SearchResult struct {
	Id   uint32
	Text string
}

func GetResults(prefix string, number int, root *TrieNode) []SearchResult {
	results := make([]SearchResult, 0, number)

	for _, c := range prefix {
		if root != nil {
			root = root.children[c]
		}
	}

	return results

}

func BuildAthleteTrie(athletes []internal.Athlete) *TrieNode {
	root := TrieNode{
		children: nil,
	}
	head := &root
	for _, ath := range athletes {
		name := ath.Name
		for _, c := range name {
			if head.children == nil {
				head.children = make(map[rune]*TrieNode)
				head.children[c] = &TrieNode{
					children: nil,
				}
				head = head.children[c]
			}
		}
		head.id = ath.ID
	}
	return &root
}
