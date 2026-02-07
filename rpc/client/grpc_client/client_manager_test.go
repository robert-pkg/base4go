package grpc_client

import (
	"strconv"
	"sync"
	"testing"
)

// 基准测试：LoadOrStore
// go test -bench=BenchmarkLoadOrStore -benchtime=5s -v --count=1 -benchmem -timeout 10m
func BenchmarkLoadOrStore(b *testing.B) {
	var (
		syncMap sync.Map
	)

	// 1. 基准测试函数命名必须以 Benchmark 开头
	// 2. 参数必须是 *testing.B 类型

	b.RunParallel(func(pb *testing.PB) {
		// 3. 并发执行测试
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			syncMap.LoadOrStore(key, i)
			i++
		}
	})
}

// 基准测试：互斥锁 + map
func BenchmarkMutexMap(b *testing.B) {
	var (
		globalMutex sync.Mutex
		simpleMap   = make(map[string]interface{})
	)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			globalMutex.Lock()
			simpleMap[key] = i
			globalMutex.Unlock()
			i++
		}
	})
}

// 基准测试：双重检查锁
func BenchmarkDoubleCheck(b *testing.B) {
	var (
		globalMutex sync.Mutex
		simpleMap   sync.Map
	)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			// Load before attempting to store
			if _, ok := simpleMap.Load(key); !ok {

				globalMutex.Lock()
				// Use LoadOrStore to atomically check and store
				if _, ok2 := simpleMap.Load(key); ok2 {
					// ok
				} else {
					simpleMap.Store(key, i)
				}

				globalMutex.Unlock()
			}
			i++
		}
	})
}

func BenchmarkRWLock(b *testing.B) {

	cm := GetClientMgr()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			cm.GetClient(key)
		}
	})
}
