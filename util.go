package tinypubsub

import (
	"fmt"
	"sync"
)

var Debug = false

func GoWithWaitGroup(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		f()
		wg.Done()
	}()
}

func Debugf(format string, a ...any) (n int, err error) {
	if Debug {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}

func Debugln(a ...any) (n int, err error) {
	if Debug {
		return fmt.Println(a...)
	}

	return 0, nil
}

func Infof(format string, a ...any) (n int, err error) {
	return fmt.Printf(format, a...)
}

func Infoln(a ...any) (n int, err error) {
	return fmt.Println(a...)
}
