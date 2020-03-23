package main

import (
	"testing"
)

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
	go tokenize(code, tokenChan)

	if ok, i, te, tg := testChannelEqualSlice(tokenChan, expect); !ok {
		t.Errorf("Expected %v, got %v at position %v\n", te, tg, i)
	}

	if len(tokenChan) != 0 {
		t.Errorf("%v tokens expected, got %v\n", len(expect), len(expect)+len(tokenChan))
	}
}

func TestLexerExpression1(t *testing.T) {

	var code []byte = []byte(`6 + 7 * variable / -(5 -- (-8 * - 10000.1234))`)

	expect := []Token{Token{TOKEN_CONSTANT, "6", 0, 0}, Token{TOKEN_OPERATOR, "+", 0, 0}, Token{TOKEN_CONSTANT, "7", 0, 0}, Token{TOKEN_OPERATOR, "*", 0, 0},
		Token{TOKEN_IDENTIFIER, "variable", 0, 0}, Token{TOKEN_OPERATOR, "/", 0, 0}, Token{TOKEN_OPERATOR, "-", 0, 0}, Token{TOKEN_PARENTHESIS_OPEN, "(", 0, 0}, Token{TOKEN_CONSTANT, "5", 0, 0},
		Token{TOKEN_OPERATOR, "-", 0, 0}, Token{TOKEN_OPERATOR, "-", 0, 0}, Token{TOKEN_PARENTHESIS_OPEN, "(", 0, 0}, Token{TOKEN_CONSTANT, "-8", 0, 0}, Token{TOKEN_OPERATOR, "*", 0, 0},
		Token{TOKEN_OPERATOR, "-", 0, 0}, Token{TOKEN_CONSTANT, "10000.1234", 0, 0}, Token{TOKEN_PARENTHESIS_CLOSE, ")", 0, 0}, Token{TOKEN_PARENTHESIS_CLOSE, ")", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
	}

	testTokens(code, expect, t)

}

func TestLexerExpression2(t *testing.T) {

	var code []byte = []byte(`a && b || (5 < false <= 8 && (false2 > variable >= 5.0) != true)`)

	expect := []Token{Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_OPERATOR, "&&", 0, 0}, Token{TOKEN_IDENTIFIER, "b", 0, 0}, Token{TOKEN_OPERATOR, "||", 0, 0},
		Token{TOKEN_PARENTHESIS_OPEN, "(", 0, 0}, Token{TOKEN_CONSTANT, "5", 0, 0}, Token{TOKEN_OPERATOR, "<", 0, 0}, Token{TOKEN_CONSTANT, "false", 0, 0}, Token{TOKEN_OPERATOR, "<=", 0, 0},
		Token{TOKEN_CONSTANT, "8", 0, 0}, Token{TOKEN_OPERATOR, "&&", 0, 0}, Token{TOKEN_PARENTHESIS_OPEN, "(", 0, 0}, Token{TOKEN_IDENTIFIER, "false2", 0, 0}, Token{TOKEN_OPERATOR, ">", 0, 0},
		Token{TOKEN_IDENTIFIER, "variable", 0, 0}, Token{TOKEN_OPERATOR, ">=", 0, 0}, Token{TOKEN_CONSTANT, "5.0", 0, 0}, Token{TOKEN_PARENTHESIS_CLOSE, ")", 0, 0}, Token{TOKEN_OPERATOR, "!=", 0, 0},
		Token{TOKEN_CONSTANT, "true", 0, 0}, Token{TOKEN_PARENTHESIS_CLOSE, ")", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
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

	expect := []Token{Token{TOKEN_KEYWORD, "if", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_OPERATOR, "==", 0, 0}, Token{TOKEN_IDENTIFIER, "b", 0, 0},
		Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_IDENTIFIER, "var", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "6", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0},
		Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "1", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
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

	expect := []Token{Token{TOKEN_KEYWORD, "if", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_OPERATOR, "==", 0, 0}, Token{TOKEN_IDENTIFIER, "b", 0, 0},
		Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "6", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0},
		Token{TOKEN_KEYWORD, "else", 0, 0}, Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0},
		Token{TOKEN_CONSTANT, "1", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
	}

	testTokens(code, expect, t)
}

func TestLexerAssignment(t *testing.T) {

	var code []byte = []byte(`
	a = 1
	a, b = 1, 2
	a, b, c = 1, 2, 3
	`)

	expect := []Token{Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "1", 0, 0},
		Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_IDENTIFIER, "b", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "1", 0, 0},
		Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_CONSTANT, "2", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0},
		Token{TOKEN_IDENTIFIER, "b", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_IDENTIFIER, "c", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0},
		Token{TOKEN_CONSTANT, "1", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_CONSTANT, "2", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_CONSTANT, "3", 0, 0},
		Token{TOKEN_EOF, "", 0, 0},
	}

	testTokens(code, expect, t)
}

func TestLexerFor1(t *testing.T) {

	var code []byte = []byte(`
	for ;; {
		a = a+1
	}
	`)

	expect := []Token{Token{TOKEN_KEYWORD, "for", 0, 0}, Token{TOKEN_SEMICOLON, ";", 0, 0}, Token{TOKEN_SEMICOLON, ";", 0, 0}, Token{TOKEN_CURLY_OPEN, "{", 0, 0},
		Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0}, Token{TOKEN_OPERATOR, "+", 0, 0},
		Token{TOKEN_CONSTANT, "1", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
	}

	testTokens(code, expect, t)
}

func TestLexerFor2(t *testing.T) {

	var code []byte = []byte(`
	for i = 5;; {
		a = 0
	}
	`)

	expect := []Token{Token{TOKEN_KEYWORD, "for", 0, 0}, Token{TOKEN_IDENTIFIER, "i", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "5", 0, 0},
		Token{TOKEN_SEMICOLON, ";", 0, 0}, Token{TOKEN_SEMICOLON, ";", 0, 0}, Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0},
		Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "0", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
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

	expect := []Token{Token{TOKEN_KEYWORD, "for", 0, 0}, Token{TOKEN_IDENTIFIER, "i", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_IDENTIFIER, "j", 0, 0},
		Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "0", 0, 0}, Token{TOKEN_SEPARATOR, ",", 0, 0}, Token{TOKEN_CONSTANT, "1", 0, 0},
		Token{TOKEN_SEMICOLON, ";", 0, 0}, Token{TOKEN_IDENTIFIER, "i", 0, 0}, Token{TOKEN_OPERATOR, "<", 0, 0}, Token{TOKEN_CONSTANT, "10", 0, 0}, Token{TOKEN_SEMICOLON, ";", 0, 0},
		Token{TOKEN_IDENTIFIER, "i", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_IDENTIFIER, "i", 0, 0}, Token{TOKEN_OPERATOR, "+", 0, 0}, Token{TOKEN_CONSTANT, "1", 0, 0},
		Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_KEYWORD, "if", 0, 0}, Token{TOKEN_IDENTIFIER, "b", 0, 0}, Token{TOKEN_OPERATOR, "==", 0, 0}, Token{TOKEN_IDENTIFIER, "a", 0, 0},
		Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_KEYWORD, "for", 0, 0}, Token{TOKEN_SEMICOLON, ";", 0, 0}, Token{TOKEN_SEMICOLON, ";", 0, 0},
		Token{TOKEN_CURLY_OPEN, "{", 0, 0}, Token{TOKEN_IDENTIFIER, "c", 0, 0}, Token{TOKEN_ASSIGNMENT, "=", 0, 0}, Token{TOKEN_CONSTANT, "6", 0, 0},
		Token{TOKEN_CURLY_CLOSE, "}", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0}, Token{TOKEN_CURLY_CLOSE, "}", 0, 0}, Token{TOKEN_EOF, "", 0, 0},
	}

	testTokens(code, expect, t)
}
