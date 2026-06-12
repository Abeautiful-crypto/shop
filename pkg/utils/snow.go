package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/BwCloudWeGo/bw-cli/pkg/config"
	"github.com/BwCloudWeGo/bw-cli/pkg/redisx"
	"sync"
	"time"
)

const (
	workerBits  uint8 = 10
	numberBits  uint8 = 12
	workerMax   int64 = -1 ^ (-1 << workerBits)
	numberMax   int64 = -1 ^ (-1 << numberBits)
	timeShift   uint8 = workerBits + numberBits
	workerShift uint8 = numberBits
	startTime   int64 = 1525705533000 // 如果在程序跑了一段时间修改了epoch这个值 可能会导致生成相同的ID
)

type Worker struct {
	mu        sync.Mutex
	timestamp int64
	workerId  int64
	number    int64
}

func NewWorker(workerId int64) (*Worker, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("Worker ID excess of quantity")
	}
	// 生成一个新节点
	return &Worker{
		timestamp: 0,
		workerId:  workerId,
		number:    0,
	}, nil
}

func (w *Worker) GetId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now().UnixNano() / 1e6
	if w.timestamp == now {
		w.number++
		if w.number > numberMax {
			for now <= w.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		w.number = 0
		w.timestamp = now
	}
	ID := int64((now-startTime)<<timeShift | (w.workerId << workerShift) | (w.number))
	return ID
}
func SnowId() int64 {
	// 生成节点实例
	node, err := NewWorker(1)
	if err != nil {
		panic(err)
	}

	return node.GetId()
}

// RedisLock 获取 Redis 锁
func RedisLock() bool {
	rdb := redisx.NewClient(config.GlobalConfig.Redis)
	ok, err := rdb.SetNX(context.Background(), "my_lock", "1", 10*time.Second).Result()
	if err != nil {
		return false
	}
	return ok //返回的bool-true
}

// Unlock 释放锁
func Unlock() {
	rdb := redisx.NewClient(config.GlobalConfig.Redis)
	rdb.Del(context.Background(), "my_lock")
}

func GetSnowId() int64 {
	var id int64
	// 抢到锁
	if RedisLock() {
		id = SnowId()
	} else {
		id = 0
	}
	defer Unlock() // 函数结束释放锁

	if id == 0 {
		fmt.Println("获取雪花算法id失败")
	}
	return id
}
