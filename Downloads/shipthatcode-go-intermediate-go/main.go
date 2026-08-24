package main
import ("bufio"; "fmt"; "os"; "strconv")

// type Shape interface { ... }
// type Circle struct { ... }
// func (c Circle) Area() float64 { ... }

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); kind := sc.Text()
    sc.Scan(); dim, _ := strconv.ParseFloat(sc.Text(), 64)
    var s interface{ Area() float64 }
    _ = kind; _ = dim
    // s = ... based on kind
    if s != nil { fmt.Printf("%.2f\n", s.Area()) }
}
