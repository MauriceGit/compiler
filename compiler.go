// lexer.go
package main

import (
	"fmt"
)

func main() {
	var program []byte = []byte(`
		value = 100 == 7
		x, test = 50.6, "blubb"
		if 5 == 6 {
			x2 = 7
		} else {
			a = b
		}
		for ;; {
			a = 0
		}

		for i = 5;; {
			a = 0
		}

		for i = 5; i < 10; i = i+1 {
			if b == a {
				for ;; {
					c = 6
				}
			}
		}

		test2 = "..."
`)

	fmt.Println(string(program))

	tokenChan := make(chan Token, 1)
	go tokenize(program, tokenChan)

	ast := parse(tokenChan)

	fmt.Printf("\n%v\n", ast)

}
