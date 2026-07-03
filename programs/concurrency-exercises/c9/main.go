package main
import ("fmt";"sort";"sync")
func merge(chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range chans {
		wg.Add(1)
		go func(c <-chan int){ defer wg.Done(); for v := range c { out <- v } }(c)
	}
	go func(){ wg.Wait(); close(out) }()
	return out
}
func gen(nums ...int) <-chan int {
	c := make(chan int)
	go func(){ for _, n := range nums { c <- n }; close(c) }()
	return c
}
func main(){
	var res []int
	for v := range merge(gen(1,2,3), gen(4,5,6)) { res = append(res, v) }
	sort.Ints(res)
	fmt.Println(res)
}
