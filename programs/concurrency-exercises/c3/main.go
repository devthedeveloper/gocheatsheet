package main
import "fmt"
func main(){
	ch := make(chan int, 1)
	ch <- 42
	close(ch)
	v, ok := <-ch; fmt.Println(v, ok)
	v, ok = <-ch; fmt.Println(v, ok)
}
