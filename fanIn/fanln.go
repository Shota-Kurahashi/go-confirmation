package main

import (
	"fmt"
	"sync"
)

//* 複数個あるチャネルから受信した値を、1つの受信用チャネルの中にまとめる方法をFanInといいます。

//? まとめたいチャネルの数が固定の場合は、select文を使って簡単に実装できます。

func fanIn1(done chan struct{}, c1, c2 <-chan int) <-chan int {
	result := make(chan int)

	go func() {
		defer fmt.Println("closed fanin")
		defer close(result)
		for {
			// caseはfor文で回せないので(=可変長は無理)
			// 統合元のチャネルがスライスでくるとかだとこれはできない
			// →応用編に続く

			// done（channelが閉じてる）場合は終了
			// そうでない場合は、c1, c2から値を受信してresultに送信
			// どちらも受信できない場合は、defaultに入ってループを継続
			select {
			case <-done:
				fmt.Println("done")
				return
			case num := <-c1:
				fmt.Println("send 1")
				result <- num
			case num := <-c2:
				fmt.Println("send 2")
				result <- num
			default:
				fmt.Println("continue")
				continue
			}
		}
	}()

	return result
}

func generate(done <-chan struct{}, i int) <-chan int {
	result := make(chan int)

	go func() {
		defer close(result)
	loop:
		for {
			select {
			case <-done:
				fmt.Println("done")
				break loop
			case result <- i:

			}
		}

	}()

	return result
}

func fanIn2(done chan struct{}, cs ...<-chan int) <-chan int {
	result := make(chan int)

	var wg sync.WaitGroup
	fmt.Println("wg.Add", len(cs), "times")
	wg.Add(len(cs))

	for i, c := range cs {
		// FanInの対象になるチャネルごとに
		// 個別にゴールーチンを立てちゃう
		go func(c <-chan int, i int) {
			defer wg.Done()

			for num := range c {
				select {
				case <-done:
					fmt.Println("wg.Done", i)
					return
				case result <- num:
					fmt.Println("send", i)
				}
			}
		}(c, i)
	}

	go func() {
		// selectでdoneが閉じられるのを待つと、
		// 個別に立てた全てのゴールーチンを終了できる保証がない
		wg.Wait()
		fmt.Println("closing fanin")
		close(result)
	}()

	return result
}

func main() {
	done := make(chan struct{})

	// gen1 := generate(done, 1)
	// gen2 := generate(done, 2)

	// fanin := fanIn1(done, gen1, gen2)

	// for i := 0; i < 10; i++ {
	// 	<-fanin
	// }

	// fmt.Println("closed done")

	// // これを使って、main関数でcloseしている間に送信された値を受信しないと
	// // チャネルがブロックされてしまってゴールーチンリークになってしまう恐れがある
	// for {
	// 	if _, ok := <-fanin; !ok {
	// 		break
	// 	}
	// }

	gen1 := generate(done, 1)

	gen2 := generate(done, 2)
	fanin := fanIn2(done, gen1, gen2)

	for i := 0; i < 10; i++ {
		fmt.Println(<-fanin)
	}
	close(done)

	// t := time.NewTimer(1 * time.Second)

	// t.Cはチャネルなので、ここで受信待ちになる

}

//* NewTimerは以下のようになっている。

// type Timer struct {
// 	C <-chan Time

// }

// func NewTimer(d Duration) *Timer

//* Afterも同じようになっている。
// func After(d Duration) <-chan Time

// これを用いるとx秒間待つという処理を以下のように書ける。

// func main() {
// 	t := time.NewTimer(1 * time.Second)

// 	// t.Cはチャネルなので、ここで受信待ちになる
// 	<-t.C
// 	fmt.Println("1秒経過")
// }

// このように、チャネルを使うことで、指定した時間が経過したことを検知できます。
