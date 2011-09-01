package db

import (
    "testing"
    "fmt"
    "rand"
)

func TestLRUCache(t *testing.T) {
    strategy := NewLRUPeriodFlushStrategy(3, 100, 0)
    storage := NewInMemoryKeyValueStorage(10)
    flusher := &FakeFlusher{}

    cache := NewKeyValueCache(storage, strategy, flusher)
    fmt.Print("Start LRU cache test ...\t")

    for i := 0; i < 10; i++ {
        str := fmt.Sprint(i)
        cache.Show(str, str)
    }

    keys, _ := cache.Keys()
    if !same(convert2string([]int{7,9,8}), keys) {
        t.Errorf("should be [7 8 9], but %v", keys)
    }

    if v, _ := cache.Get("1"); v != nil {
        t.Errorf("%v should not be in cache", v)
    }

    cache.Get("7")
    cache.Show("1", "1")

    if v, _ := cache.Get("8"); v != nil {
        keys, _ := cache.Keys()
        t.Errorf("%v should not be in cache; cache content: %v", v, keys)
    }
    fmt.Print("OK\n")
}

func BenchmarkKeyValueCache(b *testing.B) {
    strategy := NewLRUPeriodFlushStrategy(200, 100, 0)
    storage := NewInMemoryKeyValueStorage(200)
    flusher := &FakeFlusher{}

    cache := NewKeyValueCache(storage, strategy, flusher)
    for i:= 0; i < 100000; i++ {
        key_int := rand.Int() % 1000
        key := fmt.Sprintf("%d", key_int)
        v, _ := cache.Get(key)
        if v == nil {
            cache.Show(key, key)
        }
    }
}
