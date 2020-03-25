package main

import (
	"testing"
)

func tokenKeyword(s string) Token {
	return Token{TOKEN_KEYWORD, s, 0, 0}
}
func tokenAssignment() Token {
	return Token{TOKEN_ASSIGNMENT, "=", 0, 0}
}
func tokenConstant(s string) Token {
	return Token{TOKEN_CONSTANT, s, 0, 0}
}
func tokenCurlyOpen() Token {
	return Token{TOKEN_CURLY_OPEN, "{", 0, 0}
}
func tokenCurlyClose() Token {
	return Token{TOKEN_CURLY_CLOSE, "}", 0, 0}
}
func tokenSemicolon() Token {
	return Token{TOKEN_SEMICOLON, ";", 0, 0}
}
func tokenIdentifier(s string) Token {
	return Token{TOKEN_IDENTIFIER, s, 0, 0}
}
func tokenOperator(s string) Token {
	return Token{TOKEN_OPERATOR, s, 0, 0}
}
func tokenParenOpen() Token {
	return Token{TOKEN_PARENTHESIS_OPEN, "(", 0, 0}
}
func tokenParenClose() Token {
	return Token{TOKEN_PARENTHESIS_CLOSE, ")", 0, 0}
}
func tokenSeparator() Token {
	return Token{TOKEN_SEPARATOR, ",", 0, 0}
}
func tokenEOF() Token {
	return Token{TOKEN_EOF, "", 0, 0}
}

func testChannelEqualSlice(tokens chan Token, expected []Token) (bool, int, Token, Token) {
	for i, token := range expected {
		if t := <-tokens; t.tokenType != token.tokenType || t.value != token.value {
			return false, i, token, t
		}
	}
	return true, 0, Token{}, Token{}
}

func testTokens(code []byte, expect []Token, t *testing.T) {
	tokenChan := make(chan Token, 100)
	lexerErr := make(chan error, 1)
	go tokenize(code, tokenChan, lexerErr)

	select {
	case e := <-lexerErr:
		t.Errorf("%v", e.Error())
		return
	default:
	}

	if ok, i, te, tg := testChannelEqualSlice(tokenChan, expect); !ok {
		t.Errorf("Expected %v, got %v at position %v\n", te, tg, i)
	}

	if len(tokenChan) != 0 {
		t.Errorf("%v tokens expected, got %v\n", len(expect), len(expect)+len(tokenChan))
	}
}

func TestLexerExpression1(t *testing.T) {

	var code []byte = []byte(`6 + 7 * variable / -(5 -- (-8 * - 10000.1234))`)

	expect := []Token{tokenConstant("6"), tokenOperator("+"), tokenConstant("7"), tokenOperator("*"),
		tokenIdentifier("variable"), tokenOperator("/"), tokenOperator("-"), tokenParenOpen(), tokenConstant("5"),
		tokenOperator("-"), tokenOperator("-"), tokenParenOpen(), tokenConstant("-8"), tokenOperator("*"),
		tokenOperator("-"), tokenConstant("10000.1234"), tokenParenClose(), tokenParenClose(), tokenEOF(),
	}

	testTokens(code, expect, t)

}

func TestLexerExpression2(t *testing.T) {

	var code []byte = []byte(`a && b || (5 < false <= 8 && (false2 > variable >= 5.0) != true)`)

	expect := []Token{tokenIdentifier("a"), tokenOperator("&&"), tokenIdentifier("b"), tokenOperator("||"),
		tokenParenOpen(), tokenConstant("5"), tokenOperator("<"), tokenConstant("false"), tokenOperator("<="),
		tokenConstant("8"), tokenOperator("&&"), tokenParenOpen(), tokenIdentifier("false2"), tokenOperator(">"),
		tokenIdentifier("variable"), tokenOperator(">="), tokenConstant("5.0"), tokenParenClose(), tokenOperator("!="),
		tokenConstant("true"), tokenParenClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerIf(t *testing.T) {

	var code []byte = []byte(`
	if a == b {
		var = 6
	}
	a = 1
	`)

	expect := []Token{tokenKeyword("if"), tokenIdentifier("a"), tokenOperator("=="), tokenIdentifier("b"),
		tokenCurlyOpen(), tokenIdentifier("var"), tokenAssignment(), tokenConstant("6"), tokenCurlyClose(),
		tokenIdentifier("a"), tokenAssignment(), tokenConstant("1"), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerIfElse(t *testing.T) {

	var code []byte = []byte(`
	if a == b {
		a = 6
	} else {
		a = 1
	}
	`)

	expect := []Token{tokenKeyword("if"), tokenIdentifier("a"), tokenOperator("=="), tokenIdentifier("b"),
		tokenCurlyOpen(), tokenIdentifier("a"), tokenAssignment(), tokenConstant("6"), tokenCurlyClose(),
		tokenKeyword("else"), tokenCurlyOpen(), tokenIdentifier("a"), tokenAssignment(),
		tokenConstant("1"), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerAssignment(t *testing.T) {

	var code []byte = []byte(`
	a = 1
	a, b = 1, 2
	a, b, c = 1, 2, 3
	`)

	expect := []Token{tokenIdentifier("a"), tokenAssignment(), tokenConstant("1"),
		tokenIdentifier("a"), tokenSeparator(), tokenIdentifier("b"), tokenAssignment(), tokenConstant("1"),
		tokenSeparator(), tokenConstant("2"), tokenIdentifier("a"), tokenSeparator(),
		tokenIdentifier("b"), tokenSeparator(), tokenIdentifier("c"), tokenAssignment(),
		tokenConstant("1"), tokenSeparator(), tokenConstant("2"), tokenSeparator(), tokenConstant("3"),
		tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerFor1(t *testing.T) {

	var code []byte = []byte(`
	for ;; {
		a = a+1
	}
	`)

	expect := []Token{tokenKeyword("for"), tokenSemicolon(), tokenSemicolon(), tokenCurlyOpen(),
		tokenIdentifier("a"), tokenAssignment(), tokenIdentifier("a"), tokenOperator("+"),
		tokenConstant("1"), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerFor2(t *testing.T) {

	var code []byte = []byte(`
	for i = 5;; {
		a = 0
	}
	`)

	expect := []Token{tokenKeyword("for"), tokenIdentifier("i"), tokenAssignment(), tokenConstant("5"),
		tokenSemicolon(), tokenSemicolon(), tokenCurlyOpen(), tokenIdentifier("a"),
		tokenAssignment(), tokenConstant("0"), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerFor3(t *testing.T) {

	var code []byte = []byte(`
	for i, j = 0, 1; i < 10; i = i+1 {
		if b == a {
			for ;; {
				c = 6
			}
		}
	}
	`)

	expect := []Token{tokenKeyword("for"), tokenIdentifier("i"), tokenSeparator(), tokenIdentifier("j"),
		tokenAssignment(), tokenConstant("0"), tokenSeparator(), tokenConstant("1"),
		tokenSemicolon(), tokenIdentifier("i"), tokenOperator("<"), tokenConstant("10"), tokenSemicolon(),
		tokenIdentifier("i"), tokenAssignment(), tokenIdentifier("i"), tokenOperator("+"), tokenConstant("1"),
		tokenCurlyOpen(), tokenKeyword("if"), tokenIdentifier("b"), tokenOperator("=="), tokenIdentifier("a"),
		tokenCurlyOpen(), tokenKeyword("for"), tokenSemicolon(), tokenSemicolon(),
		tokenCurlyOpen(), tokenIdentifier("c"), tokenAssignment(), tokenConstant("6"),
		tokenCurlyClose(), tokenCurlyClose(), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerFunction1(t *testing.T) {

	var code []byte = []byte(`
	fun abc() {
		a = 1
	}
	`)

	expect := []Token{tokenKeyword("fun"), tokenIdentifier("abc"), tokenParenOpen(), tokenParenClose(), tokenCurlyOpen(),
		tokenIdentifier("a"), tokenAssignment(), tokenConstant("1"), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerFunction2(t *testing.T) {

	var code []byte = []byte(`
	fun abc(a int, b float, c bool) {
		a = 1
		return a
	}
	`)

	expect := []Token{tokenKeyword("fun"), tokenIdentifier("abc"), tokenParenOpen(), tokenIdentifier("a"), tokenKeyword("int"),
		tokenSeparator(), tokenIdentifier("b"), tokenKeyword("float"), tokenSeparator(), tokenIdentifier("c"), tokenKeyword("bool"),
		tokenParenClose(), tokenCurlyOpen(), tokenIdentifier("a"), tokenAssignment(), tokenConstant("1"), tokenKeyword("return"),
		tokenIdentifier("a"), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}

func TestLexerFunction3(t *testing.T) {

	var code []byte = []byte(`
	fun abc() int, float, bool{
		a = 1
		return a, 3.5, true
	}
	`)

	expect := []Token{tokenKeyword("fun"), tokenIdentifier("abc"), tokenParenOpen(), tokenParenClose(), tokenKeyword("int"),
		tokenSeparator(), tokenKeyword("float"), tokenSeparator(), tokenKeyword("bool"), tokenCurlyOpen(), tokenIdentifier("a"),
		tokenAssignment(), tokenConstant("1"), tokenKeyword("return"), tokenIdentifier("a"), tokenSeparator(), tokenConstant("3.5"),
		tokenSeparator(), tokenConstant("true"), tokenCurlyClose(), tokenEOF(),
	}

	testTokens(code, expect, t)
}
