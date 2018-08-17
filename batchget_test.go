package craw

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/wonderivan/logger"
)

func init() {
	logger.SetLogger(`{"Console": {"level": "WARN"}}`)
}

type benchCraw struct {
}

func (this *benchCraw) Init() error {
	return nil
}

func (this *benchCraw) CustomGet(key string) (data interface{}, expired time.Duration, err error) {
	return key, -1, nil
}

func (this *benchCraw) CustomSet(key string, data interface{}) error {
	return nil
}

func (this *benchCraw) Destroy() {

}

func TestGetByOrder(t *testing.T) {
	handler := NewCraw("mytest", new(benchCraw))
	defer handler.Destroy()

	total := 2000000
	keys := make([]string, total)
	fmt.Printf("fill cache use %d test data\n", total)
	for i := 0; i < total; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		handler.SetCraw(keys[i], keys[i], -1)
	}

	start := time.Now()
	for i := 0; i < total; i++ {
		_, err := handler.GetData(keys[i])
		if err != nil {
			log.Fatalf("GetData err:%s", err)
		}
	}
	interval := time.Since(start)

	fmt.Printf("test %d use %v, average:%.2fns/op,%.2fop/s \n",
		total, interval, float64(interval)/float64(total), float64(total)/float64(interval)*1000000000)
}

var outputType int
var outputGraph map[string][]string

func TestStatisticsTable(t *testing.T) {
	outputType = 1
	hitRate := []int{0, 1, 3, 5, 7, 10}
	fmt.Println("测试量|读取命中率|总用时|每条读取耗时|每秒读取条数")
	fmt.Println("|-|-|-|-|-|")
	ExecStatisticsGet(10000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(50000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(100000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(500000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(1000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(2000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	fmt.Println()

	fmt.Println("测试量|写入命中率|总用时|每条写入耗时|每秒写入条数")
	fmt.Println("|-|-|-|-|-|")
	ExecStatisticsSet(10000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(50000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(100000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(500000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(1000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(2000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	fmt.Println()
}

func TestStatisticsGraph(t *testing.T) {
	outputType = 2
	outputGraph = make(map[string][]string)
	hitRate := []int{0, 1, 3, 5, 7, 10}
	ExecStatisticsGet(10000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(50000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(100000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(500000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(1000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsGet(2000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率

	fmt.Println("读取命中率\t1W\t5W\t10W\t50W\t100W\t200W")
	for _, hit := range hitRate {
		k := fmt.Sprintf("%.2f%%", float32(hit)/10*100)
		v := outputGraph[k]
		fmt.Printf("%s", k)
		for _, v1 := range v {
			fmt.Printf("\t%s", v1)
		}
		fmt.Println()
	}
	fmt.Println()

	outputGraph = make(map[string][]string)

	ExecStatisticsSet(10000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(50000, hitRate)   // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(100000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(500000, hitRate)  // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(1000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	ExecStatisticsSet(2000000, hitRate) // 0%,10%,30%,50%,70%,100%命中率
	fmt.Println("写入命中率\t1W\t5W\t10W\t50W\t100W\t200W")
	for _, hit := range hitRate {
		k := fmt.Sprintf("%.2f%%", float32(hit)/10*100)
		v := outputGraph[k]
		fmt.Printf("%s", k)
		for _, v1 := range v {
			fmt.Printf("\t%s", v1)
		}
		fmt.Println()
	}
	fmt.Println()
}

func ExecStatisticsGet(total int, hits []int) {
	for _, hit := range hits {
		ExecStatisticsTestGet(total, hit)
	}
}

func ExecStatisticsSet(total int, hits []int) {
	for _, hit := range hits {
		ExecStatisticsTestSet(total, hit)
	}
}

// hit : 0-10  => 0%-100%
func ExecStatisticsTestGet(total, hit int) {
	handler := NewCraw("mytest1", new(benchCraw))
	defer handler.Destroy()

	keys := make([]string, total)
	for i := 0; i < total; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		if i%10 < hit {
			handler.SetCraw(keys[i], keys[i], -1)
		}
	}

	start := time.Now()
	for i := 0; i < total; i++ {
		_, err := handler.GetData(keys[i])
		if err != nil {
			log.Fatalf("GetData err:%s", err)
		}
	}
	interval := time.Since(start)
	hitRate := float32(hit) / 10 * 100
	if int(hitRate) != int(handler.HitRate()) {
		log.Fatalf("Unexpect hitRate:%v,%v", hitRate, handler.HitRate())
	}
	switch outputType {
	case 1:
		fmt.Printf("|%d|%.2f%%|%v|%.2fns/op|%.2fop|\n",
			total, hitRate,
			interval, float64(interval)/float64(total), float64(total)/float64(interval)*1000000000)
	case 2:
		outputGraph[fmt.Sprintf("%.2f%%", hitRate)] = append(outputGraph[fmt.Sprintf("%.2f%%", hitRate)], fmt.Sprintf("%.2f", float64(total)/float64(interval)*100000))
		fmt.Printf("|%d|%.2f%%|%.2fW|\n",
			total, hitRate, float64(total)/float64(interval)*100000)
	}

}

// hit : 0-10  => 0%-100%
func ExecStatisticsTestSet(total, hit int) {
	handler := NewCraw("mytest1", new(benchCraw))
	defer handler.Destroy()

	keys := make([]string, total)
	for i := 0; i < total; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		if i%10 < hit {
			handler.SetCraw(keys[i], keys[i], -1)
		}
	}

	start := time.Now()
	for i := 0; i < total; i++ {
		err := handler.SetCraw(keys[i], keys[i], -1)
		if err != nil {
			log.Fatalf("SetData err:%s", err)
		}
	}
	interval := time.Since(start)
	hitRate := float32(hit) / 10 * 100
	switch outputType {
	case 1:
		fmt.Printf("|%d|%.2f%%|%v|%.2fns/op|%.2fop|\n",
			total, hitRate,
			interval, float64(interval)/float64(total), float64(total)/float64(interval)*1000000000)
	case 2:
		outputGraph[fmt.Sprintf("%.2f%%", hitRate)] = append(outputGraph[fmt.Sprintf("%.2f%%", hitRate)], fmt.Sprintf("%.2f", float64(total)/float64(interval)*100000))
		fmt.Printf("|%d|%.2f%%|%.2fW|\n",
			total, hitRate, float64(total)/float64(interval)*100000)
	}
}
