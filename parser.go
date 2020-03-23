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


Operator priority (Descending priority!):

1: 	'*', '/'
2: 	'+', '-'
3:	'==', '!=', '<=', '>=', '<', '>'
4:	'&&', '||'

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

var (
	ErrCritical = errors.New("")
	ErrNormal   = errors.New("error")
)

type SymbolEntry struct {
	sType   Type
	varName string
	// ... more information
}

type SymbolTable struct {
	table  map[string]SymbolEntry
	parent *SymbolTable
}

type AST struct {
	block             Block
	globalSymbolTable SymbolTable
}

type Type int
type Operator int

type Node interface {
	// Notes the start position in the actual source code!
	// (lineNr, columnNr)
	startPos() (int, int)
	generateCode(asm *ASM, s *SymbolTable)
}

//
// Interface types
//
type Statement interface {
	Node
	statement()
}
type Expression interface {
	Node
	expression()
	getExpressionType() Type
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// EXPRESSIONS
/////////////////////////////////////////////////////////////////////////////////////////////////

type Variable struct {
	vType        Type
	vName        string
	vShadow      bool
	line, column int
}
type Constant struct {
	cType        Type
	cValue       string
	line, column int
}
type BinaryOp struct {
	operator  Operator
	leftExpr  Expression
	rightExpr Expression
	opType    Type
	// fixed means, that the whole binary operation is in '(' ')' and should not be combined differently
	// independent on operator priority!
	fixed        bool
	line, column int
}
type UnaryOp struct {
	operator     Operator
	expr         Expression
	opType       Type
	line, column int
}

func (_ Variable) expression() {}
func (_ Constant) expression() {}
func (_ BinaryOp) expression() {}
func (_ UnaryOp) expression()  {}

func (e Variable) startPos() (int, int) {
	return e.line, e.column
}
func (e Constant) startPos() (int, int) {
	return e.line, e.column
}
func (e BinaryOp) startPos() (int, int) {
	return e.line, e.column
}
func (e UnaryOp) startPos() (int, int) {
	return e.line, e.column
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// STATEMENTS
/////////////////////////////////////////////////////////////////////////////////////////////////

type Block struct {
	statements   []Statement
	symbolTable  SymbolTable
	line, column int
}

type Assignment struct {
	variables    []Variable
	expressions  []Expression
	line, column int
}

type Condition struct {
	expression   Expression
	block        Block
	elseBlock    Block
	line, column int
}

type Loop struct {
	assignment     Assignment
	expressions    []Expression
	incrAssignment Assignment
	block          Block
	line, column   int
}

func (a Block) statement()      {}
func (a Assignment) statement() {}
func (c Condition) statement()  {}
func (l Loop) statement()       {}

func (s Block) startPos() (int, int) {
	return s.line, s.column
}
func (s Assignment) startPos() (int, int) {
	return s.line, s.column
}
func (s Condition) startPos() (int, int) {
	return s.line, s.column
}
func (s Loop) startPos() (int, int) {
	return s.line, s.column
}

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

	start, end := "", ""
	if b.fixed {
		start = "("
		end = ")"
	}
	return fmt.Sprintf("%v%v %v %v%v", start, b.leftExpr, b.operator, b.rightExpr, end)
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
		fmt.Println("Error: Channel closed unexpectedly.")
	}
	return v
}

func (tc *TokenChannel) pushBack(t Token) {
	if tc.isCached {
		fmt.Println("Error: Can only cache one item at a time.")
		return
	}
	tc.token = t
	tc.isCached = true
}

//func (tc *TokenChannel) lastLineColumn() (int, int) {
//	if !tc.isCached {
//		return -1, -1
//	}
//	return tc.token.line, tc.token.column
//}

/////////////////////////////////////////////////////////////////////////////////////////////////
// PARSER IMPLEMENTATION
/////////////////////////////////////////////////////////////////////////////////////////////////

func (c Constant) getExpressionType() Type {
	return c.cType
}
func (v Variable) getExpressionType() Type {
	return v.vType
}
func (e UnaryOp) getExpressionType() Type {
	return e.opType
}
func (e BinaryOp) getExpressionType() Type {
	return e.opType
}

// Operator priority (Descending priority!):
// 1: 	'*', '/'
// 2: 	'+', '-'
// 3:	'==', '!=', '<=', '>=', '<', '>'
// 4:	'&&', '||'
func (o Operator) priority() int {
	switch o {
	case OP_MULT, OP_DIV:
		return 1
	case OP_PLUS, OP_MINUS:
		return 2
	case OP_EQ, OP_NE, OP_LE, OP_GE, OP_LESS, OP_GREATER:
		return 3
	case OP_AND, OP_OR:
		return 4
	default:
		fmt.Printf("Unknown operator: %v\n", o)
	}
	return 100
}

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

// expectType checks the next token against a given expected type and returns the token string
// with corresponding line and column numbers
func expectType(tokens *TokenChannel, ttype TokenType) (string, int, int, bool) {
	t := tokens.next()
	if t.tokenType != ttype {
		tokens.pushBack(t)
		return "", -1, -1, false
	}
	return t.value, t.line, t.column, true
}

// expect checks the next token against a given expected type and value and returns true, if the
// check was valid.
func expect(tokens *TokenChannel, ttype TokenType, value string) (int, int, bool) {
	t := tokens.next()
	if t.tokenType != ttype || t.value != value {
		tokens.pushBack(t)
		return -1, -1, false
	}
	return t.line, t.column, true
}

func parseVariable(tokens *TokenChannel) (Variable, bool) {

	_, _, shadowing := expect(tokens, TOKEN_KEYWORD, "shadow")

	if v, row, col, ok := expectType(tokens, TOKEN_IDENTIFIER); ok {
		return Variable{TYPE_UNKNOWN, v, shadowing, row, col}, true
	}
	// Only line/column are valid!
	return Variable{TYPE_UNKNOWN, "", false, tokens.token.line, tokens.token.column}, false
}

func parseVarList(tokens *TokenChannel) (variables []Variable, err error) {
	i := 0
	for {
		v, ok := parseVariable(tokens)
		if !ok {

			// If we don't find any variable, thats fine. Just don't end in ',', thats an error!
			if i == 0 {
				return
			}
			row, col := v.startPos()
			err = fmt.Errorf("%w[%v:%v] - Variable list ends with ','", ErrCritical, row, col)
			variables = nil
			return
		}
		variables = append(variables, v)

		// Expect separating ','. Otherwise, all good, we are through!
		if _, _, ok := expect(tokens, TOKEN_SEPARATOR, ","); !ok {
			break
		}
		i += 1
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

	if v, row, col, ok := expectType(tokens, TOKEN_CONSTANT); ok {
		return Constant{getConstType(v), v, row, col}, true
	}
	return Constant{TYPE_UNKNOWN, "", tokens.token.line, tokens.token.column}, false
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
	if row, col, ok := expect(tokens, TOKEN_PARENTHESIS_OPEN, "("); ok {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%wInvalid expression in ()", ErrNormal)
			return
		}
		if tmpE, ok := e.(BinaryOp); ok {
			tmpE.fixed = true
			// We need to reassign the variable for the 'true' to hold instead of editing the local copy
			e = tmpE

		}
		expression = e

		// Expect TOKEN_PARENTHESIS_CLOSE
		row, col, ok = expect(tokens, TOKEN_PARENTHESIS_CLOSE, ")")
		if ok {
			return
		}

		err = fmt.Errorf("%w[%v:%v] - Expected ')', got something else", ErrCritical, row, col)
		return
	}

	err = fmt.Errorf("%wInvalid simple expression", ErrNormal)
	return
}

func parseUnaryExpression(tokens *TokenChannel) (expression Expression, err error) {
	// Check for unary operator before the expression
	if row, col, ok := expect(tokens, TOKEN_OPERATOR, "-"); ok {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w[%v:%v] - Invalid expression after unary '-'", ErrCritical, row, col)
			return
		}

		expression = UnaryOp{OP_NEGATIVE, e, TYPE_UNKNOWN, row, col}
		return
	}
	// Check for unary operator before the expression
	if row, col, ok := expect(tokens, TOKEN_OPERATOR, "!"); ok {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w[%v:%v] - Invalid expression after unary '!'", ErrCritical, row, col)
			return
		}

		expression = UnaryOp{OP_NOT, e, TYPE_UNKNOWN, row, col}
		return
	}

	err = fmt.Errorf("%wInvalid unary expression", ErrNormal)
	return
}

func parseExpression(tokens *TokenChannel) (expression Expression, err error) {

	unaryExpression, parseErr := parseUnaryExpression(tokens)
	if parseErr == nil {
		expression = unaryExpression
	} else {
		simpleExpression, parseErr := parseSimpleExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%wSimple expression expected, got something else", ErrNormal)
			return
		}
		expression = simpleExpression
	}

	// Or an expression followed by a binop. Here we can continue just normally and just check
	// if token.next() == binop, and just then, throw the parsed expression into a binop one.
	if t, row, col, ok := expectType(tokens, TOKEN_OPERATOR); ok {

		// Create and return binary operation expression!
		rightHandExpr, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w[%v:%v] - Invalid expression on right hand side of binary operation", ErrCritical, row, col)
			return
		}
		row, col = expression.startPos()
		finalExpression := BinaryOp{getOperatorType(t), expression, rightHandExpr, TYPE_UNKNOWN, false, row, col}
		expression = finalExpression
	}

	// We just return the simpleExpression or unaryExpression and are happy
	return
}

func parseExpressionList(tokens *TokenChannel) (expressions []Expression, err error) {

	i := 0
	lastRow, lastCol := 0, 0
	ok := false
	for {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			// If we don't find any expression, thats fine. Just don't end in ',', thats an error!
			if i == 0 {
				return
			}

			err = fmt.Errorf("%w[%v:%v] - Expression list ends in ','", ErrCritical, lastRow, lastCol)
			expressions = nil
			return
		}
		expressions = append(expressions, e)

		// Expect separating ','. Otherwise, all good, we are through!
		if lastRow, lastCol, ok = expect(tokens, TOKEN_SEPARATOR, ","); !ok {
			break
		}
		i += 1
	}
	return
}

// parseBlock parses a list of statements from the tokens.
func parseAssignment(tokens *TokenChannel) (assignment Assignment, err error) {

	// A list of variables!
	variables, parseErr := parseVarList(tokens)
	if len(variables) == 0 {
		err = fmt.Errorf("%wExpected variable in assignment, got something else", parseErr)
		return
	}
	if parseErr != nil {
		err = fmt.Errorf("%w - Parsing the variable list for an assignment resulted in an error", parseErr)
		return
	}

	// One TOKEN_ASSIGNMENT
	if row, col, ok := expect(tokens, TOKEN_ASSIGNMENT, "="); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '=' in assignment, got something else", ErrCritical, row, col)
		return
	}

	expressions, parseErr := parseExpressionList(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid expression list in assignment", parseErr)
		return
	}

	row, col := variables[0].startPos()
	assignment = Assignment{variables, expressions, row, col}
	return
}

// if ::= 'if' exp '{' [stat] '}' [else '{' [stat] '}']
func parseCondition(tokens *TokenChannel) (condition Condition, err error) {

	startRow, startCol, ok := 0, 0, false

	if startRow, startCol, ok = expect(tokens, TOKEN_KEYWORD, "if"); !ok {
		err = fmt.Errorf("%wExpected 'if' keyword for condition, got something else", ErrNormal)
		return
	}

	expression, parseErr := parseExpression(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w, %w - Expected expression after 'if' keyword", ErrCritical, parseErr)
		return
	}

	if row, col, ok := expect(tokens, TOKEN_CURLY_OPEN, "{"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '{' after condition, got something else", ErrCritical, row, col)
		return
	}

	statements, parseErr := parseStatementList(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w - Invalid statement list in condition block", parseErr)
		return
	}

	if row, col, ok := expect(tokens, TOKEN_CURLY_CLOSE, "}"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '}' after condition block, got something else", ErrCritical, row, col)
		return
	}

	condition.expression = expression
	condition.block = statements

	// Just in case we have an else, handle it!
	if _, _, ok := expect(tokens, TOKEN_KEYWORD, "else"); ok {
		if row, col, ok := expect(tokens, TOKEN_CURLY_OPEN, "{"); !ok {
			err = fmt.Errorf("%w[%v:%v] - Expected '{' after 'else' in condition, got something else", ErrCritical, row, col)
			return
		}

		elseStatements, parseErr := parseStatementList(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w - Invalid statement list in condition else block", parseErr)
			return
		}

		if row, col, ok := expect(tokens, TOKEN_CURLY_CLOSE, "}"); !ok {
			err = fmt.Errorf("%w[%v:%v] - Expected '}' after 'else' block in condition, got something else", ErrCritical, row, col)
			return
		}

		condition.elseBlock = elseStatements
	}

	condition.line = startRow
	condition.column = startCol

	return
}

func parseLoop(tokens *TokenChannel) (loop Loop, err error) {

	startRow, startCol, ok := 0, 0, false

	if startRow, startCol, ok = expect(tokens, TOKEN_KEYWORD, "for"); !ok {
		err = fmt.Errorf("%wExpected 'for' keyword for loop, got something else", ErrNormal)
		return
	}

	// We don't care about a valid assignment. If there is none, we are fine too :)
	assignment, parseErr := parseAssignment(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid assignment in loop", parseErr)
		return
	}

	if row, col, ok := expect(tokens, TOKEN_SEMICOLON, ";"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected ';' after loop assignment, got something else", ErrCritical, row, col)
		return
	}

	expressions, parseErr := parseExpressionList(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w - Invalid expression list in loop expression", parseErr)
		return
	}

	if row, col, ok := expect(tokens, TOKEN_SEMICOLON, ";"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected ';' after loop expression, got something else", ErrCritical, row, col)
		return
	}

	// We are also fine with no assignment!
	incrAssignment, parseErr := parseAssignment(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid increment assignment in loop", parseErr)
		return
	}

	if row, col, ok := expect(tokens, TOKEN_CURLY_OPEN, "{"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '{' after loop header, got something else", ErrCritical, row, col)
		return
	}

	forBlock, parseErr := parseStatementList(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w - Error while parsing loop block", parseErr)
		return
	}

	if row, col, ok := expect(tokens, TOKEN_CURLY_CLOSE, "}"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '}' after loop block, got something else", ErrCritical, row, col)
		return
	}

	loop.assignment = assignment
	loop.expressions = expressions
	loop.incrAssignment = incrAssignment
	loop.block = forBlock
	loop.line = startRow
	loop.column = startCol
	return
}

func parseStatementList(tokens *TokenChannel) (block Block, err error) {
	for {

		switch ifStatement, parseErr := parseCondition(tokens); {
		case parseErr == nil:
			block.statements = append(block.statements, ifStatement)
			continue
		case errors.Is(parseErr, ErrCritical):
			err = parseErr
			return
		}

		switch loopStatement, parseErr := parseLoop(tokens); {
		case parseErr == nil:
			block.statements = append(block.statements, loopStatement)
			continue
		case errors.Is(parseErr, ErrCritical):
			err = parseErr
			return
		}

		switch assignment, parseErr := parseAssignment(tokens); {
		case parseErr == nil:
			block.statements = append(block.statements, assignment)
			continue
		case errors.Is(parseErr, ErrCritical):
			err = parseErr
			return
		}

		// If we don't recognize the current token as part of a known statement, we break
		// This means likely, that we are at the end of a block
		break

	}

	if len(block.statements) > 0 {
		row, col := block.statements[0].startPos()
		block.line = row
		block.column = col
	} else {
		// Backup solution! As we tried parsing statements (unsuccessfully), there must be
		// a token cached. We just take its position for now
		block.line = tokens.token.line
		block.column = tokens.token.column
	}

	return
}

func parse(tokens chan Token) (ast AST, err error) {

	var tokenChan TokenChannel
	tokenChan.c = tokens

	block, parseErr := parseStatementList(&tokenChan)
	err = parseErr
	ast.block = block

	return
}
