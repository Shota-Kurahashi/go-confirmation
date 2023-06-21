package goroutine

import (
	"fmt"
)

func infiniteFunc(done <-chan interface{}) <-chan int {

	result := make(chan int)

	go func() {
	Loop:
		for {
			select {
			case <-done:
				break Loop

			case result <- 1:

			}
		}
	}()

	return result
}

func restFunc() <-chan int {
	result := make(chan int)

	go func() {
		defer close(result)
		defer fmt.Println("restFunc() is closed.")
		for i := 0; i < 5; i++ {
			result <- i
		}

	}()

	return result
}

func interruptFunc() {
	// 終了したいタイミングでチャネルを閉じる
	done := make(chan interface{})
	defer close(done)

	for i := 0; i < 10; i++ {
		fmt.Println("infiniteFunc():", <-infiniteFunc(done))
	}

}

func Main() {
	gen1, gen2 := restFunc(), restFunc()

	for i := 0; i < 5; i++ {
		fmt.Println("gen1:", <-gen1)
		fmt.Println("gen2:", <-gen2)
	}

	interruptFunc()

}
