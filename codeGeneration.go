package main

import (
	"fmt"
)

type ASM struct {
	header []string
	//constants [][2]string

	// value -> name
	constants map[string]string

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

func (asm *ASM) addConstant(value string) string {

	if v, ok := asm.constants[value]; ok {
		return v
	}
	asm.constName += 1
	name := fmt.Sprintf("const_%v", asm.constName-1)
	asm.constants[value] = name

	return name
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

func (asm *ASM) addLabel(label string) {
	asm.program = append(asm.program, [3]string{"", label + ":", ""})
}
func (asm *ASM) addFun(name string) {
	asm.program = append(asm.program, [3]string{"", "global " + name, ""})
	asm.program = append(asm.program, [3]string{"", name + ":"})
}
func (asm *ASM) addLine(command, args string) {
	asm.program = append(asm.program, [3]string{"  ", command, args})
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
		return "idiv"
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
	case TYPE_INT, TYPE_FLOAT:
		name = asm.addConstant(c.cValue)
	case TYPE_STRING:
		name = asm.addConstant(fmt.Sprintf("\"%v\", 0", c.cValue))
	case TYPE_BOOL:
		v := "0"
		if c.cValue == "true" {
			v = "-1"
		}
		name = asm.addConstant(v)
	default:
		panic("Could not generate code for Const. Unknown type!")
	}

	register, _ := getRegister(TYPE_INT)
	asm.addLine("mov", fmt.Sprintf("%v, %v", register, name))
	asm.addLine("push", register)
}

func (v Variable) generateCode(asm *ASM, s *SymbolTable) {

	if symbol, ok := s.getVar(v.vName); ok {
		// We just push it on the stack. So here we can't use floating point registers for example!
		register, _ := getRegister(TYPE_INT)
		asm.addLine("mov", fmt.Sprintf("%v, qword [%v]", register, symbol.varName))
		asm.addLine("push", register)
		return
	}
	panic("Could not generate code for Variable. No symbol known!")
}

func (u UnaryOp) generateCode(asm *ASM, s *SymbolTable) {

	u.expr.generateCode(asm, s)

	stackReg, _ := getRegister(TYPE_INT)
	asm.addLine("pop", stackReg)

	//register, _ := getRegister(u.getExpressionType())

	if u.getResultCount() != 1 {
		panic("Code generation error: Unary expression can only handle one result")
	}
	t := u.getExpressionTypes()[0]

	switch t {
	case TYPE_BOOL:
		if u.operator == OP_NOT {
			// 'not' switches between 0 and -1. So False: 0, True: -1
			//asm.addLine("pop", register})
			asm.addLine("not", stackReg)
		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType))
		}
	case TYPE_INT:
		if u.operator == OP_NEGATIVE {
			//asm.addLine("pop", register})
			asm.addLine("neg", stackReg)

		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType))
		}
	case TYPE_FLOAT:
		if u.operator == OP_NEGATIVE {
			//asm.addLine("pop", register})

			fReg, _ := getRegister(TYPE_FLOAT)
			asm.addLine("movq", fmt.Sprintf("%v, %v", fReg, stackReg))
			asm.addLine("mulsd", fmt.Sprintf("%v, qword [negOneF]", fReg))
			asm.addLine("movq", fmt.Sprintf("%v, %v", stackReg, fReg))

		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v", u.operator, u.opType))
		}
	case TYPE_STRING:
		panic("Code generation error. No unary expression for Type String")
		return
	}

	asm.addLine("push", stackReg)
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
		asm.addLine("cmp", fmt.Sprintf("%v, %v", rLeft, rRight))
		asm.addLine(jump, labelTrue)
		asm.addLine("mov", fmt.Sprintf("%v, 0", rLeft))
		asm.addLine("jmp", labelOK)
		asm.addLabel(labelTrue)
		asm.addLine("mov", fmt.Sprintf("%v, -1", rLeft))
		asm.addLabel(labelOK)

	default:
		command := getCommand(t, op)
		switch t {
		case TYPE_INT:
			asm.addLine(command, fmt.Sprintf("%v, %v", rLeft, rRight))
		case TYPE_FLOAT:
			fLeft, fRight := getRegister(TYPE_FLOAT)
			// We need to move the values from stack/int registers to the corresponding xmm* ones and back afterwards
			// So the stack push works!
			asm.addLine("movq", fmt.Sprintf("%v, %v", fLeft, rLeft))
			asm.addLine("movq", fmt.Sprintf("%v, %v", fRight, rRight))
			asm.addLine(command, fmt.Sprintf("%v, %v", fLeft, fRight))
			asm.addLine("movq", fmt.Sprintf("%v, %v", rLeft, fLeft))
			asm.addLine("movq", fmt.Sprintf("%v, %v", rRight, fRight))
		}
	}
}

func (b BinaryOp) generateCode(asm *ASM, s *SymbolTable) {

	b.leftExpr.generateCode(asm, s)
	b.rightExpr.generateCode(asm, s)

	rLeft, rRight := getRegister(TYPE_INT)

	asm.addLine("pop", rRight)
	asm.addLine("pop", rLeft)

	if b.leftExpr.getResultCount() != 1 {
		panic("Code generation error: Binary expression can only handle one result each")
	}
	if b.rightExpr.getResultCount() != 1 {
		panic("Code generation error: Binary expression can only handle one result each")
	}
	t := b.leftExpr.getExpressionTypes()[0]

	switch t {
	case TYPE_INT, TYPE_FLOAT:
		binaryOperationNumber(b.operator, b.opType, rLeft, rRight, asm)
	case TYPE_BOOL:
		// Equal and unequal are identical for bool or int, as a bool is an integer type.
		if b.operator == OP_EQ || b.operator == OP_NE {
			binaryOperationNumber(b.operator, b.opType, rLeft, rRight, asm)
		} else {
			command := getCommand(TYPE_BOOL, b.operator)
			asm.addLine(command, fmt.Sprintf("%v, %v", rLeft, rRight))
		}

	case TYPE_STRING:
		panic("Code generation error: Strings not supported yet.")
	default:
		panic(fmt.Sprintf("Code generation error: Unknown operation type %v\n", int(b.opType)))
	}

	asm.addLine("push", rLeft)
}

func (f FunCall) generateCode(asm *ASM, s *SymbolTable) {

	for i := len(f.args) - 1; i >= 0; i-- {
		// The expression results will just accumulate on the stack!
		f.args[i].generateCode(asm, s)
	}

	if entry, ok := s.getFun(f.funName); ok {
		asm.addLine("call", entry.jumpLabel)
		// TODO: Clear stack: Is that correct? Do something else?
		//asm.addLine("add", fmt.Sprintf("rsp, %v", 8*len(entry.paramTypes)))
	} else {
		panic("Code generation error: Unknown function called")
	}

}

func debugPrint(asm *ASM, vName string, t Type) {

	format := "fmti"
	floatCount := "0"
	if t == TYPE_FLOAT {
		format = "fmtf"
		floatCount = "1"
		fReg, _ := getRegister(t)
		asm.addLine("movq", fmt.Sprintf("%v, qword [%v]", fReg, vName))
		asm.addLine("movq", "rsi, "+fReg)
	} else {
		asm.addLine("mov", fmt.Sprintf("rsi, qword [%v]", vName))
	}
	asm.addLine("mov", "rdi, "+format)
	asm.addLine("mov", "rax, "+floatCount)
	asm.addLine("call", "printf")
}

func (a Assignment) generateCode(asm *ASM, s *SymbolTable) {

	for _, e := range a.expressions {
		// Calculate expression
		e.generateCode(asm, s)
	}

	for _, v := range a.variables {

		// Just to get it from the stack into the variable.
		register, _ := getRegister(TYPE_INT)
		asm.addLine("pop", register)

		// Create corresponding variable, if it doesn't exist yet.
		if entry, ok := s.getVar(v.vName); !ok || entry.varName == "" {

			vName := asm.nextVariableName()
			// Variables are initialized with 0. This will be overwritten a few lines later!
			// TODO: What about float or string variables? Does this really work? Do we care at all because it will be overwritten anyway?
			asm.variables = append(asm.variables, [3]string{vName, "dq", "0"})
			s.setVarAsmName(v.vName, vName)
		}
		// This can not/should not fail!
		entry, _ := s.getVar(v.vName)
		vName := entry.varName

		// Move value from register of expression into variable!
		asm.addLine("mov", fmt.Sprintf("qword [%v], %v", vName, register))

	}

	for _, v := range a.variables {
		entry, _ := s.getVar(v.vName)
		debugPrint(asm, entry.varName, v.vType)
	}

}

func (c Condition) generateCode(asm *ASM, s *SymbolTable) {

	c.expression.generateCode(asm, s)

	register, _ := getRegister(TYPE_BOOL)
	// For now, we assume an else case. Even if it is just empty!
	elseLabel := asm.nextLabelName()
	endLabel := asm.nextLabelName()

	asm.addLine("pop", register)
	asm.addLine("cmp", fmt.Sprintf("%v, 0", register))
	asm.addLine("je", elseLabel)

	c.block.generateCode(asm, s)

	asm.addLine("jmp", endLabel)
	asm.addLabel(elseLabel)

	c.elseBlock.generateCode(asm, s)

	asm.addLabel(endLabel)
}

func (l Loop) generateCode(asm *ASM, s *SymbolTable) {

	register, _ := getRegister(TYPE_BOOL)
	startLabel := asm.nextLabelName()
	evalLabel := asm.nextLabelName()
	endLabel := asm.nextLabelName()

	// The initial assignment is logically moved inside the for-block
	l.assignment.generateCode(asm, &l.block.symbolTable)

	asm.addLine("jmp", evalLabel)
	asm.addLabel(startLabel)

	l.block.generateCode(asm, s)

	// The increment assignment is logically moved inside the for-block
	l.incrAssignment.generateCode(asm, &l.block.symbolTable)

	asm.addLabel(evalLabel)

	// If any of the expressions result in False (0), we jump to the end!
	for _, e := range l.expressions {
		e.generateCode(asm, &l.block.symbolTable)
		asm.addLine("pop", register)
		asm.addLine("cmp", fmt.Sprintf("%v, 0", register))
		asm.addLine("je", endLabel)
	}
	// So if we getVar here, all expressions were true. So jump to start again.
	asm.addLine("jmp", startLabel)
	asm.addLabel(endLabel)
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
	s.setFunAsmName(f.fName, asmName)
	asm.addFun(asmName)

	epilogueLabel := asm.nextLabelName()
	s.setFunEpilogueLabel(f.fName, epilogueLabel)

	// Save return address from top of stack to get to the arguments
	_, returnAddress := getRegister(TYPE_INT)
	asm.addLine("pop", returnAddress)

	// Create local variables for all function parameters
	// Pop n parameters from stack and move into local variables
	for _, v := range f.parameters {
		vName := asm.nextVariableName()
		asm.variables = append(asm.variables, [3]string{vName, "dq", "0"})
		f.block.symbolTable.setVarAsmName(v.vName, vName)

		register, _ := getRegister(TYPE_INT)
		asm.addLine("pop", register)

		// Move value from register of expression into variable!
		asm.addLine("mov", fmt.Sprintf("qword [%v], %v", vName, register))
		//debugPrint(asm, vName, v.vType)
	}

	// Push the return address back to the stack
	asm.addLine("push", returnAddress)

	//asm.addLine("push", "rbx")
	//asm.addLine("push", "rbp")

	// Generate block code
	f.block.generateCode(asm, s)

	//asm.addLine("pop", "rbp")
	//asm.addLine("pop", "rbx")

	//asm.addLine("ret", "")

	// TODO: Add proper Epilogue here:

	asm.addLabel(epilogueLabel)

	// Epilogue.

	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (r Return) generateCode(asm *ASM, s *SymbolTable) {
	// Generate code for each return expression, they will all accumulate on the stack automatically!
	// We do the first expression later!
	for i := len(r.expressions) - 2; i >= 0; i-- {
		r.expressions[i].generateCode(asm, s)
	}
	// First expression!
	if len(r.expressions) > 0 {
		r.expressions[len(r.expressions)-1].generateCode(asm, s)
		register, returnAddress := getRegister(TYPE_INT)
		asm.addLine("pop", register)
		// Save return address
		asm.addLine("mov", fmt.Sprintf("%v, [rsp+%v]", returnAddress, 8*(len(r.expressions)-1)))
		// Overwrite old return address with first return value
		asm.addLine("mov", fmt.Sprintf("[rsp+%v], %v", 8*(len(r.expressions)-1), register))
		// Push return address
		asm.addLine("push", returnAddress)
	}

	epilogueLabel := ""
	if entry, ok := s.getFun(s.activeFunctionName); ok {
		epilogueLabel = entry.epilogueLabel
	} else {
		panic("Code generation error. Function not in symbol table.")
	}

	asm.addLine("jmp", epilogueLabel)
	//asm.addLine("pop", "rbp")
	//asm.addLine("ret", "")
}

func (ast AST) generateCode() ASM {

	asm := ASM{}
	asm.constants = make(map[string]string, 0)

	asm.header = append(asm.header, "extern printf  ; C function we need for debugging")
	asm.header = append(asm.header, "section .data")

	asm.variables = append(asm.variables, [3]string{"fmti", "db", "\"%i\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"fmtf", "db", "\"%.2f\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"negOneF", "dq", "-1.0"})
	asm.variables = append(asm.variables, [3]string{"negOneI", "dq", "-1"})

	asm.sectionText = append(asm.sectionText, "section .text")

	asm.addFun("_start")

	ast.block.generateCode(&asm, &ast.globalSymbolTable)

	asm.addLine("; Exit the program nicely", "")
	asm.addLine("mov", "rbx, 0  ; exit code")
	asm.addLine("mov", "rax, 1  ; sys_exit (system call number)")
	asm.addLine("int", "0x80    ; call kernel")

	return asm
}
