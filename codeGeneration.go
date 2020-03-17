package main

import (
	"fmt"
)

type ASM struct {
	constants []string
	variables []string
	program   []string

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

func (c Constant) generateCode(asm *ASM, s *SymbolTable) string {

	switch c.cType {
	case TYPE_INT:

		name := asm.nextConstName()
		registerName := "rsi"
		asm.constants = append(asm.constants, fmt.Sprintf("%-10v %6v %10v", name, "equ", c.cValue))
		asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, %v", registerName, name))
		return registerName

	case TYPE_FLOAT:

		name := asm.nextConstName()
		registerName := "xmm0"
		asm.constants = append(asm.constants, fmt.Sprintf("%-10v %6v %10v", name, "equ", c.cValue))
		asm.program = append(asm.program, fmt.Sprintf("  movq %10v, %v", registerName, name))
		return registerName

	case TYPE_STRING:

		name := asm.nextConstName()
		registerName := "rsi"
		asm.constants = append(asm.constants, fmt.Sprintf("%-10v %6v %10v, 0", name, "equ", fmt.Sprintf("\"%v\"", c.cValue)))
		asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, %15v", registerName, name))
		return registerName

	case TYPE_BOOL:

		registerName := "rsi"
		bValue := "FALSE"
		if c.cValue == "true" {
			bValue = "TRUE"
		}
		asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, %v", registerName, bValue))
		return registerName
	}

	fmt.Println("Could not generate code for Const. Unknown type!")
	return ""
}

func (v Variable) generateCode(asm *ASM, s *SymbolTable) string {

	if symbol, ok := s.get(v.vName); ok {

		switch symbol.sType {
		case TYPE_INT, TYPE_BOOL:
			registerName := "rsi"
			asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, qword [%v]", registerName, symbol.varName))
			return registerName
		case TYPE_FLOAT:
			registerName := "xmm0"
			asm.program = append(asm.program, fmt.Sprintf("  movq %10v, qword [%v]", registerName, symbol.varName))
			return registerName
		case TYPE_STRING:
			registerName := "rsi"
			asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, %15v", registerName, symbol.varName))
			return registerName
		}
	}

	fmt.Println("Could not generate code for Variable. No symbol known!")
	return ""
}

func (u UnaryOp) generateCode(asm *ASM, s *SymbolTable) string {

	register := u.expr.generateCode(asm, s)

	switch u.opType {
	case TYPE_BOOL:
		if u.operator == OP_NOT {
			// 'not' switches between 0 and -1. So False: 0, True: -1
			asm.program = append(asm.program, fmt.Sprintf("  not  %10v", register))
		} else {
			fmt.Printf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType)
		}
	case TYPE_INT:
		if u.operator == OP_NEGATIVE {
			asm.program = append(asm.program, fmt.Sprintf("  neg  %10v", register))
		} else {
			fmt.Printf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType)
		}
	case TYPE_FLOAT:
		if u.operator == OP_NEGATIVE {
			// TODO: Add negOneF and negOneI to global constants.
			asm.variables = append(asm.variables, fmt.Sprintf("negOneF %10v %15v", "dq", "-1.0"))
			asm.program = append(asm.program, fmt.Sprintf("  mulsd%10v, qword [negOneF]", register))

		} else {
			fmt.Printf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType)
		}
	case TYPE_STRING:
		fmt.Printf("Code generation error. No unary expression for Type String\n")
	}

	return register
}

// Reassigns a value from register r1 to register r2 to a given type
func reassignRegister(r1 string, t Type, asm *ASM) string {

	mov := "mov"
	target := "rcx"

	if t == TYPE_FLOAT {
		mov = "movq"
		target = "xmm1"
	}
	asm.program = append(asm.program, fmt.Sprintf("  %-4v %10v, %v", mov, target, r1))
	return target
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

// binaryOperationFloat executes the operation on the two registers and writes the result into rLeft!
func binaryOperationNumber(op Operator, t Type, rLeft, rRight string, asm *ASM) {

	command := ""

	switch op {
	case OP_PLUS:
		command = "add"
	case OP_MINUS:
		command = "sub"
	case OP_DIV:
		command = "div"
	case OP_MULT:
		// TODO: imul does not work with sd for float...
		// Make this into a function that just returns the corresponding string...
		command = "imul"
	case OP_GE, OP_GREATER, OP_LESS, OP_LE, OP_EQ, OP_NE:

		// ==:
		//   cmp rLeft, rRight
		//   je labelEQ
		//   mov rLeft, 0
		//   jmp labelOK
		// labelEQ:
		//   mov rleft, -1
		// labelOK:

		labelTrue := asm.nextLabelName()
		labelOK := asm.nextLabelName()
		jump := getJumpType(op)

		asm.program = append(asm.program, fmt.Sprintf("  cmp  %10v, %v", rLeft, rRight))
		asm.program = append(asm.program, fmt.Sprintf("  %-4v %10v", jump, labelTrue))
		asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, 0", rLeft))
		asm.program = append(asm.program, fmt.Sprintf("  jmp  %10v", labelOK))
		asm.program = append(asm.program, fmt.Sprintf("%v:", labelTrue))
		asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, -1", rLeft))
		asm.program = append(asm.program, fmt.Sprintf("%v:", labelOK))

		return

	default:
		fmt.Printf("Code generation error. Unknown operation %v for float.", op)
		return
	}

	if t == TYPE_FLOAT {
		// for double precision! Like: addsd, subsd, divsd, ...
		command += "sd"
	}

	asm.program = append(asm.program, fmt.Sprintf("  %-4v %10v, %v", command, rLeft, rRight))
}

// binaryOperationFloat executes the operation on the two registers and writes the result into rLeft!
func binaryOperationBool(op Operator, rLeft, rRight string, asm *ASM) {

	command := ""

	switch op {
	case OP_AND:
		command = "and"
	case OP_OR:
		command = "or"
	case OP_EQ:
		// Equal and unequal are identical for bool or int, as a bool is an integer type.
		binaryOperationNumber(op, TYPE_INT, rLeft, rRight, asm)
		return
	case OP_NE:
		binaryOperationNumber(op, TYPE_INT, rLeft, rRight, asm)
		return
	default:
		fmt.Printf("Code generation error. Unknown operation %v for bool.", op)
		return
	}

	asm.program = append(asm.program, fmt.Sprintf("  %-4v %10v, %v", command, rLeft, rRight))
}

func (b BinaryOp) generateCode(asm *ASM, s *SymbolTable) string {

	rLeft := b.leftExpr.generateCode(asm, s)

	// Save the result of the left register on the stack while evaluating the right hand side.
	asm.program = append(asm.program, fmt.Sprintf("  push %10v", rLeft))

	rRight := b.rightExpr.generateCode(asm, s)
	rRight = reassignRegister(rRight, b.opType, asm)

	// Pop the register back. So both rLeft and rRight are valid right now!
	asm.program = append(asm.program, fmt.Sprintf("  pop  %10v", rLeft))

	// do the binary operation thingy.
	switch b.opType {
	case TYPE_INT, TYPE_FLOAT:
		binaryOperationNumber(b.operator, b.opType, rLeft, rRight, asm)
	case TYPE_BOOL:
		binaryOperationBool(b.operator, rLeft, rRight, asm)
	case TYPE_STRING:
		fmt.Println("Code generation error: Strings not supported yet.")
	default:
		fmt.Printf("Code generation error: Unknown operation type %v\n", int(b.opType))
	}

	return rLeft
}

func debugPrint(asm *ASM, vName string) {

	asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, qword [%v]", "rsi", vName))
	asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, %v", "rdi", "fmti"))
	asm.program = append(asm.program, fmt.Sprintf("  mov  %10v, %v", "rax", "0"))
	asm.program = append(asm.program, fmt.Sprintf("  call %13v", "printf"))

}

func (a Assignment) generateCode(asm *ASM, s *SymbolTable) {

	for i, v := range a.variables {
		e := a.expressions[i]

		// Calculate expression
		register := e.generateCode(asm, s)

		// Create corresponding variable, if it doesn't exist yet.
		if entry, ok := s.get(v.vName); !ok || entry.varName == "" {

			vName := asm.nextVariableName()
			// Variables are initialized with 0. This will be overwritten a few lines later!
			asm.variables = append(asm.variables, fmt.Sprintf("%-10v    dq           0", vName))
			s.setAsmName(v.vName, vName)
		}
		// This can not/should not fail!
		entry, _ := s.get(v.vName)
		vName := entry.varName

		// Move value from register of expression into variable!
		asm.program = append(asm.program, fmt.Sprintf("  mov  %12v [%v], %v", "qword", vName, register))

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

	asm.constants = append(asm.constants, fmt.Sprintf("extern %10v  ; C function we need for debugging", "printf"))
	asm.constants = append(asm.constants, "section .data")

	asm.constants = append(asm.constants, "fmti          db \"%i\", 10, 0")

	asm.program = append(asm.program, "section .text")
	asm.program = append(asm.program, "global _start")
	asm.program = append(asm.program, "_start:")

	ast.block.generateCode(&asm, &ast.globalSymbolTable)

	asm.program = append(asm.program, "  ; Exit the program nicely")
	asm.program = append(asm.program, "  mov         rbx, 0  ; normal exit code")
	asm.program = append(asm.program, "  mov         rax, 1  ; process termination service (?)")
	asm.program = append(asm.program, "  int         0x80    ; linux kernel service")

	return asm
}
