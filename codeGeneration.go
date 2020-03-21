package main

import (
	"fmt"
)

type ASM struct {
	header    []string
	constants [][2]string
	variables [][3]string
	program   [][3]string

	// Increasing number to generate unique const variable names
	constName int
	varName   int
	labelName int
}

func (asm *ASM) nextConstName() string {
	asm.constName += 1
	return fmt.Sprintf("const_%v", asm.constName-1)
}

func (asm *ASM) nextVariableName() string {
	asm.varName += 1
	return fmt.Sprintf("var_%v", asm.varName-1)
}

func (asm *ASM) nextLabelName() string {
	asm.labelName += 1
	return fmt.Sprintf("label_%v", asm.labelName-1)
}

func getJumpType(op Operator) string {
	switch op {
	case OP_GE:
		return "jge"
	case OP_GREATER:
		return "jg"
	case OP_LESS:
		return "jl"
	case OP_LE:
		return "jle"
	case OP_EQ:
		return "je"
	case OP_NE:
		return "jne"
	}
	return ""
}

func getCommandFloat(op Operator) string {
	switch op {
	case OP_PLUS:
		return "addsd"
	case OP_MINUS:
		return "subsd"
	case OP_MULT:
		return "mulsd"
	case OP_DIV:
		return "divsd"
	default:
		fmt.Println("Code generation error. Unknown operator for Float")
	}
	return ""
}

func getCommandInt(op Operator) string {
	switch op {
	case OP_PLUS:
		return "add"
	case OP_MINUS:
		return "sub"
	case OP_MULT:
		return "imul"
	case OP_DIV:
		return "div"
	default:
		fmt.Println("Code generation error. Unknown operator for Integer")
	}
	return ""
}
func getCommandBool(op Operator) string {
	switch op {
	case OP_AND:
		return "and"
	case OP_OR:
		return "or"
	default:
		fmt.Println("Code generation error. Unknown operator for bool")
	}
	return ""
}

func getCommand(t Type, op Operator) string {

	switch t {
	case TYPE_BOOL:
		return getCommandBool(op)
	case TYPE_FLOAT:
		return getCommandFloat(op)
	case TYPE_INT:
		return getCommandInt(op)
	case TYPE_STRING:
		fmt.Println("Code generation error. String commands not yet implemented")
	}
	return ""
}

func (c Constant) generateCode(asm *ASM, s *SymbolTable) {

	name := ""
	switch c.cType {
	case TYPE_INT:
		name = asm.nextConstName()
		asm.constants = append(asm.constants, [2]string{name, c.cValue})
	case TYPE_FLOAT:
		name = asm.nextConstName()
		asm.constants = append(asm.constants, [2]string{name, c.cValue})
	case TYPE_STRING:
		name = asm.nextConstName()
		asm.constants = append(asm.constants, [2]string{name, fmt.Sprintf("\"%v\", 0", c.cValue)})
	case TYPE_BOOL:
		name = "FALSE"
		if c.cValue == "true" {
			name = "TRUE"
		}
	default:
		fmt.Println("Could not generate code for Const. Unknown type!")
		return
	}

	asm.program = append(asm.program, [3]string{"  ", "push", name})
}

func (v Variable) generateCode(asm *ASM, s *SymbolTable) {

	if symbol, ok := s.get(v.vName); ok {
		asm.program = append(asm.program, [3]string{"  ", "push", symbol.varName})
	}
	fmt.Println("Could not generate code for Variable. No symbol known!")
}

func (u UnaryOp) generateCode(asm *ASM, s *SymbolTable) {

	u.expr.generateCode(asm, s)

	register := ""

	switch u.getExpressionType() {
	case TYPE_BOOL:
		if u.operator == OP_NOT {
			// 'not' switches between 0 and -1. So False: 0, True: -1
			register = "rsi"
			asm.program = append(asm.program, [3]string{"  ", "pop", register})
			asm.program = append(asm.program, [3]string{"  ", "not", register})
		} else {
			fmt.Printf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType)
		}
	case TYPE_INT:
		if u.operator == OP_NEGATIVE {
			register = "rsi"
			asm.program = append(asm.program, [3]string{"  ", "pop", register})
			asm.program = append(asm.program, [3]string{"  ", "neg", register})

		} else {
			fmt.Printf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType)
		}
	case TYPE_FLOAT:
		if u.operator == OP_NEGATIVE {
			register = "xmm0"
			asm.program = append(asm.program, [3]string{"  ", "pop", register})
			asm.program = append(asm.program, [3]string{"  ", "mulsd", fmt.Sprintf("%v, qword [negOneF]", register)})

		} else {
			fmt.Printf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType)
		}
	case TYPE_STRING:
		fmt.Printf("Code generation error. No unary expression for Type String\n")
		return
	}

	asm.program = append(asm.program, [3]string{"  ", "push", register})
}

// binaryOperationFloat executes the operation on the two registers and writes the result into rLeft!
func binaryOperationNumber(op Operator, t Type, rLeft, rRight string, asm *ASM) {

	switch op {
	case OP_GE, OP_GREATER, OP_LESS, OP_LE, OP_EQ, OP_NE:

		// Works for anything that should be compared.
		labelTrue := asm.nextLabelName()
		labelOK := asm.nextLabelName()
		jump := getJumpType(op)

		asm.program = append(asm.program, [3]string{"  ", "cmp", fmt.Sprintf("%v, %v", rLeft, rRight)})
		asm.program = append(asm.program, [3]string{"  ", jump, labelTrue})
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("%v, 0", rLeft)})
		asm.program = append(asm.program, [3]string{"  ", "jmp", labelOK})
		asm.program = append(asm.program, [3]string{"", labelTrue + ":", ""})
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("%v, -1", rLeft)})
		asm.program = append(asm.program, [3]string{"", labelOK + ":", ""})

	default:
		// Works for Integer and Float.
		command := getCommand(t, op)
		asm.program = append(asm.program, [3]string{"  ", command, fmt.Sprintf("%v, %v", rLeft, rRight)})
	}
}

func (b BinaryOp) generateCode(asm *ASM, s *SymbolTable) {

	b.leftExpr.generateCode(asm, s)
	b.rightExpr.generateCode(asm, s)

	rLeft := ""
	rRight := ""
	switch b.leftExpr.getExpressionType() {
	case TYPE_INT, TYPE_BOOL, TYPE_STRING:
		rLeft = "rsi"
		rRight = "rcx"
	case TYPE_FLOAT:
		rLeft = "xmm0"
		rRight = "xmm1"
	}

	asm.program = append(asm.program, [3]string{"  ", "pop", rRight})
	asm.program = append(asm.program, [3]string{"  ", "pop", rLeft})

	switch b.leftExpr.getExpressionType() {
	case TYPE_INT, TYPE_FLOAT:
		binaryOperationNumber(b.operator, b.opType, rLeft, rRight, asm)
	case TYPE_BOOL:
		// Equal and unequal are identical for bool or int, as a bool is an integer type.
		if b.operator == OP_EQ || b.operator == OP_NE {
			binaryOperationNumber(b.operator, b.opType, rLeft, rRight, asm)
		} else {
			command := getCommand(TYPE_BOOL, b.operator)
			asm.program = append(asm.program, [3]string{"  ", command, fmt.Sprintf("%v, %v", rLeft, rRight)})
		}

	case TYPE_STRING:
		fmt.Println("Code generation error: Strings not supported yet.")
	default:
		fmt.Printf("Code generation error: Unknown operation type %v\n", int(b.opType))
	}

	asm.program = append(asm.program, [3]string{"  ", "push", rLeft})
}

func debugPrint(asm *ASM, vName string) {
	asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("rsi, qword [%v]", vName)})
	asm.program = append(asm.program, [3]string{"  ", "mov", "rdi, fmti"})
	asm.program = append(asm.program, [3]string{"  ", "mov", "rax, 0"})
	asm.program = append(asm.program, [3]string{"  ", "call", "printf"})
}

func (a Assignment) generateCode(asm *ASM, s *SymbolTable) {

	for i, v := range a.variables {
		e := a.expressions[i]

		// Calculate expression
		e.generateCode(asm, s)

		register := ""
		switch e.getExpressionType() {
		case TYPE_INT, TYPE_BOOL, TYPE_STRING:
			register = "rsi"
		case TYPE_FLOAT:
			register = "xmm0"
		}

		asm.program = append(asm.program, [3]string{"  ", "pop", register})

		// Create corresponding variable, if it doesn't exist yet.
		if entry, ok := s.get(v.vName); !ok || entry.varName == "" {

			vName := asm.nextVariableName()
			// Variables are initialized with 0. This will be overwritten a few lines later!
			// TODO: What about float or string variables? Does this really work? Do we care at all because it will be overwritten anyway?
			asm.variables = append(asm.variables, [3]string{vName, "dq", "0"})
			s.setAsmName(v.vName, vName)
		}
		// This can not/should not fail!
		entry, _ := s.get(v.vName)
		vName := entry.varName

		// Move value from register of expression into variable!
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("qword [%v], %v", vName, register)})

		debugPrint(asm, vName)

	}

}

func (c Condition) generateCode(asm *ASM, s *SymbolTable) {
}

func (l Loop) generateCode(asm *ASM, s *SymbolTable) {
}

func (b Block) generateCode(asm *ASM, s *SymbolTable) {

	for _, statement := range b.statements {
		statement.generateCode(asm, &b.symbolTable)
	}

}

func (ast AST) generateCode() ASM {

	asm := ASM{}

	asm.header = append(asm.header, "extern printf  ; C function we need for debugging")
	asm.header = append(asm.header, "section .data")

	asm.constants = append(asm.constants, [2]string{"TRUE", "-1"})
	asm.constants = append(asm.constants, [2]string{"FALSE", "0"})

	asm.variables = append(asm.variables, [3]string{"fmti", "db", "\"%i\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"negOneF", "dq", "-1.0"})
	asm.variables = append(asm.variables, [3]string{"negOneI", "dq", "-1"})

	asm.program = append(asm.program, [3]string{"", "section .text", ""})
	asm.program = append(asm.program, [3]string{"", "global _start", ""})
	asm.program = append(asm.program, [3]string{"", "_start:", ""})

	ast.block.generateCode(&asm, &ast.globalSymbolTable)

	asm.program = append(asm.program, [3]string{"  ", "; Exit the program nicely", ""})
	asm.program = append(asm.program, [3]string{"  ", "mov", "rbx, 0  ; normal exit code"})
	asm.program = append(asm.program, [3]string{"  ", "mov", "rax, 1  ; process termination service (?)"})
	asm.program = append(asm.program, [3]string{"  ", "int", "0x80    ; linux kernel service"})

	return asm
}
