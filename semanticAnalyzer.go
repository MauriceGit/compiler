package main

import (
	"fmt"
)

// getVar goes through all symbol varTables recursively and looks for an entry for the given variable name v
func (s *SymbolTable) getVar(v string) (SymbolVarEntry, bool) {
	if s == nil {
		return SymbolVarEntry{}, false
	}
	if variable, ok := s.varTable[v]; ok {
		return variable, true
	}
	return s.parent.getVar(v)
}

// isLocalVar only searches the immediate local symbol varTable
func (s *SymbolTable) isLocalVar(v string) bool {
	if s == nil {
		return false
	}
	_, ok := s.varTable[v]
	return ok
}

func (s *SymbolTable) getLocalVars() (keys []string) {
	for k := range s.varTable {
		keys = append(keys, k)
	}
	return
}

func (s *SymbolTable) setVar(v string, t ComplexType, isIndexed bool) {
	s.varTable[v] = SymbolVarEntry{t, "", 0, isIndexed}
}

//func (s *SymbolTable) setIndexedVar(v string, t Type) {
//	s.varTable[v] = SymbolVarEntry{TYPE_ARRAY, "", 0, true, t}
//}

func (s *SymbolTable) setFun(name string, argTypes, returnTypes []ComplexType) {
	s.funTable[name] = SymbolFunEntry{argTypes, returnTypes, "", "", 0}
}

func (s *SymbolTable) isLocalFun(name string) bool {
	if s == nil {
		return false
	}
	_, ok := s.funTable[name]
	return ok
}

func (s *SymbolTable) getFun(name string) (SymbolFunEntry, bool) {
	if s == nil {
		return SymbolFunEntry{}, false
	}
	if fun, ok := s.funTable[name]; ok {
		return fun, true
	}
	return s.parent.getFun(name)
}

func (s *SymbolTable) setVarAsmName(v string, asmName string) {
	if s == nil {
		panic("Could not set asm variable name in symbol table!")
		return
	}
	if _, ok := s.varTable[v]; ok {
		tmp := s.varTable[v]
		tmp.varName = asmName
		s.varTable[v] = tmp
		return
	}
	s.parent.setVarAsmName(v, asmName)
}

func (s *SymbolTable) setVarAsmOffset(v string, offset int) {
	if s == nil {
		panic("Could not set asm variable name in symbol table!")
		return
	}
	if _, ok := s.varTable[v]; ok {
		tmp := s.varTable[v]
		tmp.offset = offset
		s.varTable[v] = tmp
		return
	}
	s.parent.setVarAsmOffset(v, offset)
}

func (s *SymbolTable) setFunAsmName(v string, asmName string) {
	if s == nil {
		panic("Could not set asm function name in symbol table!")
		return
	}
	if _, ok := s.funTable[v]; ok {
		tmp := s.funTable[v]
		tmp.jumpLabel = asmName
		s.funTable[v] = tmp
		return
	}
	s.parent.setFunAsmName(v, asmName)
}

func (s *SymbolTable) setFunEpilogueLabel(v string, label string) {
	if s == nil {
		panic("Could not set asm epilogue label in symbol table!")
		return
	}
	if _, ok := s.funTable[v]; ok {
		tmp := s.funTable[v]
		tmp.epilogueLabel = label
		s.funTable[v] = tmp
		return
	}
	s.parent.setFunEpilogueLabel(v, label)
}

func (s *SymbolTable) setFunReturnStackPointer(v string, offset int) {
	if s == nil {
		panic("Could not set asm return stack pointer in symbol table!")
		return
	}
	if _, ok := s.funTable[v]; ok {
		tmp := s.funTable[v]
		tmp.returnStackPointerOffset = offset
		s.funTable[v] = tmp
		return
	}
	s.parent.setFunReturnStackPointer(v, offset)
}

func analyzeUnaryOp(unaryOp UnaryOp, symbolTable *SymbolTable) (Expression, error) {

	// Re-order expression, if the expression is not fixed and the priority is of the operator is not according to the priority
	// The priority of an operator must be equal or higher in (right) sub-trees (as they are evaluated first).
	if tmpE, ok := unaryOp.expr.(BinaryOp); ok {
		if unaryOp.operator.priority() < tmpE.operator.priority() && !tmpE.fixed {

			newChild := unaryOp
			newChild.expr = tmpE.leftExpr
			tmpE.leftExpr = newChild

			var newRoot Expression
			newRoot = tmpE

			// The unary expression is now a binary, so we have to start over. The following checks won't be working any more.
			expression, err := analyzeExpression(newRoot, symbolTable)
			if err != nil {
				return newRoot, err
			}
			newRoot = expression
			return newRoot, nil
		}
	}

	expression, err := analyzeExpression(unaryOp.expr, symbolTable)
	if err != nil {
		return unaryOp, err
	}
	unaryOp.expr = expression

	if expression.getResultCount() != 1 {
		return nil, fmt.Errorf("%w[%v:%v] - Unary expression can only handle one result", ErrCritical, unaryOp.line, unaryOp.column)
	}
	t := expression.getExpressionTypes()[0]

	switch unaryOp.operator {
	case OP_NEGATIVE:
		if t.t != TYPE_FLOAT && t.t != TYPE_INT {
			return nil, fmt.Errorf("%w[%v:%v] - Unary '-' expression must be float or int, but is: %v", ErrCritical, unaryOp.line, unaryOp.column, unaryOp)
		}
		unaryOp.opType = t
		return unaryOp, nil
	case OP_NOT:
		if t.t != TYPE_BOOL {
			return nil, fmt.Errorf("%w[%v:%v] - Unary '!' expression must be bool, but is: %v", ErrCritical, unaryOp.line, unaryOp.column, unaryOp)
		}
		unaryOp.opType = ComplexType{TYPE_BOOL, nil}
		return unaryOp, nil
	}
	return nil, fmt.Errorf("%w[%v:%v] - Unknown unary expression: %v", ErrCritical, unaryOp.line, unaryOp.column, unaryOp)
}

func analyzeBinaryOp(binaryOp BinaryOp, symbolTable *SymbolTable) (Expression, error) {

	// Re-order expression, if the expression is not fixed and the priority is of the operator is not according to the priority
	// The priority of an operator must be equal or higher in (right) sub-trees (as they are evaluated first).
	if tmpE, ok := binaryOp.rightExpr.(BinaryOp); ok {
		if binaryOp.operator.priority() < tmpE.operator.priority() && !tmpE.fixed {
			newChild := binaryOp
			newChild.rightExpr = tmpE.leftExpr
			tmpE.leftExpr = newChild
			binaryOp = tmpE
		}
	}

	leftExpression, err := analyzeExpression(binaryOp.leftExpr, symbolTable)
	if err != nil {
		return binaryOp, err
	}
	binaryOp.leftExpr = leftExpression

	rightExpression, err := analyzeExpression(binaryOp.rightExpr, symbolTable)
	if err != nil {
		return binaryOp, err
	}
	binaryOp.rightExpr = rightExpression

	if binaryOp.leftExpr.getResultCount() != binaryOp.rightExpr.getResultCount() || binaryOp.leftExpr.getResultCount() != 1 {
		return nil, fmt.Errorf("%w[%v:%v] - BinaryOp %v expected two values, got %v and %v",
			ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator,
			binaryOp.leftExpr.getResultCount(), binaryOp.rightExpr.getResultCount(),
		)
	}

	tLeft := binaryOp.leftExpr.getExpressionTypes()[0]
	tRight := binaryOp.rightExpr.getExpressionTypes()[0]

	// Check types only after we possibly rearranged the expression!
	if binaryOp.leftExpr.getExpressionTypes()[0] != binaryOp.rightExpr.getExpressionTypes()[0] {
		return binaryOp, fmt.Errorf(
			"%w[%v:%v] - BinaryOp '%v' expected same type, got: '%v', '%v'",
			ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft, tRight,
		)
	}

	// We match all types explicitely to make sure that this still works or creates an error when we introduce new types
	// that are not considered yet!
	switch binaryOp.operator {
	case OP_AND, OP_OR:
		binaryOp.opType = ComplexType{TYPE_BOOL, nil}
		// We know left and right are the same type, so only compare left here.
		if tLeft.t != TYPE_BOOL {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp '%v' needs bool, got: '%v'",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, TYPE_BOOL, nil
	case OP_MOD:
		binaryOp.opType = ComplexType{TYPE_INT, nil}
		if tLeft.t != TYPE_INT {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp '%v' only works for int",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator,
			)
		}
	case OP_PLUS, OP_MINUS, OP_MULT, OP_DIV:

		if tLeft.t == TYPE_FLOAT {
			binaryOp.opType = ComplexType{TYPE_FLOAT, nil}
		} else {
			binaryOp.opType = ComplexType{TYPE_INT, nil}
		}
		if tLeft.t != TYPE_FLOAT && tLeft.t != TYPE_INT {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp '%v' needs int/float, got: '%v'",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, tLeft, nil
	case OP_LE, OP_GE, OP_LESS, OP_GREATER:
		binaryOp.opType.t = TYPE_BOOL
		if tLeft.t != TYPE_FLOAT && tLeft.t != TYPE_INT && tLeft.t != TYPE_STRING {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp '%v' needs int/float/string, got: '%v'",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, TYPE_BOOL, nil
	case OP_EQ, OP_NE:
		binaryOp.opType = ComplexType{TYPE_BOOL, nil}
		// We can actually compare all data types. So there will be no missmatch in general!
	default:
		return binaryOp, fmt.Errorf(
			"%w[%v:%v] - Invalid binary operator: '%v' for type '%v'",
			ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
		)
	}

	return binaryOp, nil
}

func analyzeFunCall(fun FunCall, symbolTable *SymbolTable) (FunCall, error) {

	funEntry, ok := symbolTable.getFun(fun.funName)
	if !ok {
		return fun, fmt.Errorf("%w[%v:%v] - Function '%v' called before declaration", ErrCritical, fun.line, fun.column, fun.funName)
	}
	fun.retTypes = funEntry.returnTypes

	// Basically unpacking expression list before providing it to the function
	expressionTypes := []ComplexType{}
	for i, e := range fun.args {
		newE, parseErr := analyzeExpression(e, symbolTable)
		if parseErr != nil {
			return fun, parseErr
		}
		fun.args[i] = newE

		expressionTypes = append(expressionTypes, newE.getExpressionTypes()...)
	}

	if len(expressionTypes) != len(funEntry.paramTypes) {
		return fun, fmt.Errorf("%w[%v:%v] - Function call to '%v' has %v parameters, but needs %v",
			ErrCritical, fun.line, fun.column, fun.funName, len(expressionTypes), len(funEntry.paramTypes),
		)
	}

	for i, t := range expressionTypes {
		if !equalType(t, funEntry.paramTypes[i]) {
			return fun, fmt.Errorf("%w[%v:%v] - Function call to '%v' got type %v as %v. parameter, but needs %v",
				ErrCritical, fun.line, fun.column, fun.funName, t, i+1, funEntry.paramTypes[i],
			)
		}
	}

	return fun, nil
}

func analyzeArrayDecl(a Array, symbolTable *SymbolTable) (Array, error) {

	var arrayType ComplexType = ComplexType{TYPE_UNKNOWN, nil}
	arraySize := 0
	for i, e := range a.aExpressions {
		newE, err := analyzeExpression(e, symbolTable)
		if err != nil {
			return a, err
		}
		a.aExpressions[i] = newE
		if i == 0 {
			arrayType = newE.getExpressionTypes()[0]
		}
		if newE.getExpressionTypes()[0] != arrayType {
			return a, fmt.Errorf("%w[%v:%v] - Not all expressions in array declaration have the same type",
				ErrCritical, a.line, a.column,
			)
		}
		arraySize += newE.getResultCount()
	}

	// If aCount is already set to something, we know, this value is correct!!!
	if a.aCount == 0 && arraySize > 0 {
		a.aCount = arraySize
	}

	// For an empty array declaration, the type is set explicitely!
	if a.aType.t == TYPE_UNKNOWN {
		a.aType = arrayType
	}

	return a, nil
}

func analyzeVariable(e Variable, symbolTable *SymbolTable) (Variable, error) {
	// Lookup variable type and annotate node.
	if vTable, ok := symbolTable.getVar(e.vName); ok {
		e.vType = vTable.sType

		// Type is not array, if it is indexed, but the type of the element itself.
		if e.vType.t == TYPE_ARRAY && e.vIsIndexedArray {

			indexExpression, tmpE := analyzeExpression(e.vIndexExpression, symbolTable)
			if tmpE != nil {
				return e, tmpE
			}

			if len(indexExpression.getExpressionTypes()) != 1 {
				return e, fmt.Errorf("%w[%v:%v] - Index expression can only have one value", ErrCritical, e.line, e.column)
			}
			if indexExpression.getExpressionTypes()[0].t != TYPE_INT {
				return e, fmt.Errorf("%w[%v:%v] - Index expression must be int", ErrCritical, e.line, e.column)
			}

			e.vIndexExpression = indexExpression

			// We already set the vType, which includes references to array types.
			//			e.vArrayType = vTable.arrayType
			//			e.vType = vTable.arrayType

		}

	} else {
		return e, fmt.Errorf("%w[%v:%v] - Variable '%v' referenced before declaration", ErrCritical, e.line, e.column, e.vName)
	}
	// Always access the very last entry for variables!
	return e, nil
}

func analyzeExpression(expression Expression, symbolTable *SymbolTable) (Expression, error) {

	switch e := expression.(type) {
	case Constant:
		return e, nil
	case Variable:
		return analyzeVariable(e, symbolTable)
	case UnaryOp:
		return analyzeUnaryOp(e, symbolTable)
	case BinaryOp:
		return analyzeBinaryOp(e, symbolTable)
	case FunCall:
		return analyzeFunCall(e, symbolTable)
	case Array:
		return analyzeArrayDecl(e, symbolTable)
	}
	row, col := expression.startPos()
	return expression, fmt.Errorf("%w[%v:%v] - Unknown type for expression '%v'", ErrCritical, row, col, expression)
}

// Returns newly created variables and variables that should shadow others!
// This is just for housekeeping and removing them later!!!!
// All new variables (and shadow ones) are updated/written to the symbol varTable
func analyzeAssignment(assignment Assignment, symbolTable *SymbolTable) (Assignment, error) {

	expressionTypes := make([]ComplexType, 0)
	// We need this temporary array to hold information about sub-types for arrays, so we can
	// pass this information on for later usage in indexed variables!
	for i, e := range assignment.expressions {
		expression, err := analyzeExpression(e, symbolTable)
		if err != nil {
			return assignment, err
		}

		tmpTypes := expression.getExpressionTypes()
		expressionTypes = append(expressionTypes, tmpTypes...)
		assignment.expressions[i] = expression

	}

	// Populate/overwrite the dictionary of variables for futher statements :)
	if len(assignment.variables) != len(expressionTypes) {

		row, col := assignment.startPos()
		if len(assignment.variables) > 0 {
			row, col = assignment.variables[0].line, assignment.variables[0].column
		}

		return assignment, fmt.Errorf(
			"%w[%v:%v] - Variables and expression count need to match", ErrCritical, row, col,
		)
	}

	for i, v := range assignment.variables {

		expressionType := expressionTypes[i]

		// Shadowing is only allowed in a different block, not right after the first variable, to avoid confusion and complicated
		// variable handling
		if symbolTable.isLocalVar(v.vName) && v.vShadow {
			return assignment, fmt.Errorf(
				"%w[%v:%v] - Variable %v is shadowing another variable in the same block. This is not allowed",
				ErrCritical, v.line, v.column, v.vName,
			)
		}

		// Only, if the variable already exists and we're not trying to shadow it!
		if vTable, ok := symbolTable.getVar(v.vName); ok {
			if !v.vShadow {

				var variableType ComplexType = vTable.sType
				if v.vIsIndexedArray {
					variableType = *vTable.sType.subType
				}

				if !equalType(variableType, expressionType) {
					return assignment, fmt.Errorf(
						"%w[%v:%v] - Assignment type missmatch between variable %v and expression %v",
						ErrCritical, v.line, v.column, v, expressionType,
					)
				}
			} else {
				if v.vIsIndexedArray {
					return assignment, fmt.Errorf("%w[%v:%v] - An indexed array write can not shadow its source",
						ErrCritical, v.line, v.column,
					)
				}

				symbolTable.setVar(v.vName, expressionType, false)
			}
		} else {
			symbolTable.setVar(v.vName, expressionType, v.vIsIndexedArray)
		}

		assignment.variables[i].vType = expressionType
	}
	return assignment, nil
}

func analyzeCondition(condition Condition, symbolTable *SymbolTable) (Condition, error) {

	// This expression MUST come out as boolean!
	e, err := analyzeExpression(condition.expression, symbolTable)
	if err != nil {
		return condition, err
	}

	if e.getResultCount() != 1 {
		row, col := e.startPos()
		return condition, fmt.Errorf("%w[%v:%v] - Condition accepts only one expression, got %v",
			ErrCritical, row, col, e.getResultCount(),
		)
	}
	t := e.getExpressionTypes()[0]
	if t.t != TYPE_BOOL {
		row, col := e.startPos()
		return condition, fmt.Errorf(
			"%w[%v:%v] - If expression expected boolean, got: %v --> <<%v>>",
			ErrCritical, row, col, t, condition.expression,
		)
	}
	condition.expression = e

	block, err := analyzeBlock(condition.block, symbolTable, nil)
	if err != nil {
		return condition, err
	}
	condition.block = block

	elseBlock, err := analyzeBlock(condition.elseBlock, symbolTable, nil)
	if err != nil {
		return condition, err
	}
	condition.elseBlock = elseBlock

	return condition, nil
}

func analyzeLoop(loop Loop, symbolTable *SymbolTable) (Loop, error) {

	nextSymbolTable := SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string]SymbolFunEntry, 0),
		symbolTable.activeFunctionName,
		symbolTable.activeFunctionReturn,
		symbolTable,
	}

	assignment, err := analyzeAssignment(loop.assignment, &nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.assignment = assignment

	for i, e := range loop.expressions {
		expression, err := analyzeExpression(e, &nextSymbolTable)
		if err != nil {
			return loop, err
		}

		for _, t := range expression.getExpressionTypes() {
			if t.t != TYPE_BOOL {
				row, col := expression.startPos()
				return loop, fmt.Errorf(
					"%w[%v:%v] - Loop expression expected boolean, got: %v (%v)",
					ErrCritical, row, col, t, expression,
				)
			}
		}

		loop.expressions[i] = expression
	}

	incrAssignment, err := analyzeAssignment(loop.incrAssignment, &nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.incrAssignment = incrAssignment

	statements, err := analyzeBlock(loop.block, symbolTable, &nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.block = statements
	loop.block.symbolTable = nextSymbolTable

	return loop, nil
}

func analyzeFunction(fun Function, symbolTable *SymbolTable) (Function, error) {

	functionSymbolTable := SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string]SymbolFunEntry, 0),
		fun.fName,
		fun.returnTypes,
		symbolTable,
	}

	for _, v := range fun.parameters {
		if v.vType.t == TYPE_UNKNOWN {
			return fun, fmt.Errorf("%w[%v:%v] - Function parameter %v has invalid type", ErrCritical, v.line, v.column, v)
		}
		if _, ok := functionSymbolTable.getVar(v.vName); ok {
			return fun, fmt.Errorf("%w[%v:%v] - Function parameter %v already exists", ErrCritical, v.line, v.column, v)
		}
		functionSymbolTable.setVar(v.vName, v.vType, false)
	}

	if symbolTable.isLocalFun(fun.fName) {
		return fun, fmt.Errorf("%w[%v:%v] - Function with the same name already exists in this scope", ErrCritical, fun.line, fun.column)
	}
	var paramTypes []ComplexType
	for _, v := range fun.parameters {
		paramTypes = append(paramTypes, v.vType)
	}

	symbolTable.setFun(fun.fName, paramTypes, fun.returnTypes)

	newBlock, err := analyzeBlock(fun.block, symbolTable, &functionSymbolTable)
	if err != nil {
		return fun, err
	}
	fun.block = newBlock

	// Checks, that every single path has one return statement so we don't have undefined function returns,
	// unassigned registers or no expected values on stack
	if err = fun.block.functionReturnAnalysis(); err != nil {
		return fun, fmt.Errorf("%w[%v:%v] - %v", ErrCritical, fun.line, fun.column, err.Error())
	}

	return fun, nil
}

func analyzeReturn(ret Return, symbolTable *SymbolTable) (Return, error) {

	typeIndex := 0
	for i, e := range ret.expressions {

		newE, err := analyzeExpression(e, symbolTable)
		if err != nil {
			return ret, err
		}
		row, col := e.startPos()
		for _, t := range newE.getExpressionTypes() {

			if typeIndex >= len(symbolTable.activeFunctionReturn) {
				return ret, fmt.Errorf("%w[%v:%v] - Too many expressions returned. Expected %v",
					ErrCritical, row, col, len(symbolTable.activeFunctionReturn),
				)
			}

			if t != symbolTable.activeFunctionReturn[typeIndex] {
				return ret, fmt.Errorf("%w[%v:%v] - Function return type does not match definition. Expected %v, got %v",
					ErrCritical, row, col, symbolTable.activeFunctionReturn[typeIndex], t)
			}
			typeIndex++
		}
		ret.expressions[i] = newE
	}
	return ret, nil
}

func analyzeStatement(statement Statement, symbolTable *SymbolTable) (Statement, error) {
	switch st := statement.(type) {
	case Condition:
		return analyzeCondition(st, symbolTable)
	case Loop:
		return analyzeLoop(st, symbolTable)
	case Assignment:
		return analyzeAssignment(st, symbolTable)
	case Function:
		return analyzeFunction(st, symbolTable)
	case Return:
		return analyzeReturn(st, symbolTable)
	case FunCall:
		return analyzeFunCall(st, symbolTable)
	}
	row, col := statement.startPos()
	return statement, fmt.Errorf("%w[%v:%v] - Unexpected statement: %v", ErrCritical, row, col, statement)
}

// analyzeBlock gets a reference to the current (now parent) symbol varTable
// Additionally, it might get a pre-filled symbol varTable for the new scope to use!
// This might be the case for function arguments or in a for-loop, where variables belong to the
// coming block only but are parsed in the TreeNode before.
func analyzeBlock(block Block, symbolTable, newBlockSymbolTable *SymbolTable) (Block, error) {

	if newBlockSymbolTable != nil {
		block.symbolTable = *newBlockSymbolTable
	} else {
		block.symbolTable = SymbolTable{
			make(map[string]SymbolVarEntry, 0),
			make(map[string]SymbolFunEntry, 0),
			symbolTable.activeFunctionName,
			symbolTable.activeFunctionReturn,
			symbolTable,
		}
	}

	for i, s := range block.statements {
		statement, err := analyzeStatement(s, &block.symbolTable)
		if err != nil {
			return block, err
		}
		block.statements[i] = statement
	}

	return block, nil
}

// analyzeTypes traverses the tree and analyzes variables with their corresponding type recursively from expressions!
// returns an error if we have a type missmatch anywhere!
func semanticAnalysis(ast AST) (AST, error) {

	ast.globalSymbolTable = SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string]SymbolFunEntry, 0),
		"",
		nil,
		nil,
	}

	ast.globalSymbolTable.setFun("printInt", []ComplexType{ComplexType{TYPE_INT, nil}}, []ComplexType{ComplexType{TYPE_INT, nil}})
	ast.globalSymbolTable.setFun("printFloat", []ComplexType{ComplexType{TYPE_FLOAT, nil}}, []ComplexType{ComplexType{TYPE_INT, nil}})

	block, err := analyzeBlock(ast.block, &ast.globalSymbolTable, nil)
	if err != nil {
		ast.globalSymbolTable = SymbolTable{}
		return ast, err
	}
	ast.block = block

	return ast, nil
}
