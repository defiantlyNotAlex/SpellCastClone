package main

import (
	"log"
	"os"
	"strings"
)

const (
	A = iota
	B
	C
	D
	E
	F
	G
	H
	I
	J
	K
	L
	M
	N
	O
	P
	Q
	R
	S
	T
	U
	V
	W
	X
	Y
	Z
	WordEnd = 31
)

type TrieNode struct {
	flags   uint32
	indexes [26]uint32
}

type Trie struct {
	nodes []TrieNode
}

func GetWordList(filepath string) []string {
	text, err := os.ReadFile(filepath)
	if err != nil {
		log.Panicf("could not read file %s: %v", filepath, err)
	}
	return strings.Split(string(text), "\r\n")
}

func NewNode() TrieNode {
	return TrieNode{
		flags:   0,
		indexes: [26]uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func GetNode(trie *Trie, index int) *TrieNode {
	return &trie.nodes[index]
}

func BuildTrie(wordlist []string) Trie {
	var trie Trie
	trie.nodes = append(trie.nodes, NewNode())

	for _, word := range wordlist {
		curr := 0
		for _, char := range word {
			node := GetNode(&trie, curr)
			i := int(char - 'a')

			if get_bit(node.flags, i) {
				curr = int(node.indexes[i])
			} else {
				node.flags = set_bit(node.flags, i)
				curr = len(trie.nodes)
				node.indexes[i] = uint32(curr)
				trie.nodes = append(trie.nodes, NewNode())
			}
		}
		node := GetNode(&trie, curr)
		node.flags = set_bit(node.flags, WordEnd)
	}
	return trie
}

func InWordList(trie *Trie, word string) bool {
	curr := 0
	for _, char := range word {
		node := GetNode(trie, curr)
		i := int(char - 'a')

		if get_bit(node.flags, i) {
			curr = int(node.indexes[i])
		} else {
			return false
		}
	}
	node := GetNode(trie, curr)
	return get_bit(node.flags, WordEnd)
}
