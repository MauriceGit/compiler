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

	// Uncomment, when we go from stack to rax/xmm0!
	//	switch c.cType {
	//	case TYPE_INT:
	//		asm.addLine("mov", "rax, "+name)
	//	case TYPE_FLOAT:
	//		register, _ := getRegister(TYPE_INT)
	//		asm.addLine("mov", fmt.Sprintf("%v, %v", register, name))
	//		asm.addLine("movq", fmt.Sprintf("xmm0, %v", register))
	//	}

}

func (v Variable) generateCode(asm *ASM, s *SymbolTable) {

	if symbol, ok := s.getVar(v.vName); ok {
		// We just push it on the stack. So here we can't use floating point registers for example!
		register, _ := getRegister(TYPE_INT)
		sign := "+"
		if symbol.offset < 0 {
			sign = ""
		}
		asm.addLine("mov", fmt.Sprintf("%v, [rbp%v%v]", register, sign, symbol.offset))
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

func debugPrint(asm *ASM, offset int, t Type) {

	format := "fmti"
	floatCount := "0"
	sign := "+"
	if offset < 0 {
		sign = ""
	}
	if t == TYPE_FLOAT {
		format = "fmtf"
		floatCount = "1"
		fReg, _ := getRegister(t)

		asm.addLine("  movq", fmt.Sprintf("%v, [rbp%v%v]", fReg, sign, offset))
		asm.addLine("  movq", "rsi, "+fReg)
	} else {
		asm.addLine("  mov", fmt.Sprintf("rsi, [rbp%v%v]", sign, offset))
	}
	asm.addLine("  mov", "rdi, "+format)
	asm.addLine("  mov", "rax, "+floatCount)
	asm.addLine("  call", "printf")
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

		// This can not/should not fail!
		entry, _ := s.getVar(v.vName)

		sign := "+"
		if entry.offset < 0 {
			sign = ""
		}

		asm.addLine("mov", fmt.Sprintf("[rbp%v%v], %v", sign, entry.offset, register))
	}

	for _, v := range a.variables {
		entry, _ := s.getVar(v.vName)
		debugPrint(asm, entry.offset, v.vType)
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

// TODO: Move FunCall generateCode back to the expression section, instead of in-between statements!
func (f FunCall) generateCode(asm *ASM, s *SymbolTable) {

	intRegisters := getFunctionRegisters(TYPE_INT)
	intRegIndex := 0
	floatRegisters := getFunctionRegisters(TYPE_FLOAT)
	floatRegIndex := 0

	// In case we have multiple return values, we want to provide a pointer to the stack space to write to
	// as very first argument to the function. So we skip the first register here on purpose!
	if len(f.retTypes) > 1 {
		intRegIndex++
	}

	// Make space on (our) stack, if f has multiple return values!
	// The return stack space is now BEFORE the parameter stack space! So we can just pop used parameters
	// from the stack after the function call.
	if len(f.retTypes) > 1 {
		asm.addLine("sub", fmt.Sprintf("rsp, %v", 8*len(f.retTypes)))
		// rsp already points to one return stack space right now, so we only offset by count-1
		asm.addLine("lea", fmt.Sprintf("rdi, [rsp+%v]", 8*(len(f.retTypes)-1)))
	}

	// Expect value in rax/xmm0 instead of stack.
	// For floating point, the first parameter is already xmm0, so we are done.
	// Int must be moved into rax.
	if len(f.args) == 1 && f.args[0].getResultCount() == 1 {
		f.args[0].generateCode(asm, s)

		// TODO: Remove when switching to rax/xmm0
		if f.args[0].getExpressionTypes()[0] == TYPE_FLOAT {
			r, _ := getRegister(TYPE_INT)
			asm.addLine("pop", r)
			asm.addLine("movq", "xmm0, "+r)
		}

		if f.args[0].getExpressionTypes()[0] == TYPE_INT {

			// TODO: Remove when switching to rax/xmm0
			asm.addLine("pop", "rax")

			asm.addLine("mov", intRegisters[intRegIndex]+", rax")
		}
	} else {

		// First generate code for all expressions, THEN put them in function parameter registers!
		// To avoid register overwriting by functions!
		for i := len(f.args) - 1; i >= 0; i-- {
			f.args[i].generateCode(asm, s)
		}

		// Iterate in reverse order to correctly pop from stack!
		for _, e := range f.args {
			for _, t := range e.getExpressionTypes() {
				switch t {
				case TYPE_INT:
					// All further values stay on the stack and are not assigned to registers!
					if intRegIndex < len(intRegisters) {
						asm.addLine("pop", intRegisters[intRegIndex])
						intRegIndex++
					}
				case TYPE_FLOAT:
					if floatRegIndex < len(floatRegisters) {
						tmpR, _ := getRegister(TYPE_INT)
						asm.addLine("pop", tmpR)
						asm.addLine("movq", fmt.Sprintf("%v, %v", floatRegisters[floatRegIndex], tmpR))
						floatRegIndex++
					}
				}
			}
		}
	}

	if entry, ok := s.getFun(f.funName); ok {

		asm.addLine("call", entry.jumpLabel)

		paramCount := len(entry.paramTypes)
		if len(f.retTypes) > 1 {
			paramCount++
		}

		// Remove the stack space used for the parameters for the function by resetting rsp by the
		// amount of stack space we used.
		paramStackCount := paramCount - intRegIndex - floatRegIndex
		asm.addLine("add", fmt.Sprintf("rsp, %v    ; Remove function parameters from stack", 8*paramStackCount))

		// TODO: Remove when switching to rax/xmm0
		//		if len(f.retTypes) == 1 {
		//			asm.addLine()
		//		}

	} else {
		panic("Code generation error: Unknown function called")
	}

}

// extractAllFunctionVariables will go through all blocks in the function and extract
// the names of all existing local variables, together with a pointer to the corresponding
// symbol table so we can correctly set the assembler stack offsets.
type FunctionVar struct {
	name        string
	symbolTable *SymbolTable
}

func (b Block) extractAllFunctionVariables() (vars []FunctionVar) {

	tmpVars := b.symbolTable.getLocalVars()
	for _, v := range tmpVars {
		vars = append(vars, FunctionVar{v, &b.symbolTable})
	}

	for _, s := range b.statements {
		switch st := s.(type) {
		case Block:
			vars = append(vars, st.extractAllFunctionVariables()...)
		case Condition:
			vars = append(vars, st.block.extractAllFunctionVariables()...)
			vars = append(vars, st.elseBlock.extractAllFunctionVariables()...)
		case Loop:
			vars = append(vars, st.block.extractAllFunctionVariables()...)
		}
	}
	return
}

func varOnStack(vName string, varsOnStack []Variable) bool {
	for _, sv := range varsOnStack {
		if vName == sv.vName {
			return true
		}
	}
	return false
}

// assignFunVariableStackOffsets sets all stack offsets based on the initialOffset
// and returns the new offset on the stack.
// If there are variables passed on stack instead of registers, we calculate the positive offset
// we expect the variable at.
func (b Block) assignFunVariableStackOffsets(initialOffset int, varsOnStack []Variable) int {
	funNames := b.extractAllFunctionVariables()

	// We already pushed rbp initially in the prologue, so we have to add that
	positiveStackOffset := 8

	// Assign the symbol table variables their corresponding stack offset.
	for _, fn := range funNames {
		// Do not assign stack offsets for variables already on the stack!
		if varOnStack(fn.name, varsOnStack) {
			fn.symbolTable.setVarAsmOffset(fn.name, positiveStackOffset)
			positiveStackOffset += 8
			continue
		}

		fn.symbolTable.setVarAsmOffset(fn.name, -initialOffset)
		initialOffset += 8
	}

	return initialOffset
}

func addFunctionPrologue(asm *ASM, variableStackSpace int) {
	asm.addLine("; prologue", "")
	// That means, that we have to add 8 to all stack reads, relative to rbp!
	asm.addLine("push", "rbp")
	// Our new base pointer we can relate to within this function, to access stack local variables or parameter!
	asm.addLine("mov", "rbp, rsp")

	// This will add space for local variables on the stack!
	// TODO: This count must be change, when we allow parameter passing on the stack! As they are already on the stack

	// IMPORTANT: rbp/rsp must be 16-bit aligned (expected by OS and libraries), otherwise we segfault.
	if (variableStackSpace+8)%16 != 0 {
		variableStackSpace += 8
	}

	asm.addLine("sub", fmt.Sprintf("rsp, %v", variableStackSpace))
	asm.addLine("push", "rdi")
	asm.addLine("push", "rsi")
	asm.addLine("", "")
}

func addFunctionEpilogue(asm *ASM) {
	asm.addLine("; epilogue", "")
	// Recover saved registers
	asm.addLine("pop", "rsi")
	asm.addLine("pop", "rdi")
	// Dealocate local variables by overwriting the stack pointer!
	asm.addLine("mov", "rsp, rbp")
	// Restore the original/callers base pointer
	asm.addLine("pop", "rbp")
	asm.addLine("", "")
}

func filterVarType(t Type, vars []Variable) (out []Variable) {
	for _, v := range vars {
		if v.vType == t {
			out = append(out, v)
		}
	}
	return
}

// getParametersOnStack filters the given parameter/variable list an returns lists of variables by type
// that are placed on the stack instead of registers!
func getParametersOnStack(parameters []Variable, isMultiReturn bool) (intStack, floatStack []Variable) {

	intRegisters := getFunctionRegisters(TYPE_INT)
	floatRegisters := getFunctionRegisters(TYPE_FLOAT)

	// If we have a multi-return function, the first parameter is used as a pointer to the stack.
	additionalRegister := 0
	if isMultiReturn {
		additionalRegister = 1
	}

	intVars := filterVarType(TYPE_INT, parameters)
	// Remaining variables are placed on the stack instead of registers.
	if len(intVars) > len(intRegisters)+additionalRegister {
		intStack = intVars[len(intRegisters)+additionalRegister:]
	}

	floatVars := filterVarType(TYPE_FLOAT, parameters)
	// Remaining variables are placed on the stack instead of registers.
	if len(floatVars) > len(floatRegisters) {
		floatStack = floatVars[len(floatRegisters):]
	}
	return
}

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

	intOnStack, floatOnStack := getParametersOnStack(f.parameters, len(f.returnTypes) > 1)
	varsOnStack := append(intOnStack, floatOnStack...)

	// In the prologue, we save both rdi and rsi on the stack.
	// That means, that we have to add 2*8 on top of the variables we want to refer on the stack!
	// This will also include parameters from function headers, as they are handled as assignment in the
	// semantic analyzer.
	varStackIndex := f.block.assignFunVariableStackOffsets(2*8, varsOnStack)

	// Prologue:
	addFunctionPrologue(asm, varStackIndex-2*8)

	intRegisters := getFunctionRegisters(TYPE_INT)
	intRegIndex := 0
	floatRegisters := getFunctionRegisters(TYPE_FLOAT)
	floatRegIndex := 0

	// Save stack pointer for return values
	if len(f.returnTypes) > 1 {
		asm.addLine("mov", fmt.Sprintf("[rbp%v], %v", -8, intRegisters[intRegIndex]))
		intRegIndex++
		s.setFunReturnStackPointer(f.fName, -8)
	}

	// Create local variables for all function parameters
	for _, v := range f.parameters {

		entry, _ := f.block.symbolTable.getVar(v.vName)
		sign := "+"
		if entry.offset < 0 {
			sign = ""
		}
		switch v.vType {
		case TYPE_INT:
			if intRegIndex < len(intRegisters) {
				asm.addLine("mov", fmt.Sprintf("[rbp%v%v], %v", sign, entry.offset, intRegisters[intRegIndex]))
				intRegIndex++
			}
		case TYPE_FLOAT:
			if floatRegIndex < len(floatRegisters) {
				asm.addLine("movq", fmt.Sprintf("[rbp%v%v], %v", sign, entry.offset, floatRegisters[floatRegIndex]))
				floatRegIndex++
			}
		}
	}

	// Generate block code
	f.block.generateCode(asm, s)

	// Epilogue.
	asm.addLabel(epilogueLabel)

	addFunctionEpilogue(asm)

	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (r Return) generateCode(asm *ASM, s *SymbolTable) {

	entry, ok := s.getFun(s.activeFunctionName)
	if !ok {
		panic("Code generation error. Function not in symbol table.")
	}

	for _, e := range r.expressions {
		e.generateCode(asm, s)

		// TODO: Uncomment when we switch to rax/xmm0
		//		// Make sure everything is actually on the stack!
		//		if r.expressions[i].getResultCount() == 1 {
		//			switch r.expressions[i].getExpressionTypes()[0] {
		//			case TYPE_INT:
		//				asm.addLine("push", "rax")
		//			case TYPE_FLOAT:
		//				tmpR, _ := getRegister(TYPE_INT)
		//				asm.addLine("movq", fmt.Sprintf("%v, xmm0", tmpR))
		//				asm.addLine("push", tmpR)
		//			}
		//		}
	}

	// If we only have one return value, we want it in rax/xmm0. But that should already be the case anyway...
	if len(entry.returnTypes) > 1 {
		// For multiple expressions/return values, they will assemble on the stack in reverse order.
		// So we have to pop them and move them onto the pre-allocated stack space!

		tmpR, stackPointerReg := getRegister(TYPE_INT)
		asm.addLine("mov", fmt.Sprintf("%v, [rbp%v]", stackPointerReg, entry.returnStackPointerOffset))

		offset := 0
		for _ = range entry.returnTypes {
			asm.addLine("pop", tmpR)
			asm.addLine("mov", fmt.Sprintf("[%v-%v], %v", stackPointerReg, offset, tmpR))
			offset += 8
		}

	}

	asm.addLine("jmp", entry.epilogueLabel)
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

	stackOffset := ast.block.assignFunVariableStackOffsets(0, []Variable{})
	addFunctionPrologue(&asm, stackOffset)

	ast.block.generateCode(&asm, &ast.globalSymbolTable)

	addFunctionEpilogue(&asm)

	asm.addLine("; Exit the program nicely", "")
	asm.addLine("mov", "rbx, 0  ; exit code")
	asm.addLine("mov", "rax, 1  ; sys_exit (system call number)")
	asm.addLine("int", "0x80    ; call kernel")

	return asm
}
