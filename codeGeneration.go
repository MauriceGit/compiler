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
	case OP_DIV, OP_MOD:
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

func getMov(t Type) string {
	switch t {
	case TYPE_INT, TYPE_BOOL:
		return "mov"
	case TYPE_FLOAT:
		return "movq"
	case TYPE_STRING:
		panic("Code generation error. String register not yet implemented")
	}
	return ""
}

func getReturnRegister(t Type) string {
	switch t {
	case TYPE_INT, TYPE_BOOL:
		return "rax"
	case TYPE_FLOAT:
		return "xmm0"
	case TYPE_STRING:
		panic("Code generation error. String register not yet implemented")
	}
	return ""
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

	switch c.cType {
	case TYPE_INT, TYPE_BOOL:
		asm.addLine("mov", "rax, "+name)
	case TYPE_FLOAT:
		register, _ := getRegister(TYPE_INT)
		asm.addLine("mov", fmt.Sprintf("%v, %v", register, name))
		asm.addLine("movq", fmt.Sprintf("xmm0, %v", register))
	}

}

// We don't know where we were before and what expression needs to be indexed.
// All we care about, is that the array pointer is in rax and the resulting value will also be in rax!
func generateArrayAccessCode(indexExpressions []Expression, asm *ASM, s *SymbolTable) {

	for _, indexExpression := range indexExpressions {
		index, address := getRegister(TYPE_INT)
		// address = rax
		asm.addLine("mov", fmt.Sprintf("%v, %v", address, getReturnRegister(TYPE_INT)))
		// Save the address
		asm.addLine("push", address)
		indexExpression.generateCode(asm, s)
		asm.addLine("pop", address)
		// TODO: Check, if this following mov can entirely be removed, by just using rax as index directly!
		asm.addLine("mov", fmt.Sprintf("%v, %v", index, getReturnRegister(TYPE_INT)))
		asm.addLine("mov", fmt.Sprintf("%v, [%v+%v*8+8]", getReturnRegister(TYPE_INT), address, index))
	}
}

func (v Variable) generateCode(asm *ASM, s *SymbolTable) {

	if symbol, ok := s.getVar(v.vName); ok {
		// We just push it on the stack. So here we can't use floating point registers for example!
		sign := "+"
		if symbol.offset < 0 {
			sign = ""
		}

		switch v.vType.t {
		case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
			asm.addLine("mov", fmt.Sprintf("%v, [rbp%v%v]", getReturnRegister(TYPE_INT), sign, symbol.offset))
		case TYPE_FLOAT:
			register, _ := getRegister(TYPE_INT)
			asm.addLine("mov", fmt.Sprintf("%v, [rbp%v%v]", register, sign, symbol.offset))
			asm.addLine("movq", fmt.Sprintf("%v, %v", getReturnRegister(TYPE_FLOAT), register))
		}

		if len(v.indexExpressions) > 0 {
			generateArrayAccessCode(v.indexExpressions, asm, s)
			if v.vType.subType.t == TYPE_FLOAT {
				asm.addLine("movq", fmt.Sprintf("%v, %v", getReturnRegister(TYPE_FLOAT), getReturnRegister(TYPE_INT)))
			}
		}

		return
	}
	panic("Could not generate code for Variable. No symbol known!")
}

func (u UnaryOp) generateCode(asm *ASM, s *SymbolTable) {

	u.expr.generateCode(asm, s)

	if u.getResultCount() != 1 {
		panic("Code generation error: Unary expression can only handle one result")
	}
	t := u.getExpressionTypes()[0]

	switch t.t {
	case TYPE_BOOL:
		if u.operator == OP_NOT {
			// 'not' switches between 0 and -1. So False: 0, True: -1
			asm.addLine("not", "rax")
		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType))
		}
	case TYPE_INT:
		if u.operator == OP_NEGATIVE {
			asm.addLine("neg", "rax")

		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v\n", u.operator, u.opType))
		}
	case TYPE_FLOAT:
		if u.operator == OP_NEGATIVE {
			asm.addLine("mulsd", "xmm0, qword [negOneF]")

		} else {
			panic(fmt.Sprintf("Code generation error. Unexpected unary type: %v for %v", u.operator, u.opType))
		}
	case TYPE_STRING:
		panic("Code generation error. No unary expression for Type String")
		return
	}
}

// binaryOperationFloat executes the operation on the two registers and writes the result into rLeft!
func binaryOperationNumber(op Operator, t Type, rLeft, rRight string, asm *ASM) {

	switch op {
	case OP_GE, OP_GREATER, OP_LESS, OP_LE, OP_EQ, OP_NE:

		// Works for anything that should be compared.
		labelTrue := asm.nextLabelName()
		labelOK := asm.nextLabelName()
		jump := getJumpType(op)

		asm.addLine("cmp", fmt.Sprintf("%v, %v", rLeft, rRight))
		asm.addLine(jump, labelTrue)
		asm.addLine("mov", fmt.Sprintf("%v, 0", rLeft))
		asm.addLine("jmp", labelOK)
		asm.addLabel(labelTrue)
		asm.addLine("mov", fmt.Sprintf("%v, -1", rLeft))
		asm.addLabel(labelOK)

	// add, sub, mul, div, mod...
	default:
		command := getCommand(t, op)
		if op == OP_DIV || op == OP_MOD {
			// Clear rdx! Needed for at least div and mod.
			asm.addLine("xor", "rdx, rdx")
		}
		// idiv calculates both the integer division, as well as the remainder.
		// The reminder will be written to rdx. So move back to rax.
		if op == OP_MOD {
			asm.addLine(command, fmt.Sprintf("%v, %v", rLeft, rRight))
			asm.addLine("mov", fmt.Sprintf("%v, rdx", rLeft))
		} else {
			asm.addLine(command, fmt.Sprintf("%v, %v", rLeft, rRight))
		}

	}
}

func (b BinaryOp) generateCode(asm *ASM, s *SymbolTable) {

	if b.leftExpr.getResultCount() != 1 {
		panic("Code generation error: Binary expression can only handle one result each")
	}
	if b.rightExpr.getResultCount() != 1 {
		panic("Code generation error: Binary expression can only handle one result each")
	}
	t := b.leftExpr.getExpressionTypes()[0]

	b.rightExpr.generateCode(asm, s)

	// mov rsi, rax
	// mov xmm1, xmm0

	//asm.addLine(getMov(t), fmt.Sprintf("%v, %v", rRight, getReturnRegister(t)))

	switch t.t {
	case TYPE_INT, TYPE_BOOL:
		asm.addLine("push", getReturnRegister(t.t))
	case TYPE_FLOAT:
		tmpR, _ := getRegister(TYPE_INT)
		asm.addLine("movq", fmt.Sprintf("%v, %v", tmpR, getReturnRegister(t.t)))
		asm.addLine("push", tmpR)
	}

	// Result will be left in rax/xmm0
	b.leftExpr.generateCode(asm, s)

	rLeft := getReturnRegister(t.t)

	_, rRight := getRegister(t.t)
	asm.addLine("pop", rRight)

	switch t.t {
	case TYPE_INT, TYPE_FLOAT:
		binaryOperationNumber(b.operator, b.opType.t, rLeft, rRight, asm)
	case TYPE_BOOL:
		// Equal and unequal are identical for bool or int, as a bool is an integer type.
		if b.operator == OP_EQ || b.operator == OP_NE {
			binaryOperationNumber(b.operator, b.opType.t, rLeft, rRight, asm)
		} else {
			panic("Code generation error. Unknown operator for bool (I think?).")
		}

	case TYPE_STRING:
		panic("Code generation error: Strings not supported yet.")
	default:
		panic(fmt.Sprintf("Code generation error: Unknown operation type %v\n", int(b.opType.t)))
	}
}

func (a Array) generateCode(asm *ASM, s *SymbolTable) {

	arg0 := getFunctionRegisters(TYPE_INT)[0]
	ret := getReturnRegister(TYPE_INT)

	// determine the first memory location by allocating zero byte.
	asm.addLine("mov", "rax, 12    ; sys_brk")
	asm.addLine("mov", arg0+", 0")
	asm.addLine("syscall", "")

	// Save our first memory register!
	// We use rsi manually here as a register that is Caller-saved and will not be
	// altered by anything that happens while creating/allocating/filling the array (well, only by myself)
	asm.addLine("mov", "rsi, "+ret)

	// n bytes is what we want to allocate here!
	// rax is the pointer to the first memory with 0 byte space
	// So we use the pointer rax and extend our memory. This is also, how we
	// can extend memory dynamically or free it (by setting it to [rax+0])

	// lea rdi, [rax+n]
	asm.addLine("lea", fmt.Sprintf("%v, [%v+%v]", arg0, ret, a.aCount*8+8))
	asm.addLine("mov", "rax, 12    ; sys_brk")
	asm.addLine("syscall", "")

	// Check memory error
	asm.addLine("cmp", "rax, 0")
	asm.addLine("jl", "exit")

	// Initialize array with 0s, if they are not initialized.
	if len(a.aExpressions) == 0 {
		// rdi is now the highest available address
		asm.addLine("mov", fmt.Sprintf("%v, %v", arg0, ret))
		// rdi now points to the last available qword
		asm.addLine("sub", arg0+", 8")
		// Number of qwords allocated. Not really sure, why we use rcx here though.
		asm.addLine("mov", fmt.Sprintf("rcx, %v", a.aCount+1))

		// rax == What to write into memory (default). So we fill it up with 0s
		asm.addLine("xor", "rax, rax")
		// go backwards in memory
		asm.addLine("std", "")
		// store the value of rax into rcx at every position!
		asm.addLine("rep", "stosq")
		// clear the flag set by 'std'
		asm.addLine("cld", "")
	} else {
		for i := len(a.aExpressions) - 1; i >= 0; i-- {
			e := a.aExpressions[i]
			// Calculate expression
			e.generateCode(asm, s)
			// Multiple return values are already on the stack, single ones not!
			if e.getResultCount() == 1 {
				switch e.getExpressionTypes()[0].t {
				case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
					asm.addLine("push", getReturnRegister(TYPE_INT))
				case TYPE_FLOAT:
					tmpR, _ := getRegister(TYPE_INT)
					asm.addLine("movq", fmt.Sprintf("%v, %v", tmpR, getReturnRegister(TYPE_FLOAT)))
					asm.addLine("push", tmpR)
				}
			}
		}

		register, _ := getRegister(TYPE_INT)
		for i := 0; i < a.aCount; i++ {
			asm.addLine("pop", register)
			asm.addLine("mov", fmt.Sprintf("[rsi+%v], %v", i*8+8, register))
		}
	}

	// Write length at first position in array!
	register, _ := getRegister(TYPE_INT)
	asm.addLine("mov", fmt.Sprintf("%v, %v", register, a.aCount))
	asm.addLine("mov", fmt.Sprintf("[rsi], %v", register))

	asm.addLine("mov", ret+", rsi")

	if len(a.indexExpressions) > 0 {
		generateArrayAccessCode(a.indexExpressions, asm, s)
		if a.getExpressionTypes()[0].t == TYPE_FLOAT {
			asm.addLine("movq", fmt.Sprintf("%v, %v", getReturnRegister(TYPE_FLOAT), getReturnRegister(TYPE_INT)))
		}
	}

}

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
		//asm.addLine("lea", fmt.Sprintf("rdi, [rsp+%v]", 8*(len(f.retTypes)-1)))
		asm.addLine("lea", fmt.Sprintf("rdi, [rsp]"))
	}

	// Expect value in rax/xmm0 instead of stack.
	// For floating point, the first parameter is already xmm0, so we are done.
	// Int must be moved into rax.
	if len(f.args) == 1 && f.args[0].getResultCount() == 1 {
		f.args[0].generateCode(asm, s)

		switch f.args[0].getExpressionTypes()[0].t {
		case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
			asm.addLine("mov", intRegisters[intRegIndex]+", rax")
			intRegIndex++
		case TYPE_FLOAT:
			floatRegIndex++
		}

	} else {

		// First generate code for all expressions, THEN put them in function parameter registers!
		// To avoid register overwriting by functions!
		for i := len(f.args) - 1; i >= 0; i-- {
			e := f.args[i]
			e.generateCode(asm, s)

			// Multiple return values are already on the stack, single ones not!
			if e.getResultCount() == 1 {
				switch e.getExpressionTypes()[0].t {
				case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
					asm.addLine("push", getReturnRegister(TYPE_INT))
				case TYPE_FLOAT:
					tmpR, _ := getRegister(TYPE_INT)
					asm.addLine("movq", fmt.Sprintf("%v, %v", tmpR, getReturnRegister(TYPE_FLOAT)))
					asm.addLine("push", tmpR)
				}
			}
		}

		// Iterate in reverse order to correctly pop from stack!
		for _, e := range f.args {
			for _, t := range e.getExpressionTypes() {
				switch t.t {
				case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
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

		if paramStackCount > 0 {
			asm.addLine("add", fmt.Sprintf("rsp, %v    ; Remove function parameters from stack", 8*paramStackCount))
		}

	} else {
		panic("Code generation error: Unknown function called")
	}

	if len(f.indexExpressions) > 0 {
		generateArrayAccessCode(f.indexExpressions, asm, s)
		if f.retTypes[0].subType.t == TYPE_FLOAT {
			asm.addLine("movq", fmt.Sprintf("%v, %v", getReturnRegister(TYPE_FLOAT), getReturnRegister(TYPE_INT)))
		}
	}

}

func (a Assignment) generateCode(asm *ASM, s *SymbolTable) {

	for i := len(a.expressions) - 1; i >= 0; i-- {
		e := a.expressions[i]
		// Calculate expression
		e.generateCode(asm, s)
		// Multiple return values are already on the stack, single ones not!
		if e.getResultCount() == 1 {
			switch e.getExpressionTypes()[0].t {
			case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
				asm.addLine("push", getReturnRegister(TYPE_INT))
			case TYPE_FLOAT:
				tmpR, _ := getRegister(TYPE_INT)
				asm.addLine("movq", fmt.Sprintf("%v, %v", tmpR, getReturnRegister(TYPE_FLOAT)))
				asm.addLine("push", tmpR)
			}
		}
	}

	// Just to get it from the stack into the variable.
	valueRegister, _ := getRegister(TYPE_INT)
	for _, v := range a.variables {
		asm.addLine("pop", valueRegister)

		// This can not/should not fail!
		entry, _ := s.getVar(v.vName)

		sign := "+"
		if entry.offset < 0 {
			sign = ""
		}

		if len(v.indexExpressions) > 0 {
			index, address := getRegister(TYPE_INT)

			asm.addLine("mov", fmt.Sprintf("%v, [rbp%v%v]", address, sign, entry.offset))
			asm.addLine("push", valueRegister)

			// We go through the indexing to determine the last index and address to write to!
			for _, indexExpression := range v.indexExpressions {

				asm.addLine("push", address)
				indexExpression.generateCode(asm, s)
				asm.addLine("mov", fmt.Sprintf("%v, %v", index, getReturnRegister(TYPE_INT)))
				asm.addLine("pop", address)

				// Go one level down.
				asm.addLine("lea", fmt.Sprintf("%v, [%v+%v*8+8]", address, address, index))
				//asm.addLine("mov", fmt.Sprintf("%v, [%v]", address, address))
			}

			asm.addLine("pop", valueRegister)
			asm.addLine("mov", fmt.Sprintf("[%v], %v", address, valueRegister))

		} else {
			asm.addLine("mov", fmt.Sprintf("[rbp%v%v], %v", sign, entry.offset, valueRegister))
		}

	}
}

func (c Condition) generateCode(asm *ASM, s *SymbolTable) {

	c.expression.generateCode(asm, s)

	register, _ := getRegister(TYPE_BOOL)
	// For now, we assume an else case. Even if it is just empty!
	elseLabel := asm.nextLabelName()
	endLabel := asm.nextLabelName()

	switch c.expression.getExpressionTypes()[0].t {
	case TYPE_BOOL, TYPE_INT:
		register = getReturnRegister(TYPE_INT)
	case TYPE_FLOAT:
		asm.addLine("movq", fmt.Sprintf("%v, %v", register, getReturnRegister(TYPE_FLOAT)))
	}

	//asm.addLine("pop", register)
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

		switch e.getExpressionTypes()[0].t {
		case TYPE_INT, TYPE_BOOL:
			register = getReturnRegister(TYPE_INT)
		case TYPE_FLOAT:
			asm.addLine("movq", fmt.Sprintf("%v, %v", register, getReturnRegister(TYPE_FLOAT)))
		}

		//asm.addLine("pop", register)
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

// extractAllFunctionVariables will go through all blocks in the function and extract
// the names of all existing local variables, together with a pointer to the corresponding
// symbol table so we can correctly set the assembler stack offsets.
type FunctionVar struct {
	name        string
	symbolTable *SymbolTable
	isRoot      bool
}

// We need the isRoot flag to make sure, that we only expect parameters on the stack, that
// are actually parameters, not variables of same name/type deeper down in the function!
func (b Block) extractAllFunctionVariables(isRoot bool) (vars []FunctionVar) {

	tmpVars := b.symbolTable.getLocalVars()
	for _, v := range tmpVars {
		vars = append(vars, FunctionVar{v, &b.symbolTable, isRoot})
	}

	for _, s := range b.statements {
		switch st := s.(type) {
		case Block:
			vars = append(vars, st.extractAllFunctionVariables(false)...)
		case Condition:
			vars = append(vars, st.block.extractAllFunctionVariables(false)...)
			vars = append(vars, st.elseBlock.extractAllFunctionVariables(false)...)
		case Loop:
			vars = append(vars, st.block.extractAllFunctionVariables(false)...)
		}
	}
	return
}

func addFunctionPrologue(asm *ASM, variableStackSpace int) {
	asm.addLine("; prologue", "")
	// That means, that we have to add 8 to all stack reads, relative to rbp!
	asm.addLine("push", "rbp")
	// Our new base pointer we can relate to within this function, to access stack local variables or parameter!
	asm.addLine("mov", "rbp, rsp")

	// IMPORTANT: rbp/rsp must be 16-bit aligned (expected by OS and libraries), otherwise we segfault.
	if (variableStackSpace+8)%16 != 0 && variableStackSpace != 0 {
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
		if v.vType.t == t {
			out = append(out, v)
		}
	}
	return
}

// getParametersOnStack filters the given parameter/variable list an returns lists of variables by type
// that are placed on the stack instead of registers!
func getParametersOnStack(parameters []Variable, isMultiReturn bool) (map[string]bool, map[string]bool) {

	intStack := make(map[string]bool, 0)
	floatStack := make(map[string]bool, 0)

	intRegisters := getFunctionRegisters(TYPE_INT)
	floatRegisters := getFunctionRegisters(TYPE_FLOAT)

	// If we have a multi-return function, the first parameter is used as a pointer to the stack.
	additionalRegister := 0
	if isMultiReturn {
		additionalRegister = 1
	}

	intVars := filterVarType(TYPE_INT, parameters)
	// Remaining variables are placed on the stack instead of registers.
	if len(intVars)+additionalRegister > len(intRegisters) {
		for _, v := range intVars[len(intRegisters)-additionalRegister:] {
			intStack[v.vName] = true
		}
	}

	floatVars := filterVarType(TYPE_FLOAT, parameters)
	// Remaining variables are placed on the stack instead of registers.
	if len(floatVars) > len(floatRegisters) {
		for _, v := range floatVars[len(floatRegisters):] {
			floatStack[v.vName] = true
		}
	}
	return intStack, floatStack
}

// Parameters on stack must be registered in the root symbol table!
func varOnStack(v FunctionVar, intStackVars, floatStackVars map[string]bool) bool {
	if !v.isRoot {
		return false
	}
	if _, ok := intStackVars[v.name]; ok {
		return true
	}
	if _, ok := floatStackVars[v.name]; ok {
		return true
	}
	return false
}

func (b Block) assignVariableStackOffset(parameters []Variable, isMultiReturn bool) int {
	// Local variables go down in the stack, so offsets are calculated  -= 8
	// We jump over our saved rdi and rsi so we start with -16 as offset
	localVarOffset := -16
	// Function parameters go up in the stack, so offsets are positiv
	// We jump over our saved rbp, so our offset starts with +16
	stackParamOffset := 16

	// If we have a multi-return function, we have to assign the stack pointer manually
	// and have to account for the additional offset!
	if isMultiReturn {
		localVarOffset -= 8
	}

	intStackVars, floatStackVars := getParametersOnStack(parameters, isMultiReturn)

	// This includes function parameters!
	allVariables := b.extractAllFunctionVariables(true)

	for _, v := range allVariables {
		if varOnStack(v, intStackVars, floatStackVars) {
			v.symbolTable.setVarAsmOffset(v.name, stackParamOffset)
			stackParamOffset += 8
		} else {
			v.symbolTable.setVarAsmOffset(v.name, localVarOffset)
			localVarOffset -= 8
		}
	}
	// +16 because the initial -16 do not count as we calculate the offset for local variables with rsp not rbp.
	// We negate the resulting number to return a positive amount of space we need, not the actual stack-offset!
	return -(localVarOffset + 16)
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

	localVarOffset := f.block.assignVariableStackOffset(f.parameters, len(f.returnTypes) > 1)

	// Prologue:
	addFunctionPrologue(asm, localVarOffset)

	intRegisters := getFunctionRegisters(TYPE_INT)
	intRegIndex := 0
	floatRegisters := getFunctionRegisters(TYPE_FLOAT)
	floatRegIndex := 0

	// Save stack pointer for return values
	if len(f.returnTypes) > 1 {
		asm.addLine("mov", fmt.Sprintf("[rbp%v], %v", -16, intRegisters[intRegIndex]))
		intRegIndex++
		s.setFunReturnStackPointer(f.fName, -16)
	}

	// Create local variables for all function parameters
	for _, v := range f.parameters {

		entry, _ := f.block.symbolTable.getVar(v.vName)
		sign := "+"
		if entry.offset < 0 {
			sign = ""
		}
		switch v.vType.t {
		case TYPE_INT, TYPE_BOOL, TYPE_ARRAY:
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

	switch len(entry.returnTypes) {
	case 0:
		// Explicitely do nothing :)
	case 1:
		// Special case here: If there is just one return, there is no need to push/pop and handle
		// everything. It is already in rax/xmm0, so we are good!
		r.expressions[0].generateCode(asm, s)
	default:
		for i := len(r.expressions) - 1; i >= 0; i-- {
			e := r.expressions[i]

			e.generateCode(asm, s)

			// Make sure everything is actually on the stack!
			if e.getResultCount() == 1 {
				switch e.getExpressionTypes()[0].t {
				case TYPE_INT:
					asm.addLine("push", getReturnRegister(TYPE_INT))
				case TYPE_FLOAT:
					tmpR, _ := getRegister(TYPE_INT)
					asm.addLine("movq", fmt.Sprintf("%v, %v", tmpR, getReturnRegister(TYPE_FLOAT)))
					asm.addLine("push", tmpR)
				}
			}
		}

		// For multiple expressions/return values, they will assemble on the stack in reverse order.
		// So we have to pop them and move them onto the pre-allocated stack space!

		tmpR, stackPointerReg := getRegister(TYPE_INT)
		asm.addLine("mov", fmt.Sprintf("%v, [rbp%v]", stackPointerReg, entry.returnStackPointerOffset))

		offset := 0
		for _ = range entry.returnTypes {
			asm.addLine("pop", tmpR)
			asm.addLine("mov", fmt.Sprintf("[%v+%v], %v", stackPointerReg, offset, tmpR))
			offset += 8
		}
	}

	asm.addLine("jmp", entry.epilogueLabel)
}

func (ast AST) addPrintCharFunction(asm *ASM) {

	asmName := asm.nextFunctionName()
	ast.globalSymbolTable.setFunAsmName("printChar", asmName)

	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asm.addFun(asmName)

	asm.addLine("push", "rsi")
	asm.addLine("push", "rdi")

	// Message length
	asm.addLine("mov", "rdx, 1")
	// We create a pointer to the message given in rdi, by pushing to the stack and referencing the value

	asm.addLine("lea", "rsi, [rsp]")
	// File handle
	asm.addLine("mov", "rdi, 1")
	// System call number
	asm.addLine("mov", "rax, 1")
	asm.addLine("syscall", "")

	asm.addLine("pop", "rdi")
	asm.addLine("pop", "rsi")

	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) addPrintIntFunction(asm *ASM) {

	asmName := asm.nextFunctionName()

	entry, ok := ast.globalSymbolTable.getFun("printChar")
	if !ok {
		fmt.Println("Code generation error. printChar can not be found")
	}
	printCharName := entry.jumpLabel

	ast.globalSymbolTable.setFunAsmName("printInt", asmName)

	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asm.addFun(asmName)

	asm.addLine("push", "rdi")
	asm.addLine("push", "rsi")

	continueLabel := asm.nextLabelName()

	asm.addLine("mov", "rax, rdi")
	// rsi with the number of digits to print
	asm.addLine("xor", "rsi, rsi")

	// Handle negative values
	asm.addLine("cmp", "rax, 0")
	asm.addLine("jge", continueLabel)
	// "-"
	asm.addLine("mov", "rdi, 0x2D")
	// Save rax before calling
	asm.addLine("push", "rax")
	asm.addLine("call", printCharName)
	asm.addLine("pop", "rax")
	asm.addLabel(continueLabel)

	// Push digits to stack as printable chars
	loop1Label := asm.nextLabelName()
	asm.addLabel(loop1Label)
	asm.addLine("xor", "rdx, rdx")
	asm.addLine("inc", "rsi")
	asm.addLine("mov", "r10, 10")
	asm.addLine("idiv", "rax, r10")
	// Make it an ascii printable char
	asm.addLine("add", "rdx, 0x30")
	asm.addLine("push", "rdx")
	asm.addLine("cmp", "rax, 0")
	asm.addLine("jne", loop1Label)

	// Print all characters we pushed to the stack
	checkLabel := asm.nextLabelName()
	loop2Label := asm.nextLabelName()
	asm.addLine("jmp", checkLabel)
	asm.addLabel(loop2Label)
	asm.addLine("pop", "rdi")
	asm.addLine("call", printCharName)
	asm.addLine("dec", "rsi")
	asm.addLabel(checkLabel)
	asm.addLine("cmp", "rsi, 0")
	asm.addLine("jne", loop2Label)

	asm.addLine("pop", "rsi")
	asm.addLine("pop", "rdi")

	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) addPrintIntLnFunction(asm *ASM) {

	asmName := asm.nextFunctionName()
	ast.globalSymbolTable.setFunAsmName("printIntLn", asmName)

	entry, ok := ast.globalSymbolTable.getFun("printChar")
	if !ok {
		fmt.Println("Code generation error. printChar can not be found")
	}
	printCharName := entry.jumpLabel
	entry, ok = ast.globalSymbolTable.getFun("printInt")
	if !ok {
		fmt.Println("Code generation error. printInt can not be found")
	}
	printIntName := entry.jumpLabel

	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asm.addFun(asmName)
	//addFunctionPrologue(asm, 0)

	asm.addLine("push", "rdi")

	asm.addLine("call", printIntName)
	// Print newline at the end
	asm.addLine("mov", "rdi, 0xA")
	asm.addLine("call", printCharName)

	asm.addLine("pop", "rdi")

	//addFunctionEpilogue(asm)
	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) addPrintFloatFunction(asm *ASM) {

	asmName := asm.nextFunctionName()
	ast.globalSymbolTable.setFunAsmName("printFloat", asmName)

	entry, ok := ast.globalSymbolTable.getFun("printChar")
	if !ok {
		fmt.Println("Code generation error. printChar can not be found")
	}
	printCharName := entry.jumpLabel
	entry, ok = ast.globalSymbolTable.getFun("printInt")
	if !ok {
		fmt.Println("Code generation error. printInt can not be found")
	}
	printIntName := entry.jumpLabel

	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asm.addFun(asmName)
	//addFunctionPrologue(asm, 0)

	asm.addLine("push", "rdi")

	okLabel := asm.nextLabelName()
	// Handle negative Floats
	asm.addLine("movq", "r10, xmm0")
	asm.addLine("cmp", "r10, 0")
	asm.addLine("jge", okLabel)

	asm.addLine("mov", "rdi, 0x2D")
	asm.addLine("call", printCharName)
	asm.addLine("mulsd", "xmm0, qword [negOneF]")

	asm.addLabel(okLabel)
	// Integer part
	asm.addLine("cvttsd2si", "rdi, xmm0")
	asm.addLine("push", "rdi")
	asm.addLine("call", printIntName)
	// Decimal .
	asm.addLine("mov", "rdi, 0x2E")
	asm.addLine("call", printCharName)
	asm.addLine("pop", "r10")
	asm.addLine("cvtsi2sd", "xmm1, r10")
	// Digits after decimal point
	asm.addLine("subsd", "xmm0, xmm1")
	// Number of digits to print (power of 10)
	asm.addLine("mov", "r10, 1000.0")
	asm.addLine("movq", "xmm1, r10")
	asm.addLine("mulsd", "xmm0, xmm1")
	asm.addLine("cvttsd2si", "rdi, xmm0")
	asm.addLine("call", printIntName)

	// Print newline at the end
	//asm.addLine("mov", "rdi, 0xA")
	//asm.addLine("call", "printChar")

	asm.addLine("pop", "rdi")

	//addFunctionEpilogue(asm)
	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) addPrintFloatLnFunction(asm *ASM) {

	asmName := asm.nextFunctionName()
	ast.globalSymbolTable.setFunAsmName("printFloatLn", asmName)

	entry, ok := ast.globalSymbolTable.getFun("printChar")
	if !ok {
		fmt.Println("Code generation error. printChar can not be found")
	}
	printCharName := entry.jumpLabel
	entry, ok = ast.globalSymbolTable.getFun("printFloat")
	if !ok {
		fmt.Println("Code generation error. printFloat can not be found")
	}
	printFloatName := entry.jumpLabel

	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asm.addFun(asmName)
	//addFunctionPrologue(asm, 0)

	asm.addLine("push", "rdi")

	asm.addLine("call", printFloatName)
	// Print newline at the end
	asm.addLine("mov", "rdi, 0xA")
	asm.addLine("call", printCharName)

	asm.addLine("pop", "rdi")

	//addFunctionEpilogue(asm)
	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) addArrayLenFunction(asm *ASM) {
	asmName := asm.nextFunctionName()
	ast.globalSymbolTable.setFunAsmName("len", asmName)

	savedProgram := asm.program
	asm.program = make([][3]string, 0)

	asm.addFun(asmName)
	// Reads the actual first entry manually, which is the length of the following array
	asm.addLine("mov", "rax, [rdi]")
	asm.addLine("ret", "")

	for _, line := range asm.program {
		asm.functions = append(asm.functions, line)
	}
	asm.program = savedProgram
}

func (ast AST) generateCode() ASM {

	asm := ASM{}
	asm.constants = make(map[string]string, 0)

	asm.header = append(asm.header, "extern printf  ; C function we need for debugging")
	asm.header = append(asm.header, "section .data")

	asm.variables = append(asm.variables, [3]string{"fmti", "db", "\"%i\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"fmtf", "db", "\"%.2f\", 10, 0"})
	asm.variables = append(asm.variables, [3]string{"negOneF", "dq", "-1.0"})

	asm.sectionText = append(asm.sectionText, "section .text")

	ast.addPrintCharFunction(&asm)

	ast.addPrintIntFunction(&asm)
	ast.addPrintIntLnFunction(&asm)

	ast.addPrintFloatFunction(&asm)
	ast.addPrintFloatLnFunction(&asm)

	ast.addArrayLenFunction(&asm)

	asm.addFun("_start")

	stackOffset := ast.block.assignVariableStackOffset([]Variable{}, false)
	// To make the stack later 16bit aligned. Not really sure if thats the issue, to be honest...
	if stackOffset == 0 {
		stackOffset = 8
	}

	addFunctionPrologue(&asm, stackOffset)

	ast.block.generateCode(&asm, &ast.globalSymbolTable)

	addFunctionEpilogue(&asm)

	asm.addLine("; Exit the program nicely", "")
	asm.addLabel("exit")
	asm.addLine("mov", "rax, 60  ; sys_exit")
	asm.addLine("mov", "rdi, 0   ; exit code")
	asm.addLine("syscall", "")

	return asm
}
