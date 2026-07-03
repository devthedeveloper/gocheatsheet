package main
import ("fmt";"runtime";"time")
// BUG: unbuffered channel + a slow sender. When the timeout fires first,
// leaky() returns but the goroutine is stuck on `ch <- 42` forever.
func leaky() int {
	ch := make(chan int)
	go func(){ time.Sleep(100*time.Millisecond); ch <- 42 }()
	select {
	case v := <-ch: return v
	case <-time.After(10*time.Millisecond): return -1 // fires first
	}
}
func main(){
	before := runtime.NumGoroutine()
	for i:=0;i<100;i++{ leaky() }
	time.Sleep(300*time.Millisecond) // let senders wake and block on send
	fmt.Println("leaked goroutines still alive:", runtime.NumGoroutine()-before)
}
