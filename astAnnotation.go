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
	s.table[v] = SymbolEntry{t}
}

func analyzeTypeUnaryOp(unaryOp UnaryOp, symbolTable *SymbolTable) (Expression, Type, error) {
	expression, t, err := analyzeTypeExpression(unaryOp.expr, symbolTable)
	if err != nil {
		return unaryOp, t, err
	}
	unaryOp.expr = expression

	switch unaryOp.operator {
	case OP_NEGATIVE:
		if t != TYPE_FLOAT && t != TYPE_INT {
			return nil, TYPE_UNKNOWN, fmt.Errorf("%w - Unary '-' expression must be float or int, but is: %v", ErrCritical, unaryOp)
		}
		return unaryOp, t, nil
	case OP_NOT:
		if t != TYPE_BOOL {
			return nil, TYPE_UNKNOWN, fmt.Errorf("%w - Unary '!' expression must be bool, but is: %v", ErrCritical, unaryOp)
		}
		return unaryOp, t, nil
	}
	return nil, TYPE_UNKNOWN, fmt.Errorf("%w - Unknown unary expression: %v", ErrCritical, unaryOp)
}

func analyzeTypeBinaryOp(binaryOp BinaryOp, symbolTable *SymbolTable) (Expression, Type, error) {
	leftExpression, tLeft, err := analyzeTypeExpression(binaryOp.leftExpr, symbolTable)
	if err != nil {
		return binaryOp, tLeft, err
	}
	binaryOp.leftExpr = leftExpression

	rightExpression, tRight, err := analyzeTypeExpression(binaryOp.rightExpr, symbolTable)
	if err != nil {
		return binaryOp, tRight, err
	}
	binaryOp.rightExpr = rightExpression

	if tLeft != tRight {
		return binaryOp, TYPE_UNKNOWN, fmt.Errorf("%w - BinaryOp %v expected same type, got: %v != %v", ErrCritical, binaryOp, tLeft, tRight)
	}

	// We match all types explicitely to make sure that this still works or creates an error when we introduce new types
	// that are not considered yet!
	switch binaryOp.operator {
	case OP_AND, OP_OR:
		// We know left and right are the same type, so only compare left here.
		if tLeft != TYPE_BOOL {
			return binaryOp, TYPE_UNKNOWN, fmt.Errorf("%w - BinaryOp %v needs bool, got: %v", ErrCritical, binaryOp.operator, tLeft)
		}
		return binaryOp, TYPE_BOOL, nil
	case OP_PLUS, OP_MINUS, OP_MULT, OP_DIV:
		if tLeft != TYPE_FLOAT && tLeft != TYPE_INT {
			return binaryOp, TYPE_UNKNOWN, fmt.Errorf("%w - BinaryOp %v needs int/float, got: %v", ErrCritical, binaryOp.operator, tLeft)
		}
		return binaryOp, tLeft, nil
	case OP_EQ, OP_NE, OP_LE, OP_GE, OP_LESS, OP_GREATER:
		if tLeft != TYPE_FLOAT && tLeft != TYPE_INT && tLeft != TYPE_STRING {
			return binaryOp, TYPE_UNKNOWN, fmt.Errorf("%w - BinaryOp %v needs int/float/string, got: %v", ErrCritical, binaryOp.operator, tLeft)
		}
		return binaryOp, TYPE_BOOL, nil
	default:
		return binaryOp, TYPE_UNKNOWN, fmt.Errorf("%w - BinaryOp %v !/not expected, got: %v", ErrCritical, binaryOp.operator, tLeft)
	}
	return binaryOp, tLeft, nil
}

func analyzeTypeExpression(expression Expression, symbolTable *SymbolTable) (Expression, Type, error) {

	switch e := expression.(type) {
	case Constant:
		return e, e.cType, nil
	case Variable:

		// Lookup variable type and annotate node.
		if vTable, ok := symbolTable.get(e.vName); ok {
			e.vType = vTable.sType
		} else {
			return e, TYPE_UNKNOWN, fmt.Errorf("%w - Variable type for %v unknown!", ErrCritical, e)
		}

		// Always access the very last entry for variables!
		return e, e.vType, nil
	case UnaryOp:
		return analyzeTypeUnaryOp(e, symbolTable)
	case BinaryOp:
		return analyzeTypeBinaryOp(e, symbolTable)
	}

	return expression, TYPE_UNKNOWN, fmt.Errorf("%w - Unknown type for expression %v", ErrCritical, expression)
}

//func pushNewVars(vars *map[string]Type, newVars map[string]Type) (err error) {
//	for k, v := range newVars {
//		if _, ok := (*vars)[k]; ok {
//			err = fmt.Errorf("%w - You cannot shadow the same variable multiple times within one block: %v", ErrCritical, k)
//		}
//		(*vars)[k] = v
//	}
//	return
//}

//func popVars(vars *map[string][]Type, toBeRemovedVars map[string]Type) {
//	for k, _ := range toBeRemovedVars {
//		if len((*vars)[k]) > 0 {
//			(*vars)[k] = (*vars)[k][:len((*vars)[k])-1]
//			if len((*vars)[k]) == 0 {
//				delete(*vars, k)
//			}
//		}
//	}
//}

func analyzeTypeCondition(condition Condition, symbolTable *SymbolTable) (Condition, error) {

	// This expression MUST come out as boolean!
	e, t, err := analyzeTypeExpression(condition.expression, symbolTable)
	if err != nil {
		return condition, err
	}
	if t != TYPE_BOOL {
		return condition, fmt.Errorf("%w - If expression expected boolean, got: %v (%v)", ErrCritical, t, condition)
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
		expression, t, err := analyzeTypeExpression(e, &nextSymbolTable)
		if err != nil {
			return loop, err
		}
		if t != TYPE_BOOL {
			return loop, fmt.Errorf("%w - For expression expected boolean, got: %v (%v)", ErrCritical, t, e)
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

	return loop, nil
}

// Returns newly created variables and variables that should shadow others!
// This is just for housekeeping and removing them later!!!!
// All new variables (and shadow ones) are updated/written to the symbol table
func analyzeTypeAssignment(assignment Assignment, symbolTable *SymbolTable) (Assignment, error) {

	// Populate/overwrite the dictionary of variables for futher statements :)
	if len(assignment.variables) != len(assignment.expressions) {
		return assignment, fmt.Errorf("%w - Assignment %v - variables and expression count need to match", ErrCritical, assignment)
	}

	for i, v := range assignment.variables {

		expression, expressionType, err := analyzeTypeExpression(assignment.expressions[i], symbolTable)
		if err != nil {
			return assignment, fmt.Errorf("%w - Invalid expression in assignment", err)
		}

		// Shadowing is only allowed in a different block, not right after the first variable, to avoid confusion and complicated
		// variable handling
		if _, ok := symbolTable.getLocal(v.vName); ok && v.vShadow {
			return assignment, fmt.Errorf("%w - Variable %v is shadowing in the same block. This is not allowed", ErrCritical, v.vName)
		}

		// Only, if the variable already exists and we're not trying to shadow it!
		if vTable, ok := symbolTable.get(v.vName); ok {
			if !v.vShadow {
				if vTable.sType != expressionType {
					return assignment, fmt.Errorf(
						"%w - Assignment type missmatch between variable %v and expression %v != %v", ErrCritical, v, vTable.sType, expressionType,
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
	return statement, fmt.Errorf("%w - Unexpected statement: %v", ErrCritical, statement)
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
