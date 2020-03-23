package main

import (
	"fmt"
)

// get goes through all symbol tables recursively and looks for an entry for the given variable name v
func (s *SymbolTable) get(v string) (SymbolEntry, bool) {
	if s == nil {
		return SymbolEntry{}, false
	}
	if variable, ok := s.table[v]; ok {
		return variable, true
	}
	return s.parent.get(v)
}

// getLocal only searches the immediate local symbol table
func (s *SymbolTable) getLocal(v string) (SymbolEntry, bool) {
	if s == nil {
		return SymbolEntry{}, false
	}
	se, ok := s.table[v]
	return se, ok
}

func (s *SymbolTable) set(v string, t Type) {
	s.table[v] = SymbolEntry{t, ""}
}

func (s *SymbolTable) setAsmName(v string, asmName string) {
	if s == nil {
		fmt.Println("Could not set asm variable name in symbol table!")
		return
	}
	if _, ok := s.table[v]; ok {

		tmp := s.table[v]
		tmp.varName = asmName
		s.table[v] = tmp
		return
	}
	s.parent.setAsmName(v, asmName)
}

func analyzeTypeUnaryOp(unaryOp UnaryOp, symbolTable *SymbolTable) (Expression, error) {
	expression, err := analyzeTypeExpression(unaryOp.expr, symbolTable)
	if err != nil {
		return unaryOp, err
	}
	unaryOp.expr = expression

	t := expression.getExpressionType()

	switch unaryOp.operator {
	case OP_NEGATIVE:
		if t != TYPE_FLOAT && t != TYPE_INT {
			return nil, fmt.Errorf("%w[%v:%v] - Unary '-' expression must be float or int, but is: %v", ErrCritical, unaryOp.line, unaryOp.column, unaryOp)
		}
		unaryOp.opType = expression.getExpressionType()
		return unaryOp, nil
	case OP_NOT:
		if t != TYPE_BOOL {
			return nil, fmt.Errorf("%w[%v:%v] - Unary '!' expression must be bool, but is: %v", ErrCritical, unaryOp.line, unaryOp.column, unaryOp)
		}
		unaryOp.opType = TYPE_BOOL
		return unaryOp, nil
	}
	return nil, fmt.Errorf("%w[%v:%v] - Unknown unary expression: %v", ErrCritical, unaryOp.line, unaryOp.column, unaryOp)
}

func analyzeTypeBinaryOp(binaryOp BinaryOp, symbolTable *SymbolTable) (Expression, error) {

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

	leftExpression, err := analyzeTypeExpression(binaryOp.leftExpr, symbolTable)
	if err != nil {
		return binaryOp, err
	}
	binaryOp.leftExpr = leftExpression

	rightExpression, err := analyzeTypeExpression(binaryOp.rightExpr, symbolTable)
	if err != nil {
		return binaryOp, err
	}
	binaryOp.rightExpr = rightExpression

	tLeft := binaryOp.leftExpr.getExpressionType()
	tRight := binaryOp.rightExpr.getExpressionType()

	// Check types only after we possibly rearranged the expression!
	if binaryOp.leftExpr.getExpressionType() != binaryOp.rightExpr.getExpressionType() {
		fmt.Println(binaryOp.leftExpr, binaryOp.rightExpr)
		return binaryOp, fmt.Errorf(
			"%w[%v:%v] - BinaryOp expected same type, got: '%v' %v '%v'",
			ErrCritical, binaryOp.line, binaryOp.column, tLeft, binaryOp.operator, tRight,
		)
	}

	// We match all types explicitely to make sure that this still works or creates an error when we introduce new types
	// that are not considered yet!
	switch binaryOp.operator {
	case OP_AND, OP_OR:
		binaryOp.opType = TYPE_BOOL
		// We know left and right are the same type, so only compare left here.
		if tLeft != TYPE_BOOL {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp %v needs bool, got: %v",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, TYPE_BOOL, nil
	case OP_PLUS, OP_MINUS, OP_MULT, OP_DIV:

		if tLeft == TYPE_FLOAT {
			binaryOp.opType = TYPE_FLOAT
		} else {
			binaryOp.opType = TYPE_INT
		}
		if tLeft != TYPE_FLOAT && tLeft != TYPE_INT {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp %v needs int/float, got: %v",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, tLeft, nil
	case OP_LE, OP_GE, OP_LESS, OP_GREATER:
		binaryOp.opType = TYPE_BOOL
		if tLeft != TYPE_FLOAT && tLeft != TYPE_INT && tLeft != TYPE_STRING {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp %v needs int/float/string, got: %v",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, TYPE_BOOL, nil
	case OP_EQ, OP_NE:
		binaryOp.opType = TYPE_BOOL
		// We can actually compare all data types. So there will be no missmatch in general!
	default:
		return binaryOp, fmt.Errorf(
			"%w[%v:%v] - Invalid binary operator: %v for type %v",
			ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
		)
	}

	return binaryOp, nil
}

func analyzeTypeExpression(expression Expression, symbolTable *SymbolTable) (Expression, error) {

	switch e := expression.(type) {
	case Constant:
		return e, nil
	case Variable:

		// Lookup variable type and annotate node.
		if vTable, ok := symbolTable.get(e.vName); ok {
			e.vType = vTable.sType
		} else {
			return e, fmt.Errorf("%w[%v:%v] - Variable type for %v unknown!", ErrCritical, e.line, e.column, e)
		}
		// Always access the very last entry for variables!
		return e, nil
	case UnaryOp:
		return analyzeTypeUnaryOp(e, symbolTable)
	case BinaryOp:
		return analyzeTypeBinaryOp(e, symbolTable)
	}
	row, col := expression.startPos()
	return expression, fmt.Errorf("%w[%v:%v] - Unknown type for expression %v", ErrCritical, row, col, expression)
}

func analyzeTypeCondition(condition Condition, symbolTable *SymbolTable) (Condition, error) {

	// This expression MUST come out as boolean!
	e, err := analyzeTypeExpression(condition.expression, symbolTable)
	if err != nil {
		return condition, err
	}
	if e.getExpressionType() != TYPE_BOOL {
		row, col := e.startPos()
		return condition, fmt.Errorf(
			"%w[%v:%v] - If expression expected boolean, got: %v --> <<%v>>",
			ErrCritical, row, col, e.getExpressionType(), condition.expression,
		)
	}
	condition.expression = e

	block, err := analyzeTypeBlock(condition.block, symbolTable, nil)
	if err != nil {
		return condition, err
	}
	condition.block = block

	elseBlock, err := analyzeTypeBlock(condition.elseBlock, symbolTable, nil)
	if err != nil {
		return condition, err
	}
	condition.elseBlock = elseBlock

	return condition, nil
}

func analyzeTypeLoop(loop Loop, symbolTable *SymbolTable) (Loop, error) {

	nextSymbolTable := SymbolTable{
		make(map[string]SymbolEntry, 0),
		symbolTable,
	}

	assignment, err := analyzeTypeAssignment(loop.assignment, &nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.assignment = assignment

	for i, e := range loop.expressions {
		expression, err := analyzeTypeExpression(e, &nextSymbolTable)
		if err != nil {
			return loop, err
		}
		if expression.getExpressionType() != TYPE_BOOL {
			row, col := expression.startPos()
			return loop, fmt.Errorf(
				"%w[%v:%v] - Loop expression expected boolean, got: %v (%v)",
				ErrCritical, row, col, expression.getExpressionType(), e,
			)
		}

		loop.expressions[i] = expression
	}

	incrAssignment, err := analyzeTypeAssignment(loop.incrAssignment, &nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.incrAssignment = incrAssignment

	statements, err := analyzeTypeBlock(loop.block, symbolTable, &nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.block = statements
	loop.block.symbolTable = nextSymbolTable

	return loop, nil
}

// Returns newly created variables and variables that should shadow others!
// This is just for housekeeping and removing them later!!!!
// All new variables (and shadow ones) are updated/written to the symbol table
func analyzeTypeAssignment(assignment Assignment, symbolTable *SymbolTable) (Assignment, error) {

	// Populate/overwrite the dictionary of variables for futher statements :)
	if len(assignment.variables) != len(assignment.expressions) {

		row, col := assignment.startPos()
		if len(assignment.variables) > 0 {
			row, col = assignment.variables[0].line, assignment.variables[0].column
		}

		return assignment, fmt.Errorf(
			"%w[%v:%v] - Assignment %v - variables and expression count need to match",
			ErrCritical, row, col, assignment,
		)
	}

	for i, v := range assignment.variables {

		expression, err := analyzeTypeExpression(assignment.expressions[i], symbolTable)
		if err != nil {
			return assignment, fmt.Errorf("Invalid expression in assignment - %w", err)
		}
		expressionType := expression.getExpressionType()

		// Shadowing is only allowed in a different block, not right after the first variable, to avoid confusion and complicated
		// variable handling
		if _, ok := symbolTable.getLocal(v.vName); ok && v.vShadow {
			return assignment, fmt.Errorf(
				"%w[%v:%v] - Variable %v is shadowing another variable in the same block. This is not allowed",
				ErrCritical, v.line, v.column, v.vName,
			)
		}

		// Only, if the variable already exists and we're not trying to shadow it!
		if vTable, ok := symbolTable.get(v.vName); ok {
			if !v.vShadow {
				if vTable.sType != expressionType {
					return assignment, fmt.Errorf(
						"%w[%v:%v] - Assignment type missmatch between variable %v and expression %v",
						ErrCritical, v.line, v.column, v, expressionType,
					)
				}
			} else {
				symbolTable.set(v.vName, expressionType)
			}
		} else {
			symbolTable.set(v.vName, expressionType)
		}

		assignment.expressions[i] = expression

		// analyze variable with type
		assignment.variables[i].vType = expressionType
	}
	return assignment, nil
}

func analyzeTypeStatement(statement Statement, symbolTable *SymbolTable) (Statement, error) {
	switch st := statement.(type) {
	case Condition:
		return analyzeTypeCondition(st, symbolTable)
	case Loop:
		return analyzeTypeLoop(st, symbolTable)
	case Assignment:
		assignment, err := analyzeTypeAssignment(st, symbolTable)
		if err != nil {
			return assignment, err
		}
		return assignment, nil
	}
	row, col := statement.startPos()
	return statement, fmt.Errorf("%w[%v:%v] - Unexpected statement: %v", ErrCritical, row, col, statement)
}

// analyzeTypeBlock gets a reference to the current (now parent) symbol table
// Additionally, it might get a pre-filled symbol table for the new scope to use!
// This might be the case for function arguments or in a for-loop, where variables belong to the
// coming block only but are parsed in the TreeNode before.
func analyzeTypeBlock(block Block, symbolTable, newBlockSymbolTable *SymbolTable) (Block, error) {

	if newBlockSymbolTable != nil {
		block.symbolTable = *newBlockSymbolTable
	} else {
		block.symbolTable = SymbolTable{
			make(map[string]SymbolEntry, 0),
			symbolTable,
		}
	}

	for i, s := range block.statements {
		statement, err := analyzeTypeStatement(s, &block.symbolTable)
		if err != nil {
			return block, err
		}
		block.statements[i] = statement
	}

	return block, nil
}

// analyzeTypes traverses the tree and analyzes variables with their corresponding type recursively from expressions!
// returns an error if we have a type missmatch anywhere!
func analyzeTypes(ast AST) (AST, error) {

	ast.globalSymbolTable = SymbolTable{
		make(map[string]SymbolEntry, 0),
		nil,
	}

	// TODO: Possibly fill global symbol table with something?
	// Right now it will stay empty just because the block we parse will create its own symbol table.

	block, err := analyzeTypeBlock(ast.block, &ast.globalSymbolTable, nil)
	if err != nil {
		ast.globalSymbolTable = SymbolTable{}
		return ast, err
	}
	ast.block = block

	return ast, nil
}
