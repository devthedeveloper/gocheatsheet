package main
import ("fmt";"sort";"sync")
func process(n int) int { return n * n }
func workerPool(jobs []int, workers int) []int {
	in := make(chan int)
	out := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(){ defer wg.Done(); for j := range in { out <- process(j) } }()
	}
	go func(){ for _, j := range jobs { in <- j }; close(in) }()
	go func(){ wg.Wait(); close(out) }()
	var res []int
	for r := range out { res = append(res, r) }
	sort.Ints(res)
	return res
}
func main(){ fmt.Println(workerPool([]int{1,2,3,4,5}, 3)) }
