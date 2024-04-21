package main

import (
	lru2 "MisakaCache/src/misakacache/lru"
	"reflect"
	"testing"
)

type String string

func (string String) GetMemoryUsed() int { // 要实现Value接口的那个方法 下面才能直接添加什么的
	return len(string)
}

func TestCache_GetValue(t *testing.T) {
	lru := lru2.NewLRU(int64(100), nil)
	lru.SetValue("key1", String("1234"))
	if value, isOK := lru.GetValue("key1"); !isOK || value != String("1234") {
		t.Fatalf("cache hit key1=1234 failed, value actually is %s", value)
	}
	if _, isOK := lru.GetValue("key2"); isOK {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestCache_RemoveOldestCache(t *testing.T) {
	key1, key2, key3 := "key1", "key2", "key3"
	value1, value2, value3 := "value1", "value2", "value3"

	capacity := len(key1 + key2 + value1 + value2)
	lru := lru2.NewLRU(int64(capacity), nil)
	lru.SetValue(key1, String(value1))
	lru.SetValue(key2, String(value2))
	lru.SetValue(key3, String(value3))

	if _, isOk := lru.GetValue(key1); isOk || lru.GetLRUEntryNumber() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestCache_OnEntryDeleted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value lru2.Value) {
		keys = append(keys, key)
	}
	lru := lru2.NewLRU(int64(20), callback)
	lru.SetValue("key1", String("value1"))
	lru.SetValue("key2", String("value2"))
	lru.SetValue("key3", String("value3"))
	lru.SetValue("key4", String("value3"))
	expected := []string{"key1", "key2"}

	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s, now keys are %s", expected, keys)
	}
}
