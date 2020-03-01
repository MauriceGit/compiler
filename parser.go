package main

import (
	"fmt"
	"regexp"
)

/*


stat 	::= assign | if | for

if 		::= 'if' exp '{' [stat] '}' [else '{' [stat] '}']
for		::= 'for' varlist '=' explist ';' exp [',' exp] ';' [assign] '{' [stat] '}'


assign 	::= varlist ‘=’ explist
varlist	::= var {‘,’ var}
explist	::= exp {‘,’ exp}
exp 	::= Numeral | String | var | '(' exp ')' | exp binop exp | unop exp
var 	::= Name
binop	::= '+' | '-' | '*' | '/'
unop	::= '-'


*/

/////////////////////////////////////////////////////////////////////////////////////////////////
// CONST
/////////////////////////////////////////////////////////////////////////////////////////////////

const (
	TYPE_INT = iota
	TYPE_STRING
	TYPE_FLOAT
	TYPE_UNKNOWN
)
const (
	OP_PLUS = iota
	OP_MINUS
	OP_MULT
	OP_DIV

	OP_NEGATIVE

	OP_EQ
	OP_NE
	OP_LE
	OP_GE
	OP_LESS
	OP_GREATER

	OP_UNKNOWN
)

/////////////////////////////////////////////////////////////////////////////////////////////////
// INTERFACES
/////////////////////////////////////////////////////////////////////////////////////////////////

type AST struct {
	block []Statement
}

type VarType int
type Operator int

type Node interface {
	// Notes the start position in the actual source code!
	// (lineNr, columnNr)
	Start() (int, int)
}

//
// Interface types
//
type Statement interface {
	//Node
	statement()
}
type Expression interface {
	//Node
	expression()
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// EXPRESSIONS
/////////////////////////////////////////////////////////////////////////////////////////////////

//
// Expression types
//
type Variable struct {
	varType  VarType
	varName  string
	varValue string
}
type Constant struct {
	constType  VarType
	constValue string
}
type BinaryOp struct {
	opType    Operator
	leftExpr  Expression
	rightExpr Expression
}
type UnaryOp struct {
	opType Operator
	expr   Expression
}

func (_ Variable) expression() {}
func (_ Constant) expression() {}
func (_ BinaryOp) expression() {}
func (_ UnaryOp) expression()  {}

/////////////////////////////////////////////////////////////////////////////////////////////////
// STATEMENTS
/////////////////////////////////////////////////////////////////////////////////////////////////

//
// Statement types
//
type Assignment struct {
	variables   []Variable
	expressions []Expression
}

type Condition struct {
	expression Expression
	block      []Statement
	elseBlock  []Statement
}

// for ::= 'for' varlist '=' explist ';' exp [',' exp] '{' [stat] '}'
type Loop struct {
	assignment     Assignment
	expressions    []Expression
	incrAssignment Assignment
	block          []Statement
}

func (a Assignment) statement() {}
func (c Condition) statement()  {}
func (l Loop) statement()       {}

/////////////////////////////////////////////////////////////////////////////////////////////////
// AST, OPS STRING
/////////////////////////////////////////////////////////////////////////////////////////////////

func (ast AST) String() string {
	s := ""
	s += fmt.Sprintln("AST:")

	for _, st := range ast.block {
		s += fmt.Sprintf("%v\n", st)
	}
	return s
}

func (o Operator) String() string {
	switch o {
	case OP_PLUS:
		return "+"
	case OP_MINUS:
		return "-"
	case OP_MULT:
		return "*"
	case OP_DIV:
		return "/"
	case OP_NEGATIVE:
		return "-"
	case OP_EQ:
		return "=="
	case OP_NE:
		return "!="
	case OP_LE:
		return "<="
	case OP_GE:
		return ">="
	case OP_LESS:
		return "<"
	case OP_GREATER:
		return ">"
	case OP_UNKNOWN:
		return "?"
	}
	return "?"
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// EXPRESSION STRING
/////////////////////////////////////////////////////////////////////////////////////////////////

func (v Variable) String() string {
	return fmt.Sprintf("%v", v.varName)
}
func (c Constant) String() string {
	return fmt.Sprintf("%v(%v)", c.constType, c.constValue)
}
func (b BinaryOp) String() string {
	return fmt.Sprintf("%v %v %v", b.leftExpr, b.opType, b.rightExpr)
}
func (u UnaryOp) String() string {
	return fmt.Sprintf("%v%v", u.opType, u.expr)
}

func (v VarType) String() string {
	switch v {
	case TYPE_INT:
		return "int"
	case TYPE_STRING:
		return "string"
	case TYPE_FLOAT:
		return "float"
	}
	return "?"
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// ASSIGNMENT STRING
/////////////////////////////////////////////////////////////////////////////////////////////////

func (a Assignment) String() (s string) {

	for i, v := range a.variables {
		s += fmt.Sprintf("%v", v)
		if i != len(a.variables)-1 {
			s += fmt.Sprintf(", ")
		}
	}
	s += fmt.Sprintf(" = ")

	for i, v := range a.expressions {
		s += fmt.Sprintf("%v", v)
		if i != len(a.expressions)-1 {
			s += fmt.Sprintf(", ")
		}
	}

	return
}

func (c Condition) String() (s string) {

	s += fmt.Sprintf("if %v {\n", c.expression)

	for _, st := range c.block {
		s += fmt.Sprintf("\t%v\n", st)
	}

	s += "}"

	if c.elseBlock != nil {
		s += " else {\n"
		for _, st := range c.elseBlock {
			s += fmt.Sprintf("\t%v\n", st)
		}
		s += "}"
	}
	return
}

func (l Loop) String() (s string) {

	s += fmt.Sprintf("for %v; ", l.assignment)

	for i, e := range l.expressions {
		s += fmt.Sprintf("%v", e)
		if i != len(l.expressions)-1 {
			s += ", "
		}
	}

	s += fmt.Sprintf("; %v", l.incrAssignment)

	s += " {\n"

	for _, st := range l.block {
		s += fmt.Sprintf("\t%v\n", st)
	}

	s += "}"
	return
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// TOKEN CHANNEL
/////////////////////////////////////////////////////////////////////////////////////////////////

// Implements a channel with one cache/lookahead, that can be pushed back in (logically)
type TokenChannel struct {
	c        chan Token
	isCached bool
	token    Token
}

func (tc *TokenChannel) next() Token {
	if tc.isCached {
		tc.isCached = false
		return tc.token
	}
	v, ok := <-tc.c
	if !ok {
		fmt.Println("Channel closed unexpectedly.")
	}
	return v
}

func (tc *TokenChannel) pushBack(t Token) {
	if tc.isCached {
		fmt.Println("Can only cache one item at a time.")
		return
	}
	tc.token = t
	tc.isCached = true
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// PARSER IMPLEMENTATION
/////////////////////////////////////////////////////////////////////////////////////////////////

func getOperatorType(o string) Operator {
	switch {
	case o == "+":
		return OP_PLUS
	case o == "-":
		return OP_MINUS
	case o == "*":
		return OP_MULT
	case o == "/":
		return OP_DIV
	case o == "==":
		return OP_EQ
	case o == "!=":
		return OP_NE
	case o == "<=":
		return OP_LE
	case o == ">=":
		return OP_GE
	case o == "<":
		return OP_LESS
	case o == ">":
		return OP_GREATER

	}
	return OP_UNKNOWN
}

func expectType(tokens *TokenChannel, ttype TokenType) (string, bool) {
	t := tokens.next()
	if t.tokenType != ttype {
		tokens.pushBack(t)
		return "", false
	}
	return t.value, true
}

func expect(tokens *TokenChannel, ttype TokenType, value string) bool {
	t := tokens.next()
	if t.tokenType != ttype || t.value != value {
		tokens.pushBack(t)
		return false
	}
	return true
}

func parseVariable(tokens *TokenChannel) (Variable, bool) {

	if v, ok := expectType(tokens, TOKEN_IDENTIFIER); ok {
		return Variable{TYPE_UNKNOWN, v, ""}, true
	}
	return Variable{}, false
}

func parseVarList(tokens *TokenChannel) (variables []Variable) {
	for {
		v, ok := parseVariable(tokens)
		if !ok {
			variables = nil
			break
		}
		variables = append(variables, v)

		// Expect separating ','. Otherwise, all good, we are through!
		if !expect(tokens, TOKEN_SEPARATOR, ",") {
			break
		}

	}
	return
}

func getConstType(c string) VarType {
	rFloat := regexp.MustCompile(`^(-?\d+\.\d*)`)
	rInt := regexp.MustCompile(`^(-?\d+)`)
	rString := regexp.MustCompile(`^(".*")`)
	cByte := []byte(c)

	if s := rFloat.FindIndex(cByte); s != nil {
		return TYPE_FLOAT
	}
	if s := rInt.FindIndex(cByte); s != nil {
		return TYPE_INT
	}
	if s := rString.FindIndex(cByte); s != nil {
		return TYPE_STRING
	}
	return TYPE_UNKNOWN
}

func parseConstant(tokens *TokenChannel) (Constant, bool) {

	if v, ok := expectType(tokens, TOKEN_CONSTANT); ok {
		return Constant{getConstType(v), v}, true
	}
	return Constant{}, false
}

// parseSimpleExpression just parses variables, constants and '('...')'
func parseSimpleExpression(tokens *TokenChannel) (expression Expression, allOK bool) {
	// Expect either a constant/variable and you're done
	if tmpV, ok := parseVariable(tokens); ok {
		expression = tmpV
		allOK = true
		return
	}
	if tmpC, ok := parseConstant(tokens); ok {
		expression = tmpC
		allOK = true
		return
	}

	// Or a '(', then continue until ')'. Parenthesis are not included in the AST, as they are implicit!
	if expect(tokens, TOKEN_PARENTHESIS_OPEN, "(") {
		e, ok := parseExpression(tokens, true)
		if !ok {
			fmt.Printf("Invalid expression in ()\n")
			return
		}
		expression = e
		allOK = true

		// Expect TOKEN_PARENTHESIS_CLOSE
		if expect(tokens, TOKEN_PARENTHESIS_CLOSE, ")") {
			return
		}

		allOK = false
		fmt.Printf("Expected ')', got something else!")
		return
	}

	return
}

func parseUnaryExpression(tokens *TokenChannel) (expression Expression, allOK bool) {
	// Check for unary operator before the expression
	if expect(tokens, TOKEN_OPERATOR, "-") {
		e, ok := parseExpression(tokens, true)
		if !ok {
			fmt.Printf("Invalid Expression.\n")
			return
		}

		expression = UnaryOp{OP_NEGATIVE, e}
		allOK = true
		return
	}
	return
}

func parseExpression(tokens *TokenChannel, required bool) (expression Expression, allOK bool) {

	unaryExpression, ok := parseUnaryExpression(tokens)
	if ok {
		expression = unaryExpression
	} else {
		simpleExpression, ok := parseSimpleExpression(tokens)
		if !ok {
			if required {
				fmt.Printf("Invalid but expected Simple Expression!\n")
			}
			return
		}
		expression = simpleExpression
	}

	// Or an expression followed by a binop. Here we can continue just normally and just check
	// if token.next() == binop, and just then, throw the parsed expression into a binop one.
	if t, ok := expectType(tokens, TOKEN_OPERATOR); ok {

		// Create and return binary operation expression!
		rightHandExpr, ok := parseExpression(tokens, true)
		if !ok {
			fmt.Printf("Invalid expression on righthand side of binary operation!\n")
			return
		}
		finalExpression := BinaryOp{getOperatorType(t), expression, rightHandExpr}
		expression = finalExpression
	}

	// We just return the simpleExpression or unaryExpression and are happy
	allOK = true
	return
}

func parseExpressionList(tokens *TokenChannel, required bool) (expressions []Expression) {
	for {
		e, ok := parseExpression(tokens, required)
		if !ok {
			if required {
				fmt.Println("Expected expression, got something else!")
			}
			expressions = nil
			break
		}
		expressions = append(expressions, e)

		// Expect separating ','. Otherwise, all good, we are through!
		if !expect(tokens, TOKEN_SEPARATOR, ",") {
			break
		}
	}
	return
}

// parseBlock parses a list of statements from the tokens.
func parseAssignment(tokens *TokenChannel) (assignment Assignment, allOK bool) {

	// We expect:
	// TODO: Block begin

	// A list of variables!
	variables := parseVarList(tokens)
	if len(variables) == 0 {
		return
	}

	// One TOKEN_ASSIGNMENT
	if !expect(tokens, TOKEN_ASSIGNMENT, "=") {
		return
	}

	// A list of expressions
	expressions := parseExpressionList(tokens, false)

	assignment = Assignment{variables, expressions}
	allOK = true
	return

	// TODO: Block end
}

// if ::= 'if' exp '{' [stat] '}' [else '{' [stat] '}']
func parseCondition(tokens *TokenChannel) (condition Condition, allOK bool) {

	if !expect(tokens, TOKEN_KEYWORD, "if") {
		return
	}

	expression, ok := parseExpression(tokens, true)
	if !ok {
		fmt.Println("Expected expression after 'if' but got something else")
		return
	}

	if !expect(tokens, TOKEN_CURLY_OPEN, "{") {
		fmt.Println("Expected '{' after if but got something else")
		return
	}

	statements := parseStatementList(tokens)

	if !expect(tokens, TOKEN_CURLY_CLOSE, "}") {
		fmt.Println("Expected '}' after if block, got something else")
		return
	}

	condition.expression = expression
	condition.block = statements

	// Just in case we have an else, handle it!
	if expect(tokens, TOKEN_KEYWORD, "else") {
		if !expect(tokens, TOKEN_CURLY_OPEN, "{") {
			fmt.Println("Expected '{' after 'else', got something else")
			return
		}

		elseStatements := parseStatementList(tokens)

		if !expect(tokens, TOKEN_CURLY_CLOSE, "}") {
			fmt.Println("Expected '}' after 'else' block, got something else")
			return
		}

		condition.elseBlock = elseStatements
	}

	allOK = true
	return

}

// for ::= 'for' varlist '=' explist ';' exp [',' exp] '{' [stat] '}'
func parseLoop(tokens *TokenChannel) (loop Loop, allOK bool) {

	if !expect(tokens, TOKEN_KEYWORD, "for") {
		return
	}

	// We don't care about a valid assignment. If there is none, we are fine too :)
	assignment, ok := parseAssignment(tokens)

	if !expect(tokens, TOKEN_SEMICOLON, ";") {
		fmt.Println("Expected ';' after 'for' assignment, got something else")
		return
	}

	expressionList := parseExpressionList(tokens, false)

	if !expect(tokens, TOKEN_SEMICOLON, ";") {
		fmt.Println("Expected ';' after 'for' expression, got something else")
		return
	}

	// We are also fine with no assignment!
	incrAssignment, ok := parseAssignment(tokens)
	if ok {
		loop.incrAssignment = incrAssignment
	}

	if !expect(tokens, TOKEN_CURLY_OPEN, "{") {
		fmt.Println("Expected '{' after 'for', got something else")
		return
	}

	forBlock := parseStatementList(tokens)

	if !expect(tokens, TOKEN_CURLY_CLOSE, "}") {
		fmt.Println("Expected '}' after 'for' block, got something else")
		return
	}

	loop.assignment = assignment
	loop.expressions = expressionList
	loop.block = forBlock
	allOK = true
	return

}

func parseStatementList(tokens *TokenChannel) (statements []Statement) {
	for {
		ifStatement, ok := parseCondition(tokens)
		if ok {
			statements = append(statements, ifStatement)
			continue
		}

		loopStatement, ok := parseLoop(tokens)
		if ok {
			statements = append(statements, loopStatement)
			continue
		}

		assignment, ok := parseAssignment(tokens)
		if ok {
			statements = append(statements, assignment)
			continue
		}

		// If we don't recognize the current token as part of a known statement, we break
		// This means likely, that we are at the end of a block
		break

	}
	return
}

func parse(tokens chan Token) AST {

	var tokenChan TokenChannel
	tokenChan.c = tokens

	var ast AST
	ast.block = parseStatementList(&tokenChan)

	return ast
}
