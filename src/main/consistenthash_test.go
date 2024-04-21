package main

import (
	"MisakaCache/src/misakacache/consistenthash"
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	test_map := consistenthash.NewMap(func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	}, 3)

	test_map.AddRealNode("2", "4", "6")

	test_case := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range test_case {
		if r := test_map.GetRealNodeByKey(k); r != v {
			t.Errorf("Asking for %s, should have yielded %s, actually is %s", k, v, r)
		}
	}

	test_map.AddRealNode("8")
	test_case["27"] = "8"
	for k, v := range test_case {
		if r := test_map.GetRealNodeByKey(k); r != v {
			t.Errorf("Asking for %s, should have yielded %s, actually is %s", k, v, r)
		}
	}
}
