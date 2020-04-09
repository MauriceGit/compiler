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
	TOKEN_SQUARE_OPEN
	TOKEN_SQUARE_CLOSE
	TOKEN_SEMICOLON
	TOKEN_EOF
	TOKEN_UNKNOWN
)

type TokenType int

type Token struct {
	tokenType TokenType
	value     string
	line      int
	column    int
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
	case TOKEN_SQUARE_OPEN:
		return "TOKEN_SQUARE_OPEN"
	case TOKEN_SQUARE_CLOSE:
		return "TOKEN_SQARE_CLOSE"
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

func parseByte(program []byte) (TokenType, bool) {

	switch program[0] {
	case ',':
		return TOKEN_SEPARATOR, true
	case ';':
		return TOKEN_SEMICOLON, true
	case '(':
		return TOKEN_PARENTHESIS_OPEN, true
	case ')':
		return TOKEN_PARENTHESIS_CLOSE, true
	case '{':
		return TOKEN_CURLY_OPEN, true
	case '}':
		return TOKEN_CURLY_CLOSE, true
	case '[':
		return TOKEN_SQUARE_OPEN, true
	case ']':
		return TOKEN_SQUARE_CLOSE, true
	}
	return TOKEN_UNKNOWN, false
}

func tokenize(program []byte, tokens chan Token, err chan error) {
	// Whitespace is just: \s without the \n, so we can track the line count explicitely.
	whitespace := regexp.MustCompile(`^[\t\f\r ]`)
	newline := regexp.MustCompile(`^\n`)
	comment := regexp.MustCompile(`^//.*\n`)
	keyword := regexp.MustCompile(`^(int|string|float|bool|if|else|for|shadow|fun|return)\W`)
	operator := regexp.MustCompile(`^(\+|\-|\*|/|%|==|!=|<=|>=|<|>|\|\||&&|!)`)
	assignment := regexp.MustCompile(`^=`)
	constant := regexp.MustCompile(`^(((\d+(\.\d+)?)|(".*"))|(true|false))`)
	identifier := regexp.MustCompile(`^[A-Za-z]\w*`)

	lineCnt := 0
	colCnt := 0

	for len(program) > 0 {

		// Whitespace has high priority, so we directly continue!
		if s := whitespace.FindIndex(program); s != nil {
			program = program[s[1]:]
			colCnt += s[1]
			continue
		}
		if s := newline.FindIndex(program); s != nil {
			program = program[s[1]:]
			lineCnt++
			colCnt = 0
			continue
		}
		// Comments also have high priority to be ignored :)
		if s := comment.FindIndex(program); s != nil {
			program = program[s[1]:]
			lineCnt++
			colCnt = 0
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
			err <- fmt.Errorf("[%v:%v] - Unknown string", lineCnt, colCnt)
			tokens <- Token{TOKEN_EOF, "", lineCnt, colCnt}
			return
		}

		tokens <- Token{tokenType, string(program[:tokenLength]), lineCnt, colCnt}
		program = program[tokenLength:]
		colCnt += tokenLength
	}

	tokens <- Token{TOKEN_EOF, "", lineCnt, colCnt}
}
