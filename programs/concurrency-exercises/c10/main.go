package main
import ("context";"errors";"fmt";"time")
func runWithTimeout(ctx context.Context, fn func() int) (int, error) {
	done := make(chan int, 1) // buffered → goroutine never leaks
	go func(){ done <- fn() }()
	select {
	case v := <-done: return v, nil
	case <-ctx.Done(): return 0, ctx.Err()
	}
}
func main(){
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	v, err := runWithTimeout(ctx, func() int { time.Sleep(200*time.Millisecond); return 99 })
	fmt.Println(v, errors.Is(err, context.DeadlineExceeded))
	v, err = runWithTimeout(context.Background(), func() int { return 7 })
	fmt.Println(v, err)
}
