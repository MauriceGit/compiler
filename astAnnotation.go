package main

import (
	"fmt"
)

func annotateTypeUnaryOp(unaryOp UnaryOp, vars map[string][]Type) (Expression, Type, error) {
	expression, t, err := annotateTypeExpression(unaryOp.expr, vars)
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

func annotateTypeBinaryOp(binaryOp BinaryOp, vars map[string][]Type) (Expression, Type, error) {
	leftExpression, tLeft, err := annotateTypeExpression(binaryOp.leftExpr, vars)
	if err != nil {
		return binaryOp, tLeft, err
	}
	binaryOp.leftExpr = leftExpression

	rightExpression, tRight, err := annotateTypeExpression(binaryOp.rightExpr, vars)
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

func annotateTypeExpression(expression Expression, vars map[string][]Type) (Expression, Type, error) {

	switch e := expression.(type) {
	case Constant:
		return e, e.cType, nil
	case Variable:
		t, ok := vars[e.vName]
		if !ok || len(t) == 0 {
			return e, TYPE_UNKNOWN, fmt.Errorf("%w - Variable type for %v unknown!", ErrCritical, e)
		}

		// annotate with type!
		e.vType = t[len(t)-1]

		// Always access the very last entry for variables!
		return e, e.vType, nil
	case UnaryOp:
		return annotateTypeUnaryOp(e, vars)
	case BinaryOp:
		return annotateTypeBinaryOp(e, vars)
	}

	return expression, TYPE_UNKNOWN, fmt.Errorf("%w - Unknown type for expression %v", ErrCritical, expression)
}

func pushNewVars(vars *map[string]Type, newVars map[string]Type) (err error) {
	for k, v := range newVars {
		if _, ok := (*vars)[k]; ok {
			err = fmt.Errorf("%w - You cannot shadow the same variable multiple times within one block: %v", ErrCritical, k)
		}
		(*vars)[k] = v
	}
	return
}

func popVars(vars *map[string][]Type, toBeRemovedVars map[string]Type) {
	for k, _ := range toBeRemovedVars {
		if len((*vars)[k]) > 0 {
			(*vars)[k] = (*vars)[k][:len((*vars)[k])-1]
			if len((*vars)[k]) == 0 {
				delete(*vars, k)
			}
		}
	}
}

func annotateTypeCondition(condition Condition, vars map[string][]Type) (Condition, error) {

	// This expression MUST come out as boolean!
	e, t, err := annotateTypeExpression(condition.expression, vars)
	if err != nil {
		return condition, err
	}
	if t != TYPE_BOOL {
		return condition, fmt.Errorf("%w - If expression expected boolean, got: %v (%v)", ErrCritical, t, condition)
	}
	condition.expression = e

	block, err := annotateTypeBlock(condition.block, vars)
	if err != nil {
		return condition, err
	}
	condition.block = block

	elseBlock, err := annotateTypeBlock(condition.elseBlock, vars)
	if err != nil {
		return condition, err
	}
	condition.elseBlock = elseBlock

	return condition, nil
}

func annotateTypeLoop(loop Loop, vars map[string][]Type) (Loop, error) {
	// Everything that happens in a loop, stays in a loop ;)
	localShadowVars := make(map[string]Type, 0)

	defer popVars(&vars, localShadowVars)

	assignment, assignmentVars, err := annotateTypeAssignment(loop.assignment, vars)
	if err != nil {
		return loop, err
	}
	loop.assignment = assignment

	// Ignore errors because we basically have a new block anyway.
	pushNewVars(&localShadowVars, assignmentVars)

	for i, e := range loop.expressions {
		expression, t, err := annotateTypeExpression(e, vars)
		if err != nil {
			return loop, err
		}
		if t != TYPE_BOOL {
			return loop, fmt.Errorf("%w - For expression expected boolean, got: %v (%v)", ErrCritical, t, e)
		}

		loop.expressions[i] = expression
	}

	assignment, incrVars, err := annotateTypeAssignment(loop.incrAssignment, vars)
	if err != nil {
		return loop, err
	}
	loop.incrAssignment = assignment

	// Ignore errors because we basically have a new block anyway.
	pushNewVars(&localShadowVars, incrVars)

	statements, err := annotateTypeBlock(loop.block, vars)
	if err != nil {
		return loop, err
	}
	loop.block = statements

	return loop, nil
}

// Returns newly created variables and variables that should shadow others!
// This is just for housekeeping and removing them later!!!!
// All new variables (and shadow ones) are also already added into the 'vars' map!!!!!
func annotateTypeAssignment(assignment Assignment, vars map[string][]Type) (Assignment, map[string]Type, error) {

	shadowVars := make(map[string]Type, 0)
	// Populate/overwrite the dictionary of variables for futher statements :)
	if len(assignment.variables) != len(assignment.expressions) {
		return assignment, nil, fmt.Errorf("%w - Assignment %v - variables and expression count need to match", ErrCritical, assignment)
	}

	for i, v := range assignment.variables {

		expression, t, err := annotateTypeExpression(assignment.expressions[i], vars)
		if err != nil {
			return assignment, nil, err
		}

		if vCache, ok := vars[v.vName]; ok {
			if !v.vShadow && vCache[len(vCache)-1] != t {
				return assignment, nil, fmt.Errorf("%w - Variable %v, type %v already exists and is assigned a wrong type %v", ErrCritical, v, vCache[len(vCache)-1], t)
			}
		}

		// If we already have a type, it is not shadowing (new!) and the expression is of a different type.
		if v.vType != TYPE_UNKNOWN && !v.vShadow && v.vType != t {
			return assignment, nil, fmt.Errorf("%w - Variable type %v is different to assigned expression type %v", ErrCritical, v.vType, t)
		}
		assignment.expressions[i] = expression

		// Annotate variable with type
		assignment.variables[i].vType = t

		// New variable or shadowing one
		if _, ok := vars[v.vName]; !ok || v.vShadow {
			vars[v.vName] = append(vars[v.vName], t)
			shadowVars[v.vName] = t
		}
	}
	return assignment, shadowVars, nil
}

func annotateTypeStatement(statement Statement, vars map[string][]Type, shadowVars map[string]Type) (Statement, error) {
	switch st := statement.(type) {
	case Condition:
		return annotateTypeCondition(st, vars)
	case Loop:
		return annotateTypeLoop(st, vars)
	case Assignment:
		assignment, newShadowVars, err := annotateTypeAssignment(st, vars)
		if err != nil {
			return assignment, err
		}
		err = pushNewVars(&shadowVars, newShadowVars)
		if err != nil {
			return assignment, err
		}

		return assignment, nil
	}
	return statement, fmt.Errorf("%w - Unexpected statement: %v", ErrCritical, statement)
}

func annotateTypeBlock(block Block, vars map[string][]Type) (Block, error) {

	// Only includes variables that are meant to shadow outer variables!
	// This is now a different variable (which also means, that it can change its Type!) until the end of the current block!
	// They need to be removed from 'vars' at the end of the block.
	shadowVars := make(map[string]Type, 0)

	// Remove all variables that were created in this block from the vars dictionary.
	// Additionally also shadowing ones. Old types from before should be correct again!
	defer popVars(&vars, shadowVars)

	for i, s := range block.statements {
		statement, err := annotateTypeStatement(s, vars, shadowVars)
		if err != nil {
			return block, err
		}
		block.statements[i] = statement
	}

	return block, nil
}

// annotateTypes traverses the tree and annotates variables with their corresponding type recursively from expressions!
// returns an error if we have a type missmatch anywhere!
func annotateTypes(ast AST) (AST, error) {
	vars := make(map[string][]Type, 0)

	block, err := annotateTypeBlock(ast.block, vars)
	if err != nil {
		return ast, err
	}
	ast.block = block

	return ast, nil
}
