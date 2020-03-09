// lexer.go
package main

import (
	"fmt"
)

func main() {
	var program []byte = []byte(`
	for ;; {
		a = a+1
	}

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
