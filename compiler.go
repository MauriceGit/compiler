// lexer.go
package main

import (
	"fmt"
)

func main() {
	var program []byte = []byte(`

value = 100 == 7
x, test = 50.6, "blubb"
for i = 0; i < 10; i = i+1 {
	blubb
}
x = 6.8
i = 6.8
b = !true

`)

	fmt.Println(string(program))

	tokenChan := make(chan Token, 1)
	go tokenize(program, tokenChan)

	ast := parse(tokenChan)

	ast, err := annotateTypes(ast)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n%v\n", ast)

}
