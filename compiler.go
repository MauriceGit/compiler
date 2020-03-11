// lexer.go
package main

import (
	"fmt"
)

func main() {
	var program []byte = []byte(`

value = 100 * (7 + 8)

`)

	fmt.Println(string(program))

	tokenChan := make(chan Token, 1)
	go tokenize(program, tokenChan)

	ast, err := parse(tokenChan)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast, err = analyzeTypes(ast)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n%v\n", ast)

}
