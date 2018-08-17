package craw

import (
	"fmt"
	"log"
	"testing"
	"time"
)

type cacheGet struct {
	testV int
}

func (this *cacheGet) Init() error {
	this.testV = 0
	return nil
}

func (this *cacheGet) CustomGet(key string) (data interface{}, expired time.Duration, err error) {
	this.testV++
	return this.testV, 0, nil
}

func (this *cacheGet) CustomSet(key string, data interface{}) error {
	return nil
}

func (this *cacheGet) Destroy() {

}

func TestCrawGetPut(t *testing.T) {
	handler := NewCraw("mytest", new(cacheGet))
	defer handler.Destroy()
	count := 0
	for count < 10 {
		v, err := handler.GetData("k1")
		if err != nil {
			log.Fatalf("GetData err:%s", err)
		}
		log.Printf("GetData type(%T), value:%v", v, v)

		count = v.(int)

		err = handler.SetCraw("k1", count+1, 1)
		if err != nil {
			log.Fatalf("GetData err:%s", err)
		}
	}
}

func TestCrawHitRate(t *testing.T) {
	handler := NewCraw("mytest", new(cacheGet))
	defer handler.Destroy()
	count := 0
	for count < 100000 {
		key := fmt.Sprintf("key%d", count)
		_, err := handler.GetData(key)
		if err != nil {
			log.Fatalf("GetData err:%s", err)
		}
		//log.Printf("GetData %s type(%T), value:%v", key, v, v)

		if count%3 == 0 {
			key = fmt.Sprintf("key%d", count+1)
			err = handler.SetCraw(key, count+1, 10)
			if err != nil {
				log.Fatalf("GetData err:%s", err)
			}
			//log.Printf("SetData %s value:%v", key, v)
		}
		count++
	}

	log.Printf("hit rate:%v", handler.HitRate()) // output: 1/3 = 33.3333%
}

// go test -benchmem -run=^$ -bench ^BenchmarkPut$
func BenchmarkPut(b *testing.B) {
	handler := NewCraw("mytest2", new(cacheGet))
	defer handler.Destroy()

	b.ResetTimer()
	b.N = 10000000
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	for i := 0; i < b.N; i++ {
		handler.SetCraw(keys[i], keys[i], -1)
	}
}

// go test -benchmem -run=^$ -bench ^BenchmarkGet$
func BenchmarkGet(b *testing.B) {
	handler := NewCraw("mytest", new(cacheGet))
	defer handler.Destroy()

	b.N = 10000000
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		handler.SetCraw(keys[i], keys[i], -1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.GetData(keys[i])
		if err != nil {
			log.Fatalf("GetData err:%s", err)
		}
	}
}
