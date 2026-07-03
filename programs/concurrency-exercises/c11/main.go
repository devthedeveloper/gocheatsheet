package main
import ("fmt";"sync")
// bounded parallelism, results in INPUT order.
func parallelMap(in []int, limit int, f func(int) int) []int {
	out := make([]int, len(in))
	sem := make(chan struct{}, limit) // token bucket = concurrency cap
	var wg sync.WaitGroup
	for i, v := range in {
		wg.Add(1)
		sem <- struct{}{} // acquire (blocks if `limit` already running)
		go func(i, v int){
			defer wg.Done()
			defer func(){ <-sem }() // release
			out[i] = f(v) // distinct indices → no race, order preserved
		}(i, v)
	}
	wg.Wait()
	return out
}
func main(){
	fmt.Println(parallelMap([]int{1,2,3,4,5,6}, 2, func(n int) int { return n*10 }))
}
