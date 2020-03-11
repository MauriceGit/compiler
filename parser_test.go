package main

import (
	"fmt"
	"testing"
)

func compareExpression(e1, e2 Expression) (bool, string) {

	switch v1 := e1.(type) {
	case Constant:
		if v2, ok := e2.(Constant); ok && v1 == v2 {
			return true, ""
		}
		return false, fmt.Sprintf("%v != %v (Constant)", e1, e2)
	case Variable:
		if v2, ok := e2.(Variable); ok && v1 == v2 {
			return true, ""
		}
		return false, fmt.Sprintf("%v != %v (Variable)", e1, e2)
	case BinaryOp:
		if v2, ok := e2.(BinaryOp); ok {
			ok1, err1 := compareExpression(v1.leftExpr, v2.leftExpr)
			ok2, err2 := compareExpression(v1.rightExpr, v2.rightExpr)
			ok3 := v1.operator == v2.operator
			return ok1 && ok2 && ok3, err1 + err2
		}
		return false, fmt.Sprintf("%v != %v (BinaryOp)", e1, e2)
	case UnaryOp:
		if v2, ok := e2.(UnaryOp); ok {
			ok1, err1 := compareExpression(v1.expr, v2.expr)
			return v1.operator == v2.operator && ok1, err1
		}
		return false, fmt.Sprintf("%v != %v (UnaryOp)", e1, e2)
	}
	return false, fmt.Sprintf("%v is not an expression", e1)
}

func compareExpressions(ee1, ee2 []Expression) (bool, string) {
	if len(ee1) != len(ee2) {
		return false, fmt.Sprintf("Different lengths: %v, %v", ee1, ee2)
	}
	for i, v1 := range ee1 {
		if b, e := compareExpression(v1, ee2[i]); !b {
			return false, e
		}
	}
	return true, ""
}

// First time really where I have to say - **** generics (in - not having them!)
func compareVariables(vv1, vv2 []Variable) (bool, string) {
	if len(vv1) != len(vv2) {
		return false, fmt.Sprintf("Different lengths: %v, %v", vv1, vv2)
	}
	for i, v1 := range vv1 {
		if v1 != vv2[i] {
			return false, fmt.Sprintf("Variables are different: %v != %v", v1, vv2[i])
		}
	}
	return true, ""
}

func compareStatement(s1, s2 Statement) (bool, string) {
	switch v1 := s1.(type) {
	case Assignment:
		if v2, ok := s2.(Assignment); ok {
			ok1, err1 := compareVariables(v1.variables, v2.variables)
			ok2, err2 := compareExpressions(v1.expressions, v2.expressions)
			return ok1 && ok2, err1 + err2
		}
		return false, fmt.Sprintf("%v not an Assignment", s2)
	case Condition:
		if v2, ok := s2.(Condition); ok {
			ok1, err1 := compareExpression(v1.expression, v2.expression)
			ok2, err2 := compareBlock(v1.block, v2.block)
			ok3, err3 := compareBlock(v1.elseBlock, v2.elseBlock)
			return ok1 && ok2 && ok3, err1 + err2 + err3
		}
		return false, fmt.Sprintf("%v not a Condition", s2)
	case Loop:
		if v2, ok := s2.(Loop); ok {
			ok1, err1 := compareStatement(v1.assignment, v2.assignment)
			ok2, err2 := compareExpressions(v1.expressions, v2.expressions)
			ok3, err3 := compareStatement(v1.incrAssignment, v2.incrAssignment)
			ok4, err4 := compareBlock(v1.block, v2.block)
			return ok1 && ok2 && ok3 && ok4, err1 + err2 + err3 + err4
		}
	}
	return false, fmt.Sprintf("Expected statement, got: %v", s1)
}

func compareBlock(ss1, ss2 Block) (bool, string) {
	if len(ss1.statements) != len(ss2.statements) {
		return false, fmt.Sprintf("Statement lists of different lengths: %v, %v", ss1, ss2)
	}
	for i, v1 := range ss1.statements {
		if b, e := compareStatement(v1, ss2.statements[i]); !b {
			return false, e
		}
	}
	// TODO: Compare symbol table
	return true, ""
}

func compareASTs(generated AST, expected AST) (bool, string) {
	return compareBlock(generated.block, expected.block)
}

func testAST(code []byte, expected AST, t *testing.T) {
	tokenChan := make(chan Token, 1)
	go tokenize(code, tokenChan)
	generated, err := parse(tokenChan)
	if err != nil {
		t.Errorf("Parsing error: %v", err)
	}

	if b, e := compareASTs(generated, expected); !b {
		t.Errorf("Trees don't match: %v", e)
	}
}

func TestParserExpression1(t *testing.T) {

	var code []byte = []byte(`shadow a = 6 + 7 * variable / -(5 -- (-8 * - 10000.1234))`)

	expected := AST{
		Block{
			[]Statement{
				Assignment{
					[]Variable{Variable{TYPE_UNKNOWN, "a", true}},
					[]Expression{
						BinaryOp{
							OP_PLUS, Constant{TYPE_INT, "6"}, BinaryOp{
								OP_MULT, Constant{TYPE_INT, "7"}, BinaryOp{
									OP_DIV, Variable{TYPE_UNKNOWN, "variable", false}, UnaryOp{
										OP_NEGATIVE, BinaryOp{
											OP_MINUS, Constant{TYPE_INT, "5"}, UnaryOp{
												OP_NEGATIVE, BinaryOp{
													OP_MULT, Constant{TYPE_INT, "-8"}, UnaryOp{
														OP_NEGATIVE, Constant{TYPE_FLOAT, "10000.1234"}, TYPE_UNKNOWN,
													}, TYPE_UNKNOWN, false,
												}, TYPE_UNKNOWN,
											}, TYPE_UNKNOWN, false,
										}, TYPE_UNKNOWN,
									}, TYPE_UNKNOWN, false,
								}, TYPE_UNKNOWN, false,
							}, TYPE_UNKNOWN, false,
						},
					},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserExpression2(t *testing.T) {

	var code []byte = []byte(`a = a && b || (5 < false <= 8 && (false2 > variable >= 5.0) != true)`)

	expected := AST{
		Block{
			[]Statement{
				Assignment{
					[]Variable{Variable{TYPE_UNKNOWN, "a", false}},
					[]Expression{
						BinaryOp{
							OP_AND, Variable{TYPE_UNKNOWN, "a", false}, BinaryOp{
								OP_OR, Variable{TYPE_UNKNOWN, "b", false}, BinaryOp{
									OP_LESS, Constant{TYPE_INT, "5"}, BinaryOp{
										OP_LE, Constant{TYPE_BOOL, "false"}, BinaryOp{
											OP_AND, Constant{TYPE_INT, "8"}, BinaryOp{
												OP_NE, BinaryOp{
													OP_GREATER,
													Variable{TYPE_UNKNOWN, "false2", false},
													BinaryOp{OP_GE, Variable{TYPE_UNKNOWN, "variable", false}, Constant{TYPE_FLOAT, "5.0"}, TYPE_UNKNOWN, false},
													TYPE_UNKNOWN, false,
												},
												Constant{TYPE_BOOL, "true"},
												TYPE_UNKNOWN, false,
											}, TYPE_UNKNOWN, false,
										}, TYPE_UNKNOWN, false,
									}, TYPE_UNKNOWN, false,
								}, TYPE_UNKNOWN, false,
							}, TYPE_UNKNOWN, false,
						},
					},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserIf(t *testing.T) {

	var code []byte = []byte(`
	if a == b {
		a = 6
	}
	a = 1
	`)

	expected := AST{
		Block{
			[]Statement{
				Condition{
					BinaryOp{OP_EQ, Variable{TYPE_UNKNOWN, "a", false}, Variable{TYPE_UNKNOWN, "b", false}, TYPE_UNKNOWN, false},
					Block{[]Statement{Assignment{[]Variable{Variable{TYPE_UNKNOWN, "a", false}}, []Expression{Constant{TYPE_INT, "6"}}}}, SymbolTable{}},
					Block{[]Statement{}, SymbolTable{}},
				},
				Assignment{
					[]Variable{Variable{TYPE_UNKNOWN, "a", false}},
					[]Expression{Constant{TYPE_INT, "1"}},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserIfElse(t *testing.T) {

	var code []byte = []byte(`
	if a == b {
		a = 6
	} else {
		a = 1
	}
	`)

	expected := AST{
		Block{
			[]Statement{
				Condition{
					BinaryOp{OP_EQ, Variable{TYPE_UNKNOWN, "a", false}, Variable{TYPE_UNKNOWN, "b", false}, TYPE_UNKNOWN, false},
					Block{[]Statement{Assignment{[]Variable{Variable{TYPE_UNKNOWN, "a", false}}, []Expression{Constant{TYPE_INT, "6"}}}}, SymbolTable{}},
					Block{[]Statement{Assignment{
						[]Variable{Variable{TYPE_UNKNOWN, "a", false}},
						[]Expression{Constant{TYPE_INT, "1"}},
					}}, SymbolTable{}},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserAssignment(t *testing.T) {

	var code []byte = []byte(`
	a = 1
	a, b = 1, 2
	a, b, c = 1, 2, 3
	`)

	expected := AST{
		Block{
			[]Statement{
				Assignment{
					[]Variable{Variable{TYPE_UNKNOWN, "a", false}},
					[]Expression{Constant{TYPE_INT, "1"}},
				},
				Assignment{
					[]Variable{Variable{TYPE_UNKNOWN, "a", false}, Variable{TYPE_UNKNOWN, "b", false}},
					[]Expression{Constant{TYPE_INT, "1"}, Constant{TYPE_INT, "2"}},
				},
				Assignment{
					[]Variable{Variable{TYPE_UNKNOWN, "a", false}, Variable{TYPE_UNKNOWN, "b", false}, Variable{TYPE_UNKNOWN, "c", false}},
					[]Expression{Constant{TYPE_INT, "1"}, Constant{TYPE_INT, "2"}, Constant{TYPE_INT, "3"}},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserFor1(t *testing.T) {

	var code []byte = []byte(`
	for ;; {
		a = a+1
	}
	`)

	expected := AST{
		Block{
			[]Statement{
				Loop{
					Assignment{[]Variable{}, []Expression{}},
					[]Expression{},
					Assignment{[]Variable{}, []Expression{}},
					Block{[]Statement{
						Assignment{
							[]Variable{Variable{TYPE_UNKNOWN, "a", false}},
							[]Expression{BinaryOp{OP_PLUS, Variable{TYPE_UNKNOWN, "a", false}, Constant{TYPE_INT, "1"}, TYPE_UNKNOWN, false}},
						},
					}, SymbolTable{}},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserFor2(t *testing.T) {

	var code []byte = []byte(`
	for i = 5;; {
		a = 0
	}
	`)

	expected := AST{
		Block{
			[]Statement{
				Loop{
					Assignment{[]Variable{Variable{TYPE_UNKNOWN, "i", false}}, []Expression{Constant{TYPE_INT, "5"}}},
					[]Expression{},
					Assignment{[]Variable{}, []Expression{}},
					Block{[]Statement{
						Assignment{
							[]Variable{Variable{TYPE_UNKNOWN, "a", false}},
							[]Expression{Constant{TYPE_INT, "0"}},
						},
					}, SymbolTable{}},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}

func TestParserFor3(t *testing.T) {

	var code []byte = []byte(`
	for i, j = 0, 1; i < 10; i = i+1 {
		if b == a {
			for ;; {
				c = 6
			}
		}
	}
	`)

	expected := AST{
		Block{
			[]Statement{
				Loop{
					Assignment{
						[]Variable{Variable{TYPE_UNKNOWN, "i", false}, Variable{TYPE_UNKNOWN, "j", false}},
						[]Expression{Constant{TYPE_INT, "0"}, Constant{TYPE_INT, "1"}},
					},
					[]Expression{BinaryOp{OP_LESS, Variable{TYPE_UNKNOWN, "i", false}, Constant{TYPE_INT, "10"}, TYPE_UNKNOWN, false}},
					Assignment{
						[]Variable{Variable{TYPE_UNKNOWN, "i", false}},
						[]Expression{BinaryOp{OP_PLUS, Variable{TYPE_UNKNOWN, "i", false}, Constant{TYPE_INT, "1"}, TYPE_UNKNOWN, false}},
					},
					Block{[]Statement{
						Condition{
							BinaryOp{OP_EQ, Variable{TYPE_UNKNOWN, "b", false}, Variable{TYPE_UNKNOWN, "a", false}, TYPE_UNKNOWN, false},
							Block{[]Statement{
								Loop{
									Assignment{[]Variable{}, []Expression{}},
									[]Expression{},
									Assignment{[]Variable{}, []Expression{}},
									Block{[]Statement{
										Assignment{
											[]Variable{Variable{TYPE_UNKNOWN, "c", false}},
											[]Expression{Constant{TYPE_INT, "6"}},
										},
									}, SymbolTable{}},
								},
							}, SymbolTable{}},
							Block{[]Statement{}, SymbolTable{}},
						},
					}, SymbolTable{}},
				},
			},
			SymbolTable{},
		},
		SymbolTable{},
	}

	testAST(code, expected, t)
}
