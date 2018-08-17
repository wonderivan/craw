package craw

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/wonderivan/logger"
	lrucache "github.com/wonderivan/lrucache"
)

// 缓存数据操作接口
type CrawInterface interface {
	// 初始化
	Init() error
	// 获取远端数据的方式
	CustomGet(key string) (data interface{}, expired time.Duration, err error)
	// 设置远端数据的方式，设置后会清除对应的缓存
	CustomSet(key string, data interface{}) (err error)
	// 销毁
	Destroy()
}

type customValue struct {
	data interface{}
	err  error
	wait chan bool
}

// 获取远端数据模块
type Craw struct {
	dispose CrawInterface
	keyMap  map[string]*customValue
	craw    *lrucache.LruCache
	enable  bool
	sync.RWMutex

	// 统计命中率
	totalAccess   *uint64 // 总访问量
	staggerAccess *uint64 // 未命中的访问数
}

// 创建一个craw
//
// 参数 crawName:缓存名称，config缓存配置， dispose缓存数据处理方法
func NewCraw(crawName string, dispose CrawInterface, config ...string) *Craw {
	logger.Info("craw(%s) use setting:%s", crawName, config)
	c := new(Craw)
	c.dispose = dispose
	c.keyMap = make(map[string]*customValue)
	c.craw = lrucache.NewLruCache(crawName, config...)
	c.totalAccess = new(uint64)
	c.staggerAccess = new(uint64)
	err := c.dispose.Init()
	if err == nil {
		c.enable = true
	}

	return c
}

// 销毁Craw
func (c *Craw) Destroy() {
	c.enable = false
	c.craw.Destroy()
	c.dispose.Destroy()
}

// 获取远端数据更新到缓存并返回
//
// 参数 key:要查找的远端数据的key
// 返回值 interface{}:找到的数据  error:成功为nil
func (c *Craw) GetData(key string) (interface{}, error) {
	if !c.enable {
		return nil, errors.New("this craw has been destroyed.")
	}

	atomic.AddUint64(c.totalAccess, 1)

	// 先查找craw
	data, exist := c.craw.GetEx(key)
	if exist {
		return data, nil
	}

	c.Lock()
	value, ok := c.keyMap[key]

	if !ok {
		atomic.AddUint64(c.staggerAccess, 1)

		valueOpt := &customValue{nil, nil, make(chan bool)}
		c.keyMap[key] = valueOpt
		c.Unlock()
		// 此key没找过
		// 开始找
		var expired time.Duration
		valueOpt.data, expired, valueOpt.err = c.dispose.CustomGet(key)
		if valueOpt.err == nil {
			if valueOpt.err == nil {
				c.craw.Put(key, valueOpt.data, expired)
			}
		}
		c.Lock()
		delete(c.keyMap, key)
		c.Unlock()
		close(valueOpt.wait)
		return valueOpt.data, valueOpt.err
	} else {
		c.Unlock()
		// 其他线程在找同个key，阻塞
		<-value.wait
		// 其他线程执行完，从map的data取
		return value.data, value.err
	}
}

// 强制从远端获取数据，并更新craw
//
// 参数 key:要查找的远端数据的key
// 返回值 interface 更新到cache的数据
// 返回值 成功为nil
func (c *Craw) UpdateData(key string) (data interface{}, err error) {
	if !c.enable {
		return nil, errors.New("this craw has been destroyed.")
	}
	data, expired, err := c.dispose.CustomGet(key)
	if err == nil && data != nil {
		c.craw.Put(key, data, expired)
	}
	return data, err
}

//删除craw指定数据
//
// 参数 key:要查找的远端数据的key
// 参数可选 delay:延迟删除数据的时间 单位(s)
// 返回值 成功为nil
func (c *Craw) DeleteData(key string, delay ...time.Duration) (err error) {
	if !c.enable {
		return errors.New("this craw has been destroyed.")
	}
	if len(delay) > 0 {
		return c.craw.DelayDelete(key, delay[0])
	}
	return c.craw.Delete(key)

}

// 保存数据到远端，并删除Craw中已有的缓存值
//
// 参数 key:要保存到远端的数据的key  data:要保存到远端的数据
// 返回值 error:成功为nil
func (c *Craw) SetData(key string, data interface{}) (err error) {
	if !c.enable {
		return errors.New("this craw has been destroyed.")
	}

	err = c.dispose.CustomSet(key, data)
	if err != nil {
		c.craw.Delete(key)
	}
	return
}

//清除craw的所有数据
func (c *Craw) ClearAll() error {
	if !c.enable {
		return errors.New("this craw has been destroyed.")
	}
	c.craw.ClearAll()
	return nil
}

// 清除craw中指定的包含前缀prefix的key的数据
func (c *Craw) ClearPrefixKeys(Prefix string) error {
	if !c.enable {
		return errors.New("this craw has been destroyed.")
	}
	c.craw.ClearPrefixKeys(Prefix)
	return nil
}

// 获取craw的缓存命中率
//
// 返回值 float64:计算的结果，XX.XXXXX%
func (c *Craw) HitRate() float64 {
	custom := atomic.LoadUint64(c.staggerAccess)
	total := atomic.LoadUint64(c.totalAccess)
	if total == 0 {
		return 0
	}
	return float64(total-custom) / float64(total) * 100
}

// 重置craw命中率并返回重置之前命中率
//
// 返回值 float64:计算的结果，XX.XXXXX%
func (c *Craw) ResetHitRate() float64 {
	custom := atomic.SwapUint64(c.staggerAccess, 0)
	total := atomic.SwapUint64(c.totalAccess, 0)
	if total == 0 {
		return 0
	}
	return float64(total-custom) / float64(total) * 100
}

// 设置craw数据，不更新远端
//
// 参数 key:要保存的数据的key  data:要保存的数据,expired要保存的数据的过期时间，<0不过期
// 返回值 error:成功为nil
func (c *Craw) SetCraw(key string, data interface{}, expired time.Duration) error {
	if !c.enable {
		return errors.New("this craw has been destroyed.")
	}
	c.craw.Put(key, data, expired)
	return nil
}

// 查询craw中指定的key是否存在
func (c *Craw) IsExist(key string) (bool, error) {
	if !c.enable {
		return false, errors.New("this craw has been destroyed.")
	}
	return c.craw.IsExist(key), nil
}
