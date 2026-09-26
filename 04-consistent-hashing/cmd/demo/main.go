package main

import (
	"fmt"

	"consistent-hashing/ring"
)

func main() {
	keys := make([]string, 100)
	for i := range keys {
		keys[i] = fmt.Sprintf("user:%d", i)
	}
	before := ring.New(100)
	for _, node := range []string{"node-a", "node-b", "node-c"} {
		before.AddNode(node)
	}
	after := ring.New(100)
	for _, node := range []string{"node-a", "node-b", "node-c", "node-d"} {
		after.AddNode(node)
	}

	moved := 0
	for i, key := range keys {
		beforeNode := before.Lookup(key)
		afterNode := after.Lookup(key)
		if beforeNode != afterNode {
			moved++
		}
		if i < 12 {
			fmt.Printf("%-7s: %s -> %s\n", key, beforeNode, afterNode)
		}
	}

	fmt.Printf("hash ring moved=%d/%d keys\n", moved, len(keys))

	removed := ring.New(100)
	for _, node := range []string{"node-a", "node-b", "node-c"} {
		removed.AddNode(node)
	}
	removed.RemoveNode("node-b")

	moved = 0
	for i, key := range keys {
		beforeNode := before.Lookup(key)
		afterNode := removed.Lookup(key)
		if beforeNode != afterNode {
			moved++
		}
		if i < 12 {
			fmt.Printf("remove %-7s: %s -> %s\n", key, beforeNode, afterNode)
		}
	}
	fmt.Printf("hash ring removal moved=%d/%d keys\n", moved, len(keys))
}
