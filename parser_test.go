// lexer_test.go
package main

import (
	//"fmt"
	"testing"
)

func compareASTs(generated AST, expected AST) bool {
	return false
}

func testAST(code []byte, expected AST, t *testing.T) {
	tokenChan := make(chan Token, 1)
	go tokenize(code, tokenChan)
	generated := parse(tokenChan)

	if !compareASTs(generated, expected) {
		t.Errorf("Trees don't match.\n")
	}
}

func TestParserExpression1(t *testing.T) {

	var code []byte = []byte(`a = 6 + 7 * variable / -(5 -- (-8 * - 10000.1234))`)

	expected := AST{
		[]Statement{
			Assignment{
				[]Variable{Variable{TYPE_UNKNOWN, "a", ""}},
				[]Expression{
					BinaryOp{
						OP_PLUS, Constant{TYPE_INT, "6"}, BinaryOp{
							OP_MULT, Constant{TYPE_INT, "7"}, BinaryOp{
								OP_DIV, Variable{TYPE_UNKNOWN, "variable", ""}, UnaryOp{
									OP_NEGATIVE, BinaryOp{
										OP_MINUS, Constant{TYPE_INT, "5"}, UnaryOp{
											OP_NEGATIVE, BinaryOp{
												OP_MULT, UnaryOp{OP_NEGATIVE, Constant{TYPE_INT, "8"}}, UnaryOp{
													OP_NEGATIVE, UnaryOp{OP_NEGATIVE, Constant{TYPE_FLOAT, "10000.1234"}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testAST(code, expected, t)

}
