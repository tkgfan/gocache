package main

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

// 测试接口函数是否可以正常使用
func TestGetterFunc_Get(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Error("callback failed")
	}
}

// 测试Get接口
func TestGet(t *testing.T) {
	db := map[string]string{
		"Tom":  "630",
		"Jack": "543",
		"Sam":  "432",
	}
	loadCounts := make(map[string]int, len(db))
	scoresGroup := NewGroup("scoresGroup", 1<<11, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		if view, err := scoresGroup.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := scoresGroup.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("gocache %s miss", k)
		}
	}

	if view, err := scoresGroup.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty,but %s got", view)
	}
}
