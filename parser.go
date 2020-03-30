package main

import (
	"errors"
	"fmt"
	"regexp"
)

/*


stat	 	::= assign | if | for | funDecl | ret
statlist	::= stat [statlist]

if 			::= 'if' exp '{' [statlist] '}' [else '{' [statlist] '}']
for			::= 'for' [assign] ';' [explist] ';' [assign] '{' [statlist] '}'

type		::= 'int' | 'float' | 'string' | 'bool'
typelist	::= type [',' typelist]
paramlist	::= var type [',' paramlist]
funDecl		::= 'fun' Name '(' [paramlist] ')' [typelist] '{' [statlist] '}'

ret			::= 'return' [explist]

assign 		::= varlist ‘=’ explist
varlist		::= varDecl [‘,’ varlist]
explist		::= exp [‘,’ explist]
exp 		::= Numeral | String | var | '(' exp ')' | exp binop exp | unop exp | funCall
funCall		::= Name '(' [explist] ')'
varDecl		::= [shadow] var
var 		::= Name
binop		::= '+' | '-' | '*' | '/' | '==' | '!=' | '<=' | '>=' | '<' | '>' | '&&' | '||'
unop		::= '-' | '!'


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
	ErrNormal   = errors.New("error - ")
)

type SymbolVarEntry struct {
	sType Type
	// Refers to the name used in the final assembler
	varName string
	// ... more information
}

// This is needed when code for function calls is generated
// and we need to know how many and what kind of variables are
// pushed onto the stack or popped from afterwards.
type SymbolFunEntry struct {
	paramTypes    []Type
	returnTypes   []Type
	jumpLabel     string
	epilogueLabel string
	// ... more information
}

type SymbolTable struct {
	varTable map[string]SymbolVarEntry
	funTable map[string]SymbolFunEntry
	// activeFunctionReturn references the function return types, if we are within a function, otherwise nil
	// This is required to check validity and code generation of return statements
	activeFunctionName   string
	activeFunctionReturn []Type
	parent               *SymbolTable
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
	getExpressionTypes() []Type
	getResultCount() int
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
type FunCall struct {
	funName      string
	args         []Expression
	retTypes     []Type
	line, column int
}

func (_ Variable) expression() {}
func (_ Constant) expression() {}
func (_ BinaryOp) expression() {}
func (_ UnaryOp) expression()  {}
func (_ FunCall) expression()  {}

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
func (e FunCall) startPos() (int, int) {
	return e.line, e.column
}

func (e Constant) getExpressionTypes() []Type {
	return []Type{e.cType}
}
func (e Variable) getExpressionTypes() []Type {
	return []Type{e.vType}
}
func (e UnaryOp) getExpressionTypes() []Type {
	return []Type{e.opType}
}
func (e BinaryOp) getExpressionTypes() []Type {
	return []Type{e.opType}
}
func (e FunCall) getExpressionTypes() []Type {
	// TODO: Change return to []Type for all methods!
	return e.retTypes
}

func (e Constant) getResultCount() int {
	return 1
}
func (e Variable) getResultCount() int {
	return 1
}
func (e UnaryOp) getResultCount() int {
	return 1
}
func (e BinaryOp) getResultCount() int {
	return 1
}
func (e FunCall) getResultCount() int {
	return len(e.retTypes)
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

type Function struct {
	fName        string
	parameters   []Variable
	returnTypes  []Type
	block        Block
	line, column int
}

type Return struct {
	expressions  []Expression
	line, column int
}

func (a Block) statement()      {}
func (a Assignment) statement() {}
func (c Condition) statement()  {}
func (l Loop) statement()       {}
func (f Function) statement()   {}
func (r Return) statement()     {}

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
func (s Function) startPos() (int, int) {
	return s.line, s.column
}
func (s Return) startPos() (int, int) {
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

func (s Return) String() string {
	return fmt.Sprintf("return %v", s.expressions)
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

/////////////////////////////////////////////////////////////////////////////////////////////////
// PARSER IMPLEMENTATION
/////////////////////////////////////////////////////////////////////////////////////////////////

// Operator priority (Descending priority!):
// 0:	'-', '!'
// 1: 	'*', '/'
// 2: 	'+', '-'
// 3:	'==', '!=', '<=', '>=', '<', '>'
// 4:	'&&', '||'
func (o Operator) priority() int {
	switch o {
	case OP_NEGATIVE, OP_NOT:
		return 0
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
func (tokens *TokenChannel) expectType(ttype TokenType) (string, int, int, bool) {
	t := tokens.next()
	if t.tokenType != ttype {
		tokens.pushBack(t)
		return "", t.line, t.column, false
	}
	return t.value, t.line, t.column, true
}

// expect checks the next token against a given expected type and value and returns true, if the
// check was valid.
func (tokens *TokenChannel) expect(ttype TokenType, value string) (int, int, bool) {
	t := tokens.next()
	if t.tokenType != ttype || t.value != value {
		tokens.pushBack(t)
		return t.line, t.column, false
	}
	return t.line, t.column, true
}

func parseVariable(tokens *TokenChannel) (Variable, bool) {

	_, _, shadowing := tokens.expect(TOKEN_KEYWORD, "shadow")

	if v, row, col, ok := tokens.expectType(TOKEN_IDENTIFIER); ok {
		return Variable{TYPE_UNKNOWN, v, shadowing, row, col}, true
	}
	// Only line/column are valid!
	return Variable{}, false
}

func parseVarList(tokens *TokenChannel) (variables []Variable, err error) {
	lastRow, lastCol := 0, 0
	i := 0
	for {
		v, ok := parseVariable(tokens)
		if !ok {

			// If we don't find any variable, thats fine. Just don't end in ',', thats an error!
			// We throw a normal error, so the parser up the chain can handle it how it likes.
			if i == 0 {
				err = fmt.Errorf("%wVariable list is empty or invalid", ErrNormal)
				return
			}
			err = fmt.Errorf("%w[%v:%v] - Variable list ends with ','", ErrCritical, lastRow, lastCol)
			variables = nil
			return
		}
		variables = append(variables, v)

		// Expect separating ','. Otherwise, all good, we are through!
		if lastRow, lastCol, ok = tokens.expect(TOKEN_SEPARATOR, ","); !ok {
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

	if v, row, col, ok := tokens.expectType(TOKEN_CONSTANT); ok {
		return Constant{getConstType(v), v, row, col}, true
	}
	return Constant{TYPE_UNKNOWN, "", tokens.token.line, tokens.token.column}, false
}

// parseIdentifierExpression parses the _usage_ of variables and function calls.
// Meaning, it will not process shadow variables or declarations!
func parseIdentifierExpression(tokens *TokenChannel) (expression Expression, err error) {

	name, startRow, startCol, ok := tokens.expectType(TOKEN_IDENTIFIER)
	if !ok {
		err = fmt.Errorf("%wExpect identifier", ErrNormal)
		return
	}

	// Depending on if we find the '(' or not, this is a function call or normal variable!
	if _, _, ok := tokens.expect(TOKEN_PARENTHESIS_OPEN, "("); !ok {
		expression = Variable{TYPE_UNKNOWN, name, false, startRow, startCol}
		return
	}

	// We now parse for a function call!
	arguments, parseErr := parseExpressionList(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = parseErr
		return
	}

	if row, col, ok := tokens.expect(TOKEN_PARENTHESIS_CLOSE, ")"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Function call expects ')' after parameter list", ErrCritical, row, col)
		return
	}

	fun := FunCall{
		funName:  name,
		args:     arguments,
		retTypes: []Type{},
		line:     startRow,
		column:   startCol,
	}
	expression = fun
	return
}

// parseSimpleExpression just parses variables, constants and '('...')'
func parseSimpleExpression(tokens *TokenChannel) (expression Expression, err error) {

	// We need to check function calls first, otherwise we will parse it as a variable!
	tmpE, parseErr := parseIdentifierExpression(tokens)
	if parseErr == nil {
		expression = tmpE
		return
	}
	if errors.Is(parseErr, ErrCritical) {
		err = parseErr
		return
	}

	if tmpC, ok := parseConstant(tokens); ok {
		expression = tmpC
		return
	}

	// Or a '(', then continue until ')'.
	if row, col, ok := tokens.expect(TOKEN_PARENTHESIS_OPEN, "("); ok {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w%v", ErrCritical, parseErr.Error())
			return
		}
		if tmpE, ok := e.(BinaryOp); ok {
			tmpE.fixed = true
			// We need to reassign the variable for the 'true' to hold instead of editing the local copy
			e = tmpE

		}
		expression = e

		// Expect TOKEN_PARENTHESIS_CLOSE
		row, col, ok = tokens.expect(TOKEN_PARENTHESIS_CLOSE, ")")
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
	if row, col, ok := tokens.expect(TOKEN_OPERATOR, "-"); ok {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w[%v:%v] - Invalid expression after unary '-'", ErrCritical, row, col)
			return
		}

		expression = UnaryOp{OP_NEGATIVE, e, TYPE_UNKNOWN, row, col}
		return
	}
	// Check for unary operator before the expression
	if row, col, ok := tokens.expect(TOKEN_OPERATOR, "!"); ok {
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
			err = fmt.Errorf("%wSimple expression expected, got something else", parseErr)
			return
		}
		expression = simpleExpression
	}

	// Or an expression followed by a binop. Here we can continue just normally and just check
	// if token.next() == binop, and just then, throw the parsed expression into a binop one.
	if t, row, col, ok := tokens.expectType(TOKEN_OPERATOR); ok {

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
	lastRow, lastCol, ok := 0, 0, false
	for {
		e, parseErr := parseExpression(tokens)
		if parseErr != nil {

			// If we don't find any expression, thats fine. Just don't end in ',', thats an error!
			if i == 0 {
				// We propagate the error from the parser. This might be normal or critical.
				err = fmt.Errorf("%w - Expression list is empty or invalid", parseErr)
				return
			}

			err = fmt.Errorf("%w[%v:%v] - Expression list ends in ','", ErrCritical, lastRow, lastCol)
			expressions = nil
			return
		}
		expressions = append(expressions, e)

		// Expect separating ','. Otherwise, all good, we are through!
		if lastRow, lastCol, ok = tokens.expect(TOKEN_SEPARATOR, ","); !ok {
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
	// No variables will return an ErrNormal. So all good, severity is handled up stream.
	if len(variables) == 0 {
		err = fmt.Errorf("%wExpected variable in assignment, got something else", parseErr)
		return
	}
	// This is most likely a critical error, like: a, = ...
	if parseErr != nil {
		err = fmt.Errorf("%w - Parsing the variable list for an assignment resulted in an error", parseErr)
		return
	}

	// One TOKEN_ASSIGNMENT
	// If we got this far, we have a valid variable list. So from here on out, this _needs_ to be valid!
	if row, col, ok := tokens.expect(TOKEN_ASSIGNMENT, "="); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '=' in assignment, got something else", ErrCritical, row, col)
		return
	}

	expressions, parseErr := parseExpressionList(tokens)
	// For now we also accept an empty expression list (ErrNormal). If this is valid or not, is handled in the
	// semanticAnalyzer.
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

	if startRow, startCol, ok = tokens.expect(TOKEN_KEYWORD, "if"); !ok {
		err = fmt.Errorf("%wExpected 'if' keyword for condition, got something else", ErrNormal)
		return
	}

	expression, parseErr := parseExpression(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w%v - Expected expression after 'if' keyword", ErrCritical, parseErr.Error())
		return
	}

	if row, col, ok := tokens.expect(TOKEN_CURLY_OPEN, "{"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '{' after condition, got something else", ErrCritical, row, col)
		return
	}

	statements, parseErr := parseBlock(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w - Invalid statement list in condition block", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_CURLY_CLOSE, "}"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '}' after condition block, got something else", ErrCritical, row, col)
		return
	}

	condition.expression = expression
	condition.block = statements

	// Just in case we have an else, handle it!
	if _, _, ok := tokens.expect(TOKEN_KEYWORD, "else"); ok {
		if row, col, ok := tokens.expect(TOKEN_CURLY_OPEN, "{"); !ok {
			err = fmt.Errorf("%w[%v:%v] - Expected '{' after 'else' in condition, got something else", ErrCritical, row, col)
			return
		}

		elseStatements, parseErr := parseBlock(tokens)
		if parseErr != nil {
			err = fmt.Errorf("%w - Invalid statement list in condition else block", parseErr)
			return
		}

		if row, col, ok := tokens.expect(TOKEN_CURLY_CLOSE, "}"); !ok {
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

	if startRow, startCol, ok = tokens.expect(TOKEN_KEYWORD, "for"); !ok {
		err = fmt.Errorf("%wExpected 'for' keyword for loop, got something else", ErrNormal)
		return
	}

	// We don't care about a valid assignment. If there is none, we are fine too :)
	assignment, parseErr := parseAssignment(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid assignment in loop", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_SEMICOLON, ";"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected ';' after loop assignment, got something else", ErrCritical, row, col)
		return
	}

	expressions, parseErr := parseExpressionList(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid expression list in loop expression", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_SEMICOLON, ";"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected ';' after loop expression, got something else", ErrCritical, row, col)
		return
	}

	// We are also fine with no assignment!
	incrAssignment, parseErr := parseAssignment(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid increment assignment in loop", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_CURLY_OPEN, "{"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '{' after loop header, got something else", ErrCritical, row, col)
		return
	}

	forBlock, parseErr := parseBlock(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w - Error while parsing loop block", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_CURLY_CLOSE, "}"); !ok {
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

func parseType(tokens *TokenChannel) (t Type, err error) {
	funName, row, col, ok := tokens.expectType(TOKEN_KEYWORD)
	if !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected a type", ErrCritical, row, col)
		return
	}
	switch funName {
	case "int":
		t = TYPE_INT
	case "float":
		t = TYPE_FLOAT
	case "bool":
		t = TYPE_BOOL
	case "string":
		t = TYPE_STRING
	default:
		err = fmt.Errorf("%w[%v:%v] - Not a valid type. Should be: 'int' | 'float' | 'bool' | 'string'", ErrCritical, row, col)
		return
	}
	return
}
func parseArgList(tokens *TokenChannel) (variables []Variable, err error) {
	i := 0
	lastRow, lastCol := 0, 0
	for {
		v, ok := parseVariable(tokens)
		if !ok {
			// If we don't find any variable, thats fine. Just don't end in ',', thats an error!
			if i == 0 {
				// We propagate the error from the parser. This might be normal or critical.
				err = fmt.Errorf("%wNot a variable", ErrNormal)
				return
			}

			err = fmt.Errorf("%w[%v:%v] - Variable list ends in ','", ErrCritical, lastRow, lastCol)
			variables = nil
			return
		}

		t, parseErr := parseType(tokens)
		if parseErr != nil {
			err = parseErr
			return
		}
		v.vType = t

		// Function parameters are always shadowing!
		v.vShadow = true
		variables = append(variables, v)

		// Expect separating ','. Otherwise, all good, we are through!
		if lastRow, lastCol, ok = tokens.expect(TOKEN_SEPARATOR, ","); !ok {
			break
		}
		i += 1
	}
	return
}

func parseTypeList(tokens *TokenChannel) (types []Type, err error) {
	lastRow, lastCol, ok := 0, 0, false
	i := 0
	for {
		t, parseErr := parseType(tokens)
		if parseErr != nil {

			// If we don't find any type, thats fine. Just don't end in ',', thats an error!
			// We throw a normal error, so the parser up the chain can handle it how it likes.
			if i == 0 {
				err = fmt.Errorf("%wType list is empty or invalid", ErrNormal)
				return
			}
			err = fmt.Errorf("%w[%v:%v] - Type list ends with ','", ErrCritical, lastRow, lastCol)
			types = nil
			return
		}
		types = append(types, t)

		// Expect separating ','. Otherwise, all good, we are through!
		if lastRow, lastCol, ok = tokens.expect(TOKEN_SEPARATOR, ","); !ok {
			break
		}
		i += 1
	}
	return
}

func parseFunction(tokens *TokenChannel) (fun Function, err error) {
	startRow, startCol, ok := 0, 0, false
	var parseErr error = nil

	if startRow, startCol, ok = tokens.expect(TOKEN_KEYWORD, "fun"); !ok {
		err = fmt.Errorf("%wExpected 'fun' keyword for function", ErrNormal)
		return
	}

	funName, row, col, ok := tokens.expectType(TOKEN_IDENTIFIER)
	if !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected identifier after 'fun'", ErrCritical, row, col)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_PARENTHESIS_OPEN, "("); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '(' for function parameters", ErrCritical, row, col)
		return
	}

	variables, parseErr := parseArgList(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid argument list in function header", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_PARENTHESIS_CLOSE, ")"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected ')' after function parameters", ErrCritical, row, col)
		return
	}

	returnTypes, parseErr := parseTypeList(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid return-type list", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_CURLY_OPEN, "{"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '{' after function header", ErrCritical, row, col)
		return
	}

	funBlock, parseErr := parseBlock(tokens)
	if parseErr != nil {
		err = fmt.Errorf("%w - Error while parsing loop block", parseErr)
		return
	}

	if row, col, ok := tokens.expect(TOKEN_CURLY_CLOSE, "}"); !ok {
		err = fmt.Errorf("%w[%v:%v] - Expected '}' after function block", ErrCritical, row, col)
		return
	}

	fun.fName = funName
	fun.parameters = variables
	fun.returnTypes = returnTypes
	fun.block = funBlock
	fun.line = startRow
	fun.column = startCol

	return
}

func parseReturn(tokens *TokenChannel) (ret Return, err error) {
	startRow, startCol, ok := 0, 0, false
	if startRow, startCol, ok = tokens.expect(TOKEN_KEYWORD, "return"); !ok {
		err = fmt.Errorf("%wExpected 'return' keyword", ErrNormal)
		return
	}

	expressions, parseErr := parseExpressionList(tokens)
	if errors.Is(parseErr, ErrCritical) {
		err = fmt.Errorf("%w - Invalid expression list in return statement", parseErr)
		return
	}
	ret.expressions = expressions
	ret.line = startRow
	ret.column = startCol
	return
}

func parseBlock(tokens *TokenChannel) (block Block, err error) {
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

		switch function, parseErr := parseFunction(tokens); {
		case parseErr == nil:
			block.statements = append(block.statements, function)
			continue
		case errors.Is(parseErr, ErrCritical):
			err = parseErr
			return
		}

		switch ret, parseErr := parseReturn(tokens); {
		case parseErr == nil:
			block.statements = append(block.statements, ret)
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

	block, parseErr := parseBlock(&tokenChan)
	err = parseErr
	ast.block = block

	return
}
