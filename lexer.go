// lexer.go
package main

import (
	"fmt"
	"regexp"
)

const (
	TOKEN_KEYWORD = iota
	TOKEN_IDENTIFIER
	TOKEN_OPERATOR
	TOKEN_ASSIGNMENT
	TOKEN_CONSTANT
	TOKEN_SEPARATOR
	TOKEN_PARENTHESIS_OPEN
	TOKEN_PARENTHESIS_CLOSE
	TOKEN_CURLY_OPEN
	TOKEN_CURLY_CLOSE
	TOKEN_SEMICOLON
	TOKEN_EOF
	TOKEN_UNKNOWN
)

type TokenType int

type Token struct {
	tokenType TokenType
	value     string
}

func (t TokenType) String() string {
	switch t {
	case TOKEN_KEYWORD:
		return "TOKEN_KEYWORD"
	case TOKEN_IDENTIFIER:
		return "TOKEN_IDENTIFIER"
	case TOKEN_OPERATOR:
		return "TOKEN_OPERATOR"
	case TOKEN_ASSIGNMENT:
		return "TOKEN_ASSIGNMENT"
	case TOKEN_CONSTANT:
		return "TOKEN_CONSTANT"
	case TOKEN_SEPARATOR:
		return "TOKEN_SEPARATOR"
	case TOKEN_PARENTHESIS_OPEN:
		return "TOKEN_PARENTHESIS_OPEN"
	case TOKEN_PARENTHESIS_CLOSE:
		return "TOKEN_PARENTHESIS_CLOSE"
	case TOKEN_CURLY_OPEN:
		return "TOKEN_CURLY_OPEN"
	case TOKEN_CURLY_CLOSE:
		return "TOKEN_CURLY_CLOSE"
	case TOKEN_SEMICOLON:
		return "TOKEN_SEMICOLON"
	case TOKEN_EOF:
		return "TOKEN_EOF"
	}
	return "Unknown Token"
}

func (t Token) String() string {
	return fmt.Sprintf("(%v %v)", t.value, t.tokenType)
}

func byteToToken(b byte) TokenType {

	switch b {
	case ',':
		return TOKEN_SEPARATOR
	case ';':
		return TOKEN_SEMICOLON
	case '(':
		return TOKEN_PARENTHESIS_OPEN
	case ')':
		return TOKEN_PARENTHESIS_CLOSE
	case '{':
		return TOKEN_CURLY_OPEN
	case '}':
		return TOKEN_CURLY_CLOSE
	}
	return TOKEN_UNKNOWN
}

func parseByte(program []byte) (TokenType, bool) {

	tokens := []byte{',', ';', '(', ')', '{', '}'}
	for _, t := range tokens {
		if program[0] == t {
			return byteToToken(t), true
		}
	}
	return TOKEN_UNKNOWN, false

}

func tokenize(program []byte, tokens chan Token) {
	space := regexp.MustCompile(`^((\s+)|\n)`)
	keyword := regexp.MustCompile(`^(int|string|float|if|else|for) `)
	operator := regexp.MustCompile(`^(\+|\-|\*|/|==|!=|<=|>=|<|>|\|\||&&)`)
	assignment := regexp.MustCompile(`^=`)
	constant := regexp.MustCompile(`^(((-?\d+(\.\d+)?)|(".*"))|(true|false))`)
	identifier := regexp.MustCompile(`^[A-Za-z]\w*`)

	for len(program) > 0 {

		// Whitespace has high priority, so we directly continue!
		if s := space.FindIndex(program); s != nil {
			program = program[s[1]:]
			continue
		}

		var tokenType TokenType
		tokenLength := 0

		if tt, ok := parseByte(program); ok {
			tokenLength = 1
			tokenType = tt
		}

		if s := keyword.FindIndex(program); s != nil && s[1] > tokenLength {
			tokenLength = s[1] - 1
			tokenType = TOKEN_KEYWORD
		}
		if s := operator.FindIndex(program); s != nil && s[1] > tokenLength {
			tokenLength = s[1]
			tokenType = TOKEN_OPERATOR
		}
		if s := assignment.FindIndex(program); s != nil && s[1] > tokenLength {
			tokenLength = s[1]
			tokenType = TOKEN_ASSIGNMENT
		}
		if s := constant.FindIndex(program); s != nil && s[1] > tokenLength {
			tokenLength = s[1]
			tokenType = TOKEN_CONSTANT
		}
		// Lowest priority for parsing!
		if s := identifier.FindIndex(program); s != nil && s[1] > tokenLength {
			tokenLength = s[1]
			tokenType = TOKEN_IDENTIFIER
		}

		if tokenLength == 0 {
			fmt.Printf("Unknown string: %v\n", string(program))
			break
		}

		tokens <- Token{tokenType, string(program[:tokenLength])}
		program = program[tokenLength:]
	}

	tokens <- Token{TOKEN_EOF, ""}

}
