package main

import (
	"errors"
	"fmt"
	"regexp"
)

/*


stat 	::= assign | if | for

if 		::= 'if' exp '{' [stat] '}' [else '{' [stat] '}']
for		::= 'for' [assign] ';' explist ';' [assign] '{' [stat] '}'


assign 	::= varlist ‘=’ explist
varlist	::= var {‘,’ var}
explist	::= exp {‘,’ exp}
exp 	::= Numeral | String | var | '(' exp ')' | exp binop exp | unop exp
var 	::= [shadow] Name
binop	::= '+' | '-' | '*' | '/' | '==' | '!=' | '<=' | '>=' | '<' | '>' | '&&' | '||'
unop	::= '-' | '!'


*/

/////////////////////////////////////////////////////////////////////////////////////////////////
// CONST
/////////////////////////////////////////////////////////////////////////////////////////////////

const (
	TYPE_INT = iota
	TYPE_STRING
	TYPE_FLOAT
	TYPE_BOOL
	// TYPE_FUNCTION ?
	TYPE_UNKNOWN
)
const (
	OP_PLUS = iota
	OP_MINUS
	OP_MULT
	OP_DIV

	OP_NEGATIVE
	OP_NOT

	OP_EQ
	OP_NE
	OP_LE
	OP_GE
	OP_LESS
	OP_GREATER

	OP_AND
	OP_OR

	OP_UNKNOWN
)

/////////////////////////////////////////////////////////////////////////////////////////////////
// INTERFACES
/////////////////////////////////////////////////////////////////////////////////////////////////

type SymbolEntry struct {
	sType      Type
	sShadowing bool
	// ... more information
}

type SymbolTable map[string]SymbolEntry

type AST struct {
	block             Block
	globalSymbolTable SymbolTable
}

type Type int
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

type Variable struct {
	vType   Type
	vName   string
	vShadow bool
}
type Constant struct {
	cType  Type
	cValue string
}
type BinaryOp struct {
	operator  Operator
	leftExpr  Expression
	rightExpr Expression
	opType    Type
}
type UnaryOp struct {
	operator Operator
	expr     Expression
	opType   Type
}

func (_ Variable) expression() {}
func (_ Constant) expression() {}
func (_ BinaryOp) expression() {}
func (_ UnaryOp) expression()  {}

/////////////////////////////////////////////////////////////////////////////////////////////////
// STATEMENTS
/////////////////////////////////////////////////////////////////////////////////////////////////

type Block struct {
	statements        []Statement
	parentSymbolTable SymbolTable
	symbolTable       SymbolTable
}

type Assignment struct {
	variables   []Variable
	expressions []Expression
}

type Condition struct {
	expression Expression
	block      Block
	elseBlock  Block
}

type Loop struct {
	assignment     Assignment
	expressions    []Expression
	incrAssignment Assignment
	block          Block
}

func (a Block) statement()      {}
func (a Assignment) statement() {}
func (c Condition) statement()  {}
func (l Loop) statement()       {}

/////////////////////////////////////////////////////////////////////////////////////////////////
// AST, OPS STRING
/////////////////////////////////////////////////////////////////////////////////////////////////

func (ast AST) String() string {
	s := fmt.Sprintln("AST:")

	for _, st := range ast.block.statements {
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
	case OP_AND:
		return "&&"
	case OP_OR:
		return "||"
	case OP_NOT:
		return "!"
	case OP_UNKNOWN:
		return "?"
	}
	return "?"
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// EXPRESSION STRING
/////////////////////////////////////////////////////////////////////////////////////////////////

func (v Variable) String() string {
	shadowString := ""
	if v.vShadow {
		shadowString = "shadow "
	}
	return fmt.Sprintf("%v%v(%v)", shadowString, v.vType, v.vName)
}
func (c Constant) String() string {
	return fmt.Sprintf("%v(%v)", c.cType, c.cValue)
}
func (b BinaryOp) String() string {
	return fmt.Sprintf("%v %v %v", b.leftExpr, b.operator, b.rightExpr)
}
func (u UnaryOp) String() string {
	return fmt.Sprintf("%v(%v)", u.operator, u.expr)
}

func (v Type) String() string {
	switch v {
	case TYPE_INT:
		return "int"
	case TYPE_STRING:
		return "string"
	case TYPE_FLOAT:
		return "float"
	case TYPE_BOOL:
		return "bool"
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

	for _, st := range c.block.statements {
		s += fmt.Sprintf("\t%v\n", st)
	}

	s += "}"

	if c.elseBlock.statements != nil {
		s += " else {\n"
		for _, st := range c.elseBlock.statements {
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

	for _, st := range l.block.statements {
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
	switch o {
	case "+":
		return OP_PLUS
	case "-":
		return OP_MINUS
	case "*":
		return OP_MULT
	case "/":
		return OP_DIV
	case "==":
		return OP_EQ
	case "!=":
		return OP_NE
	case "<=":
		return OP_LE
	case ">=":
		return OP_GE
	case "<":
		return OP_LESS
	case ">":
		return OP_GREATER
	case "&&":
		return OP_AND
	case "||":
		return OP_OR
	case "!":
		return OP_NOT

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

	shadowing := expect(tokens, TOKEN_KEYWORD, "shadow")

	if v, ok := expectType(tokens, TOKEN_IDENTIFIER); ok {
		return Variable{TYPE_UNKNOWN, v, shadowing}, true
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

func getConstType(c string) Type {
	rFloat := regexp.MustCompile(`^(-?\d+\.\d*)`)
	rInt := regexp.MustCompile(`^(-?\d+)`)
	rString := regexp.MustCompile(`^(".*")`)
	rBool := regexp.MustCompile(`^(true|false)`)
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
	if s := rBool.FindIndex(cByte); s != nil {
		return TYPE_BOOL
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
func parseSimpleExpression(tokens *TokenChannel) (expression Expression, err error) {
	// Expect either a constant/variable and you're done
	if tmpV, ok := parseVariable(tokens); ok {
		expression = tmpV
		return
	}

	if tmpC, ok := parseConstant(tokens); ok {
		expression = tmpC
		return
	}

	// Or a '(', then continue until ')'. Parenthesis are not included in the AST, as they are implicit!
	if expect(tokens, TOKEN_PARENTHESIS_OPEN, "(") {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = errors.New(fmt.Sprintf("Invalid expression in ()"))
			return
		}
		expression = e

		// Expect TOKEN_PARENTHESIS_CLOSE
		if expect(tokens, TOKEN_PARENTHESIS_CLOSE, ")") {
			return
		}

		err = errors.New(fmt.Sprintf("Expected ')', got something else"))
		return
	}

	err = errors.New(fmt.Sprintf("Invalid simple expression"))
	return
}

func parseUnaryExpression(tokens *TokenChannel) (expression Expression, err error) {
	// Check for unary operator before the expression
	if expect(tokens, TOKEN_OPERATOR, "-") {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = errors.New(fmt.Sprintf("Invalid expression after unary '-'"))
			return
		}

		expression = UnaryOp{OP_NEGATIVE, e, TYPE_UNKNOWN}
		return
	}
	// Check for unary operator before the expression
	if expect(tokens, TOKEN_OPERATOR, "!") {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = errors.New(fmt.Sprintf("Invalid expression after unary '!'"))
			return
		}

		expression = UnaryOp{OP_NOT, e, TYPE_UNKNOWN}
		return
	}

	err = errors.New(fmt.Sprintf("Invalid unary expression"))
	return
}

func parseExpression(tokens *TokenChannel) (expression Expression, err error) {

	unaryExpression, parseErr := parseUnaryExpression(tokens)
	if parseErr == nil {
		expression = unaryExpression
	} else {
		simpleExpression, parseErr := parseSimpleExpression(tokens)
		if parseErr != nil {
			err = errors.New(fmt.Sprintf("Simple expression expected, got something else"))
			return
		}
		expression = simpleExpression
	}

	// Or an expression followed by a binop. Here we can continue just normally and just check
	// if token.next() == binop, and just then, throw the parsed expression into a binop one.
	if t, ok := expectType(tokens, TOKEN_OPERATOR); ok {

		// Create and return binary operation expression!
		rightHandExpr, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = errors.New(fmt.Sprintf("Invalid expression on right hand side of binary operation"))
			return
		}
		finalExpression := BinaryOp{getOperatorType(t), expression, rightHandExpr, TYPE_UNKNOWN}
		expression = finalExpression
	}

	// We just return the simpleExpression or unaryExpression and are happy
	return
}

func parseExpressionList(tokens *TokenChannel) (expressions []Expression, err error) {

	i := 0
	for {
		e, parseErr := parseExpression(tokens)

		if parseErr != nil {
			// If we don't find any expression, thats fine. Just don't end in ',', thats an error!
			if i == 0 {
				return
			}

			err = errors.New(fmt.Sprintf("Expected expression in expression list after ',', got something else"))
			expressions = nil
			return
		}
		expressions = append(expressions, e)

		// Expect separating ','. Otherwise, all good, we are through!
		if !expect(tokens, TOKEN_SEPARATOR, ",") {
			break
		}
		i += 1
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
	// TODO: Ignore an error here, because we don't care about no expressions!
	// We will check that later though!
	expressions, parseErr := parseExpressionList(tokens)
	if parseErr != nil {
		fmt.Println("Invalid expression list -- ", parseErr)
		allOK = false
		return
	}

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

	expression, parseErr := parseExpression(tokens)
	if parseErr != nil {
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

	// TODO: Ignore an error here, because we don't care about no expressions!
	expressionList, parseErr := parseExpressionList(tokens)
	if parseErr != nil {
		fmt.Println("Invalid expression list -- ", parseErr)
		allOK = false
		return
	}

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

func parseStatementList(tokens *TokenChannel) (block Block) {
	for {
		ifStatement, ok := parseCondition(tokens)
		if ok {
			block.statements = append(block.statements, ifStatement)
			continue
		}

		loopStatement, ok := parseLoop(tokens)
		if ok {
			block.statements = append(block.statements, loopStatement)
			continue
		}

		assignment, ok := parseAssignment(tokens)
		if ok {
			block.statements = append(block.statements, assignment)
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
