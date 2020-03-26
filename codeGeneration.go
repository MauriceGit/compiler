package main

import (
	"fmt"
)

type ASM struct {
	header      []string
	constants   [][2]string
	variables   [][3]string
	sectionText []string
	functions   [][3]string
	program     [][3]string

	// Increasing number to generate unique const variable names
	constName int
	varName   int
	labelName int
	funName   int
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

func (asm *ASM) nextFunctionName() string {
	asm.funName += 1
	return fmt.Sprintf("fun_%v", asm.funName-1)
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
		panic("Code generation error. Unknown operator for Float")
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
		panic("Code generation error. Unknown operator for Integer")
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
		panic("Code generation error. Unknown operator for bool")
	}
	return ""
}

func getRegister(t Type) (string, string) {
	switch t {
	case TYPE_INT, TYPE_BOOL:
		return "r10", "r11"
	case TYPE_FLOAT:
		return "xmm0", "xmm1"
	case TYPE_STRING:
		panic("Code generation error. String register not yet implemented")
	}
	return "", ""
}

// Right now we limit ourself to a maximum of 6 integer parameters and/or 8 floating parameters!
// https://wiki.cdot.senecacollege.ca/wiki/X86_64_Register_and_Instruction_Quick_Start
func getFunctionRegisters(t Type) []string {
	switch t {
	case TYPE_INT, TYPE_BOOL:
		return []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
	case TYPE_FLOAT:
		return []string{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7"}
	case TYPE_STRING:
		panic("Code generation error. String register not yet implemented")
	}
	return []string{}
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
		panic("Code generation error. String commands not yet implemented")
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
		panic("Could not generate code for Const. Unknown type!")
	}

	register, _ := getRegister(TYPE_INT)
	asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("%v, %v", register, name)})
	asm.program = append(asm.program, [3]string{"  ", "push", register})
}

func (v Variable) generateCode(asm *ASM, s *SymbolTable) {

	if symbol, ok := s.getVar(v.vName); ok {
		// We just push it on the stack. So here we can't use floating point registers for example!
		register, _ := getRegister(TYPE_INT)
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("%v, qword [%v]", register, symbol.varName)})
		asm.program = append(asm.program, [3]string{"  ", "push", register})
		return
	}
	panic("Could not generate code for Variable. No symbol known!")
}

func (u UnaryOp) generateCode(asm *ASM, s *SymbolTable) {

	u.expr.generateCode(asm, s)

	stackReg, _ := getRegister(TYPE_INT)
	asm.program = append(asm.program, [3]string{"  ", "pop", stackReg})

	//register, _ := getRegister(u.getExpressionType())

	switch u.getExpressionType() {
	case TYPE_BOOL:
		if u.operator == OP_NOT {
			// 'not' switches between 0 and -1. So False: 0, True: -1
			//asm.program = append(asm.program, [3]string{"  ", "pop", register})
			asm.program = append(asm.program, [3]string{"  ", "not", stackReg})
		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType))
		}
	case TYPE_INT:
		if u.operator == OP_NEGATIVE {
			//asm.program = append(asm.program, [3]string{"  ", "pop", register})
			asm.program = append(asm.program, [3]string{"  ", "neg", stackReg})

		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType))
		}
	case TYPE_FLOAT:
		if u.operator == OP_NEGATIVE {
			//asm.program = append(asm.program, [3]string{"  ", "pop", register})

			fReg, _ := getRegister(TYPE_FLOAT)
			asm.program = append(asm.program, [3]string{"  ", "movq", fmt.Sprintf("%v, %v", fReg, stackReg)})
			asm.program = append(asm.program, [3]string{"  ", "mulsd", fmt.Sprintf("%v, qword [negOneF]", fReg)})
			asm.program = append(asm.program, [3]string{"  ", "movq", fmt.Sprintf("%v, %v", stackReg, fReg)})

		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v", u.operator, u.opType))
		}
	case TYPE_STRING:
		panic("Code generation error. No unary expression for Type String")
		return
	}

	asm.program = append(asm.program, [3]string{"  ", "push", stackReg})
}

// binaryOperationFloat executes the operation on the two registers and writes the result into rLeft!
func binaryOperationNumber(op Operator, t Type, rLeft, rRight string, asm *ASM) {

	switch op {
	case OP_GE, OP_GREATER, OP_LESS, OP_LE, OP_EQ, OP_NE:

		// Works for anything that should be compared.
		labelTrue := asm.nextLabelName()
		labelOK := asm.nextLabelName()
		jump := getJumpType(op)

		// TODO: Verify that this works with float as well!
		asm.program = append(asm.program, [3]string{"  ", "cmp", fmt.Sprintf("%v, %v", rLeft, rRight)})
		asm.program = append(asm.program, [3]string{"  ", jump, labelTrue})
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("%v, 0", rLeft)})
		asm.program = append(asm.program, [3]string{"  ", "jmp", labelOK})
		asm.program = append(asm.program, [3]string{"", labelTrue + ":", ""})
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("%v, -1", rLeft)})
		asm.program = append(asm.program, [3]string{"", labelOK + ":", ""})

	default:
		command := getCommand(t, op)
		switch t {
		case TYPE_INT:
			asm.program = append(asm.program, [3]string{"  ", command, fmt.Sprintf("%v, %v", rLeft, rRight)})
		case TYPE_FLOAT:
			fLeft, fRight := getRegister(TYPE_FLOAT)
			// We need to move the values from stack/int registers to the corresponding xmm* ones and back afterwards
			// So the stack push works!
			asm.program = append(asm.program, [3]string{"  ", "movq", fmt.Sprintf("%v, %v", fLeft, rLeft)})
			asm.program = append(asm.program, [3]string{"  ", "movq", fmt.Sprintf("%v, %v", fRight, rRight)})
			asm.program = append(asm.program, [3]string{"  ", command, fmt.Sprintf("%v, %v", fLeft, fRight)})
			asm.program = append(asm.program, [3]string{"  ", "movq", fmt.Sprintf("%v, %v", rLeft, fLeft)})
			asm.program = append(asm.program, [3]string{"  ", "movq", fmt.Sprintf("%v, %v", rRight, fRight)})
		}
	}
}

func (b BinaryOp) generateCode(asm *ASM, s *SymbolTable) {

	b.leftExpr.generateCode(asm, s)
	b.rightExpr.generateCode(asm, s)

	rLeft, rRight := getRegister(TYPE_INT)

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
		panic("Code generation error: Strings not supported yet.")
	default:
		panic(fmt.Sprintf("Code generation error: Unknown operation type %v\n", int(b.opType)))
	}

	asm.program = append(asm.program, [3]string{"  ", "push", rLeft})
}

func debugPrint(asm *ASM, vName string, t Type) {

	format := "fmti"
	floatCount := "0"
	if t == TYPE_FLOAT {
		format = "fmtf"
		floatCount = "1"
		fReg, _ := getRegister(t)
		asm.program = append(asm.program, [3]string{"    ", "movq", fmt.Sprintf("%v, qword [%v]", fReg, vName)})
		asm.program = append(asm.program, [3]string{"    ", "movq", "rsi, " + fReg})
	} else {
		asm.program = append(asm.program, [3]string{"    ", "mov", fmt.Sprintf("rsi, qword [%v]", vName)})
	}
	asm.program = append(asm.program, [3]string{"    ", "mov", "rdi, " + format})
	asm.program = append(asm.program, [3]string{"    ", "mov", "rax, " + floatCount})
	asm.program = append(asm.program, [3]string{"    ", "call", "printf"})
}

func (a Assignment) generateCode(asm *ASM, s *SymbolTable) {

	for i, v := range a.variables {
		e := a.expressions[i]

		// Calculate expression
		e.generateCode(asm, s)

		// Just to get it from the stack into the variable.
		register, _ := getRegister(TYPE_INT)
		asm.program = append(asm.program, [3]string{"  ", "pop", register})

		// Create corresponding variable, if it doesn't exist yet.
		if entry, ok := s.getVar(v.vName); !ok || entry.varName == "" {

			vName := asm.nextVariableName()
			// Variables are initialized with 0. This will be overwritten a few lines later!
			// TODO: What about float or string variables? Does this really work? Do we care at all because it will be overwritten anyway?
			asm.variables = append(asm.variables, [3]string{vName, "dq", "0"})
			s.setAsmName(v.vName, vName)
		}
		// This can not/should not fail!
		entry, _ := s.getVar(v.vName)
		vName := entry.varName

		// Move value from register of expression into variable!
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("qword [%v], %v", vName, register)})

		debugPrint(asm, vName, v.vType)
	}

}

func (c Condition) generateCode(asm *ASM, s *SymbolTable) {

	c.expression.generateCode(asm, s)

	register, _ := getRegister(TYPE_BOOL)
	// For now, we assume an else case. Even if it is just empty!
	elseLabel := asm.nextLabelName()
	endLabel := asm.nextLabelName()

	asm.program = append(asm.program, [3]string{"  ", "pop", register})
	asm.program = append(asm.program, [3]string{"  ", "cmp", fmt.Sprintf("%v, 0", register)})
	asm.program = append(asm.program, [3]string{"  ", "je", elseLabel})

	c.block.generateCode(asm, s)

	asm.program = append(asm.program, [3]string{"  ", "jmp", endLabel})
	asm.program = append(asm.program, [3]string{"", elseLabel + ":", ""})

	c.elseBlock.generateCode(asm, s)

	asm.program = append(asm.program, [3]string{"", endLabel + ":", ""})
}

func (l Loop) generateCode(asm *ASM, s *SymbolTable) {

	register, _ := getRegister(TYPE_BOOL)
	startLabel := asm.nextLabelName()
	evalLabel := asm.nextLabelName()
	endLabel := asm.nextLabelName()

	// The initial assignment is logically moved inside the for-block
	l.assignment.generateCode(asm, &l.block.symbolTable)

	asm.program = append(asm.program, [3]string{"  ", "jmp", evalLabel})
	asm.program = append(asm.program, [3]string{"", startLabel + ":", ""})

	l.block.generateCode(asm, s)

	// The increment assignment is logically moved inside the for-block
	l.incrAssignment.generateCode(asm, &l.block.symbolTable)

	asm.program = append(asm.program, [3]string{"", evalLabel + ":", ""})

	// If any of the expressions result in False (0), we jump to the end!
	for _, e := range l.expressions {
		e.generateCode(asm, &l.block.symbolTable)
		asm.program = append(asm.program, [3]string{"  ", "pop", register})
		asm.program = append(asm.program, [3]string{"  ", "cmp", fmt.Sprintf("%v, 0", register)})
		asm.program = append(asm.program, [3]string{"  ", "je", endLabel})
	}
	// So if we getVar here, all expressions were true. So jump to start again.
	asm.program = append(asm.program, [3]string{"  ", "jmp", startLabel})
	asm.program = append(asm.program, [3]string{"", endLabel + ":", ""})
}

func (b Block) generateCode(asm *ASM, s *SymbolTable) {

	for _, statement := range b.statements {
		statement.generateCode(asm, &b.symbolTable)
	}

}

// By convention, the first six integer arguments are passed in the following registers:
// 		%rdi,%rsi,%rdx,%rcx,%r8,%r9
// Additional integer arguments are passed on the stack.
// TODO: Not yet implemented
// The first eight floating point arguments are passed in the SSE registers:
//		%xmm0,%xmm1, ...,%xmm7
// Additional floating arguments are passed on the stack.
//
// For calling functions (C) with a variable count of arguments, the register %al (rax) must be set with
// how many %xmm registers are used.
func (f Function) generateCode(asm *ASM, s *SymbolTable) {

	// As function declarations can not be nesting in assembler, we save the current program slice,
	// provide an empty 'program' to fill into and move that into the function part!
	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asmName := asm.nextFunctionName()
	asm.program = append(asm.program, [3]string{"", "global " + asmName, ""})
	asm.program = append(asm.program, [3]string{"", asmName + ":", ""})

	//asm.program = append(asm.program, [3]string{"  ", "pop", register})

	// Create local variables for all function parameters
	// Pop n parameters from stack and move into local variables
	// Needs to be in reversed order so we can push in correct order when calling the function
	for i := len(f.parameters) - 1; i >= 0; i-- {
		v := f.parameters[i]
		vName := asm.nextVariableName()
		asm.variables = append(asm.variables, [3]string{vName, "dq", "0"})
		f.block.symbolTable.setAsmName(v.vName, vName)

		register, _ := getRegister(TYPE_INT)
		asm.program = append(asm.program, [3]string{"  ", "pop", register})

		// Move value from register of expression into variable!
		asm.program = append(asm.program, [3]string{"  ", "mov", fmt.Sprintf("qword [%v], %v", vName, register)})
		debugPrint(asm, vName, v.vType)
	}

	// Generate block code
	f.block.generateCode(asm, s)

	// Generate code for each return expression, they will all accumulate on the stack automatically!
	// Careful, the will need to be popped in reversed order when function is called!
	for _, e := range f.returns {
		e.generateCode(asm, &f.block.symbolTable)
	}

	asm.program = append(asm.program, [3]string{"  ", "ret", ""})

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) generateCode() ASM {

	asm := ASM{}

	asm.header = append(asm.header, "extern printf  ; C function we need for debugging")
	asm.header = append(asm.header, "section .data")

	asm.constants = append(asm.constants, [2]string{"TRUE", "-1"})
	asm.constants = append(asm.constants, [2]string{"FALSE", "0"})

	asm.variables = append(asm.variables, [3]string{"fmti", "db", "\"%i\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"fmtf", "db", "\"%.2f\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"negOneF", "dq", "-1.0"})
	asm.variables = append(asm.variables, [3]string{"negOneI", "dq", "-1"})

	asm.sectionText = append(asm.sectionText, "section .text")

	asm.program = append(asm.program, [3]string{"", "global _start", ""})
	asm.program = append(asm.program, [3]string{"", "_start:", ""})

	ast.block.generateCode(&asm, &ast.globalSymbolTable)

	asm.program = append(asm.program, [3]string{"  ", "; Exit the program nicely", ""})
	asm.program = append(asm.program, [3]string{"  ", "mov", "rbx, 0  ; normal exit code"})
	asm.program = append(asm.program, [3]string{"  ", "mov", "rax, 1  ; process termination service (?)"})
	asm.program = append(asm.program, [3]string{"  ", "int", "0x80    ; linux kernel service"})

	return asm
}
