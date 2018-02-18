package main

import "fmt"

func main() {
	func() {
		fmt.Println("Hello World")
	}()
}
