package main

import (
	"fmt"
	"sort"

	"sharding-router/shard"
)

func main() {
	router := shard.New([]string{"shard-a", "shard-b", "shard-c"})
	store := shard.NewMap(router)
	keys := []string{"user-1", "user-2", "user-3", "order-1", "order-2", "cart-1", "cart-2", "profile-1"}

	for _, key := range keys {
		if err := store.Put(key, "value-"+key); err != nil {
			panic(err)
		}
	}

	value, found, owner, err := store.Get("user-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("stored user-1: value=%q found=%t shard=%s\n", value, found, owner)

	before := routeAll(router, keys)
	fmt.Println("before adding shard-d:")
	printRoutes(before)

	if err := router.AddShard("shard-d"); err != nil {
		panic(err)
	}
	afterAdd := routeAll(router, keys)
	moved, err := store.Rebalance()
	if err != nil {
		panic(err)
	}
	fmt.Printf("after adding shard-d: routes_moved=%d/%d records_rebalanced=%d\n", countMoved(before, afterAdd), len(keys), moved)
	printRoutes(afterAdd)

	value, found, owner, err = store.Get("user-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("user-1 after add: value=%q found=%t shard=%s\n", value, found, owner)

	if err := router.RemoveShard("shard-d"); err != nil {
		panic(err)
	}
	afterRemove := routeAll(router, keys)
	moved, err = store.Rebalance()
	if err != nil {
		panic(err)
	}
	fmt.Printf("after removing shard-d: routes_moved=%d/%d records_rebalanced=%d\n", countMoved(afterAdd, afterRemove), len(keys), moved)

	rangeRouter, err := shard.NewRangeRouter([]shard.KeyRange{
		{Start: "", End: "m", Shard: "range-a"},
		{Start: "m", End: "t", Shard: "range-b"},
		{Start: "t", End: "", Shard: "range-c"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("range routing:")
	for _, key := range []string{"apple", "profile", "user", "zebra"} {
		owner, err := rangeRouter.Route(key)
		if err != nil {
			panic(err)
		}
		fmt.Printf("key=%s shard=%s\n", key, owner)
	}
}

func routeAll(router *shard.Router, keys []string) map[string]string {
	routes := make(map[string]string, len(keys))
	for _, key := range keys {
		owner, err := router.Route(key)
		if err != nil {
			panic(err)
		}
		routes[key] = owner
	}
	return routes
}

func printRoutes(routes map[string]string) {
	keys := make([]string, 0, len(routes))
	for key := range routes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("key=%s shard=%s\n", key, routes[key])
	}
}

func countMoved(before, after map[string]string) int {
	moved := 0
	for key, owner := range before {
		if after[key] != owner {
			moved++
		}
	}
	return moved
}
