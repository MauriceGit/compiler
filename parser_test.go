package main

import (
	"fmt"
	"testing"
)

func (e Constant) eq(e2 Constant) bool {
	return e.cType == e2.cType && e.cValue == e2.cValue
}
func (e Variable) eq(e2 Variable) bool {
	return e.vName == e2.vName && e.vShadow == e2.vShadow && e.vType == e2.vType
}

func compareFunCall(v1, v2 FunCall) (bool, string) {

	if v1.funName != v2.funName {
		return false, fmt.Sprintf("Function names are different: %v != %v", v1.funName, v2.funName)
	}
	ok1, err1 := compareExpressions(v1.args, v2.args)
	ok2, err2 := compareTypes(v1.retTypes, v2.retTypes)
	return ok1 && ok2, err1 + err2
}

func compareExpression(e1, e2 Expression) (bool, string) {

	switch v1 := e1.(type) {
	case Constant:
		if v2, ok := e2.(Constant); ok && v1.eq(v2) {
			return true, ""
		}
		return false, fmt.Sprintf("%v != %v (Constant)", e1, e2)
	case Variable:
		if v2, ok := e2.(Variable); ok && v1.eq(v2) {
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
	case FunCall:
		if v2, ok := e2.(FunCall); ok {
			return compareFunCall(v1, v2)
		}
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
		if !v1.eq(vv2[i]) {
			return false, fmt.Sprintf("Variables are different: %v != %v", v1, vv2[i])
		}
	}
	return true, ""
}

func compareTypes(tt1, tt2 []ComplexType) (bool, string) {
	if len(tt1) != len(tt2) {
		return false, fmt.Sprintf("Different lengths: %v, %v", tt1, tt2)
	}
	for i, t := range tt1 {
		if !equalType(t, tt2[i]) {
			return false, fmt.Sprintf("Type comparison: %v != %v", t, tt2[i])
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
	case Function:
		if v2, ok := s2.(Function); ok {
			if v1.fName != v2.fName {
				return false, fmt.Sprintf("Function names are different: %v != %v", v1.fName, v2.fName)
			}
			ok1, err1 := compareVariables(v1.parameters, v2.parameters)
			ok2, err2 := compareTypes(v1.returnTypes, v2.returnTypes)
			ok3, err3 := compareBlock(v1.block, v2.block)
			return ok1 && ok2 && ok3, err1 + err2 + err3
		}
	case Return:
		if v2, ok := s2.(Return); ok {
			return compareExpressions(v1.expressions, v2.expressions)
		}
	case FunCall:
		if v2, ok := s2.(FunCall); ok {
			return compareFunCall(v1, v2)
		}
	default:
		return false, fmt.Sprintf("Unknown statement: %v", s1)
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
	lexerErr := make(chan error, 1)
	go tokenize(code, tokenChan, lexerErr)

	generated, err := parse(tokenChan)

	select {
	case e := <-lexerErr:
		t.Errorf("%v", e.Error())
		return
	default:
	}
	if err != nil {
		t.Errorf("Parsing error: %v", err)
	}

	if b, e := compareASTs(generated, expected); !b {
		t.Errorf("Trees don't match: %v", e)
	}
}

func newVar(value string, shadow bool) Variable {
	return Variable{ComplexType{TYPE_UNKNOWN, nil}, value, shadow, false, nil, 0, 0}
}
func newIndexedVar(value string, e Expression) Variable {
	return Variable{ComplexType{TYPE_UNKNOWN, nil}, value, false, true, e, 0, 0}
}
func newParam(t ComplexType, value string) Variable {
	return Variable{t, value, true, false, nil, 0, 0}
}
func newConst(t Type, value string) Constant {
	return Constant{t, value, 0, 0}
}
func newUnary(op Operator, e Expression) UnaryOp {
	return UnaryOp{op, e, ComplexType{TYPE_UNKNOWN, nil}, 0, 0}
}
func newBinary(op Operator, eLeft, eRight Expression, fixed bool) BinaryOp {
	return BinaryOp{op, eLeft, eRight, ComplexType{TYPE_UNKNOWN, nil}, fixed, 0, 0}
}
func newAssignment(variables []Variable, expressions []Expression) Assignment {
	return Assignment{variables, expressions, 0, 0}
}
func newCondition(e Expression, block, elseBlock Block) Condition {
	return Condition{e, block, elseBlock, 0, 0}
}
func newLoop(a Assignment, exprs []Expression, incrA Assignment, b Block) Loop {
	return Loop{a, exprs, incrA, b, 0, 0}
}
func newFunction(name string, params []Variable, returnTypes []ComplexType, b Block) Function {
	return Function{name, params, returnTypes, b, 0, 0}
}
func newReturn(expressions []Expression) Return {
	return Return{expressions, 0, 0}
}
func newFunCall(name string, exprs []Expression) FunCall {
	return FunCall{name, exprs, []ComplexType{}, 0, 0}
}
func newBlock(statements []Statement) Block {
	return Block{statements, SymbolTable{}, 0, 0}
}
func newAST(statements []Statement) AST {
	return AST{newBlock(statements), SymbolTable{}}
}
func newSimpleTypeList(ts []Type) (tcs []ComplexType) {
	for _, t := range ts {
		tcs = append(tcs, ComplexType{t, nil})
	}
	return
}

func TestParserExpression1(t *testing.T) {

	var code []byte = []byte(`shadow a = 6 + 7 * variable / -(5 -- (8 * - 10000.1234))`)

	expected := newAST(
		[]Statement{
			newAssignment(
				[]Variable{newVar("a", true)},
				[]Expression{
					newBinary(
						OP_PLUS, newConst(TYPE_INT, "6"), newBinary(
							OP_MULT, newConst(TYPE_INT, "7"), newBinary(
								OP_DIV, newVar("variable", false), newUnary(
									OP_NEGATIVE, newBinary(
										OP_MINUS, newConst(TYPE_INT, "5"), newUnary(
											OP_NEGATIVE, newBinary(
												OP_MULT,
												newConst(TYPE_INT, "8"),
												newUnary(OP_NEGATIVE, newConst(TYPE_FLOAT, "10000.1234")),
												false,
											),
										), false,
									),
								), false,
							), false,
						), false,
					),
				},
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserExpression2(t *testing.T) {

	var code []byte = []byte(`a = a && b || (5 < false <= 8 && (false2 > variable >= 5.0) != true)`)

	expected := newAST(
		[]Statement{
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{
					newBinary(
						OP_AND, newVar("a", false), newBinary(
							OP_OR, newVar("b", false), newBinary(
								OP_LESS, newConst(TYPE_INT, "5"), newBinary(
									OP_LE, newConst(TYPE_BOOL, "false"), newBinary(
										OP_AND, newConst(TYPE_INT, "8"), newBinary(
											OP_NE, newBinary(
												OP_GREATER,
												newVar("false2", false),
												newBinary(OP_GE, newVar("variable", false), newConst(TYPE_FLOAT, "5.0"), false),
												false,
											),
											newConst(TYPE_BOOL, "true"),
											false,
										), false,
									), false,
								), false,
							), false,
						), false,
					),
				},
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserIf(t *testing.T) {

	var code []byte = []byte(`
	if a == b {
		a = 6
	}
	a = 1
	`)

	expected := newAST(
		[]Statement{
			newCondition(
				newBinary(OP_EQ, newVar("a", false), newVar("b", false), false),
				newBlock([]Statement{newAssignment([]Variable{newVar("a", false)}, []Expression{newConst(TYPE_INT, "6")})}),
				newBlock([]Statement{}),
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newConst(TYPE_INT, "1")},
			),
		},
	)

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

	expected := newAST(
		[]Statement{
			newCondition(
				newBinary(OP_EQ, newVar("a", false), newVar("b", false), false),
				newBlock([]Statement{newAssignment([]Variable{newVar("a", false)}, []Expression{newConst(TYPE_INT, "6")})}),
				newBlock([]Statement{newAssignment(
					[]Variable{newVar("a", false)},
					[]Expression{newConst(TYPE_INT, "1")},
				)}),
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserAssignment1(t *testing.T) {

	var code []byte = []byte(`
	a = 1
	a, b = 1, 2
	a, b, c = 1, 2, 3
	`)

	expected := newAST(
		[]Statement{
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newConst(TYPE_INT, "1")},
			),
			newAssignment(
				[]Variable{newVar("a", false), newVar("b", false)},
				[]Expression{newConst(TYPE_INT, "1"), newConst(TYPE_INT, "2")},
			),
			newAssignment(
				[]Variable{newVar("a", false), newVar("b", false), newVar("c", false)},
				[]Expression{newConst(TYPE_INT, "1"), newConst(TYPE_INT, "2"), newConst(TYPE_INT, "3")},
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserAssignment2(t *testing.T) {

	var code []byte = []byte(`
	a = 1
	a++
	a--
	a *= 1
	a /= 1
	a -= 1
	a += 1
	`)

	expected := newAST(
		[]Statement{
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newConst(TYPE_INT, "1")},
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newBinary(OP_PLUS, newVar("a", false), newConst(TYPE_INT, "1"), false)},
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newBinary(OP_MINUS, newVar("a", false), newConst(TYPE_INT, "1"), false)},
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newBinary(OP_MULT, newVar("a", false), newConst(TYPE_INT, "1"), false)},
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newBinary(OP_DIV, newVar("a", false), newConst(TYPE_INT, "1"), false)},
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newBinary(OP_MINUS, newVar("a", false), newConst(TYPE_INT, "1"), false)},
			),
			newAssignment(
				[]Variable{newVar("a", false)},
				[]Expression{newBinary(OP_PLUS, newVar("a", false), newConst(TYPE_INT, "1"), false)},
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserFor1(t *testing.T) {

	var code []byte = []byte(`
	for ;; {
		a = a+1
	}
	`)

	expected := newAST(
		[]Statement{
			newLoop(
				newAssignment([]Variable{}, []Expression{}),
				[]Expression{},
				newAssignment([]Variable{}, []Expression{}),
				newBlock([]Statement{
					newAssignment(
						[]Variable{newVar("a", false)},
						[]Expression{newBinary(OP_PLUS, newVar("a", false), newConst(TYPE_INT, "1"), false)},
					),
				}),
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserFor2(t *testing.T) {

	var code []byte = []byte(`
	for i = 5;; {
		a = 0
	}
	`)

	expected := newAST(
		[]Statement{
			newLoop(
				newAssignment([]Variable{newVar("i", false)}, []Expression{newConst(TYPE_INT, "5")}),
				[]Expression{},
				newAssignment([]Variable{}, []Expression{}),
				newBlock([]Statement{
					newAssignment(
						[]Variable{newVar("a", false)},
						[]Expression{newConst(TYPE_INT, "0")},
					),
				}),
			),
		},
	)

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

	expected := newAST(
		[]Statement{
			newLoop(
				newAssignment(
					[]Variable{newVar("i", false), newVar("j", false)},
					[]Expression{newConst(TYPE_INT, "0"), newConst(TYPE_INT, "1")},
				),
				[]Expression{newBinary(OP_LESS, newVar("i", false), newConst(TYPE_INT, "10"), false)},
				newAssignment(
					[]Variable{newVar("i", false)},
					[]Expression{newBinary(OP_PLUS, newVar("i", false), newConst(TYPE_INT, "1"), false)},
				),
				newBlock([]Statement{
					newCondition(
						newBinary(OP_EQ, newVar("b", false), newVar("a", false), false),
						newBlock([]Statement{
							newLoop(
								newAssignment([]Variable{}, []Expression{}),
								[]Expression{},
								newAssignment([]Variable{}, []Expression{}),
								newBlock([]Statement{
									newAssignment(
										[]Variable{newVar("c", false)},
										[]Expression{newConst(TYPE_INT, "6")},
									),
								})),
						}),
						newBlock([]Statement{}),
					),
				}),
			),
		},
	)

	testAST(code, expected, t)
}

func TestParserFunction1(t *testing.T) {

	var code []byte = []byte(`
	fun abc() {
		a = 1
	}
	`)

	expected := newAST(
		[]Statement{
			newFunction("abc", []Variable{}, []ComplexType{}, newBlock(
				[]Statement{newAssignment([]Variable{newVar("a", false)}, []Expression{newConst(TYPE_INT, "1")})},
			)),
		},
	)

	testAST(code, expected, t)
}

func TestParserFunction2(t *testing.T) {

	var code []byte = []byte(`
	fun abc(a int, b float, c bool) {
		a = 1
		return
	}
	`)

	expected := newAST(
		[]Statement{
			newFunction("abc", []Variable{
				newParam(ComplexType{TYPE_INT, nil}, "a"),
				newParam(ComplexType{TYPE_FLOAT, nil}, "b"),
				newParam(ComplexType{TYPE_BOOL, nil}, "c"),
			}, []ComplexType{}, newBlock(
				[]Statement{newAssignment([]Variable{newVar("a", false)}, []Expression{newConst(TYPE_INT, "1")}), newReturn(nil)},
			)),
		},
	)

	testAST(code, expected, t)
}

func TestParserFunction3(t *testing.T) {

	var code []byte = []byte(`
	fun abc() int, float, bool{
		a = 1
		return a, 3.5, true
	}
	`)

	expected := newAST(
		[]Statement{
			newFunction("abc", []Variable{}, newSimpleTypeList([]Type{TYPE_INT, TYPE_FLOAT, TYPE_BOOL}), newBlock(
				[]Statement{
					newAssignment([]Variable{newVar("a", false)}, []Expression{newConst(TYPE_INT, "1")}),
					newReturn([]Expression{newVar("a", false), newConst(TYPE_FLOAT, "3.5"), newConst(TYPE_BOOL, "true")}),
				},
			)),
		},
	)

	testAST(code, expected, t)
}

func TestParserFunCall(t *testing.T) {

	var code []byte = []byte(`
	fun abc() int {
		return 1
	}
	a = abc()
	abc()
	`)

	expected := newAST(
		[]Statement{
			newFunction("abc", []Variable{}, newSimpleTypeList([]Type{TYPE_INT}), newBlock(
				[]Statement{
					newReturn([]Expression{newConst(TYPE_INT, "1")}),
				},
			)),
			newAssignment([]Variable{newVar("a", false)}, []Expression{newFunCall("abc", nil)}),
			newFunCall("abc", nil),
		},
	)

	testAST(code, expected, t)
}
