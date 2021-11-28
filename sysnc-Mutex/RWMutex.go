/*
读写锁
Rlock 不阻塞
lock 阻塞
*/
package main

import (
	"fmt"
	"sync"
	"time"
)

var m *sync.RWMutex

func main() {
	wg := sync.WaitGroup{}
	wg.Add(20)
	var rwMutex sync.RWMutex
	Data := 0
	for i := 0; i < 10; i++ {
		go func(t int) {
			rwMutex.Lock()
			defer rwMutex.Unlock()
			defer fmt.Println("rwMutex Unlocked")
			Data += t
			fmt.Printf("Write Data: %v %d \n", Data, t)
			wg.Done()
			// 这句代码让写锁的效果显示出来，写锁定下是需要解锁后才能写的。
			time.Sleep(3 * time.Second)
		}(i)
		go func(t int) {
			rwMutex.RLock()
			defer rwMutex.RUnlock()
			defer fmt.Println("rwMutex RUnlocked")
			fmt.Printf("Read data: %v\n", Data)
			wg.Done()
			time.Sleep(1 * time.Second)
			// 这句代码第一次运行后，读解锁。
			// 循环到第二个时，读锁定后，这个goroutine就没有阻塞，同时读成功。
		}(i)
	}
	time.Sleep(5 * time.Second)
	wg.Wait()
}
