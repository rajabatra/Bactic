package internal

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type node struct {
	children map[rune]*node
	values   []SearchItem
}

type Trie struct {
	root          *node
	normalization bool
	casesensitive bool
}

func NewTrie() *Trie {
	return &Trie{
		root: &node{
			children: make(map[rune]*node),
			values:   make([]SearchItem, 0),
		},
		normalization: false,
		casesensitive: true,
	}
}

func (t *Trie) WithNorm() {
	t.normalization = true
}

func (t *Trie) WithoutNorm() {
	t.normalization = false
}

func (t *Trie) CaseSensitive() {
	t.casesensitive = true
}

func (t *Trie) CaseInsensitive() {
	t.casesensitive = false
}

func (t *Trie) Insert(entries ...SearchItem) {
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	for _, entry := range entries {
		name := entry.Name
		if len(name) == 0 {
			continue
		}

		if t.normalization {
			normal, _, err := transform.String(transformer, name)
			if err != nil {
				panic(err) // TODO: determine behavior
			}
			name = normal
		}

		if !t.casesensitive {
			name = strings.ToLower(name)
		}

		currentNode := t.root
		for i, c := range name {
			child, ok := currentNode.children[c]
			if !ok {
				child = new(node)
				child.children = make(map[rune]*node)
				if i == len(name)-len(string(c)) {
					child.values = append(child.values, entry)
				}
				currentNode.children[c] = child
			}
			currentNode = child
		}
	}
}

func (t *Trie) Search(query string, count int) []SearchItem {
	results := make([]SearchItem, 0, count)
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if t.normalization {
		var err error
		query, _, err = transform.String(transformer, query)
		if err != nil {
			return results // TODO: determine behavior
		}
	}

	if !t.casesensitive {
		query = strings.ToLower(query)
	}

	currentNode := t.root
	for _, c := range query {
		child, found := currentNode.children[c]
		if !found {
			return results
		}
		currentNode = child
	}

	i := 0

	// dfs through our nodes until we have filled our results buffer
	var dfs func(n *node)
	dfs = func(n *node) {
		if i == count {
			return
		} else if len(n.children) == 0 {
			results = append(results, n.values[:min(count-i, len(n.values))]...)
		} else {
			for _, child := range n.children {
				dfs(child)
			}
		}
		i += min(count-i, len(n.values))
	}

	dfs(currentNode)

	return results
}
