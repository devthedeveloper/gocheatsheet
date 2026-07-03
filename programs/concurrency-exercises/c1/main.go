package main
import ("fmt";"sync")
func main(){
	var wg sync.WaitGroup
	got := make([]int, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(){ defer wg.Done(); got[i] = i }()
	}
	wg.Wait()
	fmt.Println(got)
}
