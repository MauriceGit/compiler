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

func (s *SymbolTable) setType(v string, members []StructMem) {

	// Only set type definitions in the upper-most symbol table!
	if s != nil && s.parent == nil {
		s.typeTable[v] = SymbolTypeEntry{members, 0}
		return
	}
	s.parent.setType(v, members)
}

func (s *SymbolTable) getType(v string) (SymbolTypeEntry, bool) {

	// Only query type definitions from the upper-most symbol table!
	if s != nil && s.parent == nil {
		entry, ok := s.typeTable[v]
		return entry, ok
	}
	return s.parent.getType(v)
}

func (s *SymbolTable) setFun(name string, argTypes, returnTypes []ComplexType, inline bool) {
	l := s.funTable[name]
	l = append(l, SymbolFunEntry{argTypes, returnTypes, "", "", 0, inline, false})
	s.funTable[name] = l
}

func (s *SymbolTable) getLocalFun(name string, paramTypes []ComplexType, strict bool) (SymbolFunEntry, bool) {

	if s == nil {
		return SymbolFunEntry{}, false
	}
	l, ok := s.funTable[name]
	if !ok {
		return SymbolFunEntry{}, false
	}

	for _, le := range l {
		if equalTypes(le.paramTypes, paramTypes, strict) {
			return le, true
		}
	}
	return SymbolFunEntry{}, false
}

func (s *SymbolTable) isLocalFun(name string, paramTypes []ComplexType) bool {
	_, ok := s.getLocalFun(name, paramTypes, true)
	return ok
}

func (s *SymbolTable) getFun(name string, paramTypes []ComplexType, strict bool) (SymbolFunEntry, bool) {

	if s == nil {
		return SymbolFunEntry{}, false
	}

	if localFun, ok := s.getLocalFun(name, paramTypes, strict); ok {
		return localFun, true
	}

	return s.parent.getFun(name, paramTypes, strict)
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

func (s *SymbolTable) setLoopJumpLabels(br, cont string) {
	s.activeLoopBreakLabel = br
	s.activeLoopContinueLabel = cont
}

func (s *SymbolTable) getLoopBreakLabel() string {
	if s.activeLoopBreakLabel != "" {
		return s.activeLoopBreakLabel
	}
	return s.parent.getLoopBreakLabel()
}

func (s *SymbolTable) getLoopContinueLabel() string {
	if s.activeLoopContinueLabel != "" {
		return s.activeLoopContinueLabel
	}
	return s.parent.getLoopContinueLabel()
}

func (s *SymbolTable) setFunAsmName(v string, asmName string, paramTypes []ComplexType, strict bool) {
	if s == nil {
		panic("Could not set asm function name in symbol table!")
	}

	l, ok := s.funTable[v]
	if !ok {
		panic("No such function in funTable. Can not set asm name")
	}
	for i, le := range l {
		if equalTypes(le.paramTypes, paramTypes, strict) {
			le.jumpLabel = asmName
			l[i] = le
			s.funTable[v] = l
			return
		}
	}

	s.parent.setFunAsmName(v, asmName, paramTypes, strict)
}

func (s *SymbolTable) funIsInline(name string, paramTypes []ComplexType, strict bool) bool {
	if e, ok := s.getFun(name, paramTypes, strict); ok {
		return e.inline
	}
	panic("No such function in funTable. Cannot query used status")
}
func (s *SymbolTable) funIsUsed(name string, paramTypes []ComplexType, strict bool) bool {
	if e, ok := s.getFun(name, paramTypes, strict); ok {
		return e.isUsed
	}
	panic("No such function in funTable. Cannot query used status")
}

func (s *SymbolTable) setFunEpilogueLabel(v string, label string, paramTypes []ComplexType) {
	if s == nil {
		panic("Could not set asm epilogue label in symbol table!")
		return
	}

	l, ok := s.funTable[v]
	if !ok {
		panic("No such function in funTable. Can not set epilogue label.")
	}
	for i, le := range l {
		if equalTypes(le.paramTypes, paramTypes, true) {
			le.epilogueLabel = label
			l[i] = le
			s.funTable[v] = l
			return
		}

	}

	s.parent.setFunEpilogueLabel(v, label, paramTypes)
}

func (s *SymbolTable) setFunIsUsed(v string, paramTypes []ComplexType, isUsed bool) {

	if s == nil {
		panic("Could not set isUsed flag in symbol table!")
		return
	}

	if l, ok := s.funTable[v]; ok {
		for i, le := range l {

			if equalTypes(le.paramTypes, paramTypes, false) {
				le.isUsed = isUsed
				l[i] = le
				s.funTable[v] = l
				return
			}
		}
	}

	s.parent.setFunIsUsed(v, paramTypes, isUsed)
}

func (s *SymbolTable) setFunReturnStackPointer(v string, offset int, paramTypes []ComplexType) {
	if s == nil {
		panic("Could not set asm return stack pointer in symbol table!")
		return
	}

	l, ok := s.funTable[v]
	if !ok {
		panic("No such function in funTable. Can not set return stack pointer.")
	}
	for i, le := range l {
		if equalTypes(le.paramTypes, paramTypes, true) {
			le.returnStackPointerOffset = offset
			l[i] = le
			s.funTable[v] = l
			return
		}

	}

	s.parent.setFunReturnStackPointer(v, offset, paramTypes)
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
	t := expression.getExpressionTypes(symbolTable)[0]

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
		unaryOp.opType = ComplexType{TYPE_BOOL, "", nil}
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

	tLeft := binaryOp.leftExpr.getExpressionTypes(symbolTable)[0]
	tRight := binaryOp.rightExpr.getExpressionTypes(symbolTable)[0]

	// Check types only after we possibly rearranged the expression!
	if binaryOp.leftExpr.getExpressionTypes(symbolTable)[0] != binaryOp.rightExpr.getExpressionTypes(symbolTable)[0] {
		return binaryOp, fmt.Errorf(
			"%w[%v:%v] - BinaryOp '%v' expected same type, got: '%v', '%v'",
			ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft, tRight,
		)
	}

	// We match all types explicitely to make sure that this still works or create an error when we introduce new types
	// that are not considered yet!
	switch binaryOp.operator {
	case OP_AND, OP_OR:
		binaryOp.opType = ComplexType{TYPE_BOOL, "", nil}
		// We know left and right are the same type, so only compare left here.
		if tLeft.t != TYPE_BOOL {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp '%v' needs bool, got: '%v'",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
			)
		}
		//return binaryOp, TYPE_BOOL, nil
	case OP_MOD:
		binaryOp.opType = ComplexType{TYPE_INT, "", nil}
		if tLeft.t != TYPE_INT {
			return binaryOp, fmt.Errorf(
				"%w[%v:%v] - BinaryOp '%v' only works for int",
				ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator,
			)
		}
	case OP_PLUS, OP_MINUS, OP_MULT, OP_DIV:

		if tLeft.t == TYPE_FLOAT {
			binaryOp.opType = ComplexType{TYPE_FLOAT, "", nil}
		} else {
			binaryOp.opType = ComplexType{TYPE_INT, "", nil}
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
		binaryOp.opType = ComplexType{TYPE_BOOL, "", nil}
		// We can actually compare all data types. So there will be no missmatch in general!
	default:
		return binaryOp, fmt.Errorf(
			"%w[%v:%v] - Invalid binary operator: '%v' for type '%v'",
			ErrCritical, binaryOp.line, binaryOp.column, binaryOp.operator, tLeft,
		)
	}

	return binaryOp, nil
}

func analyzeDirectAccess(t ComplexType, directAccess []DirectAccess, symbolTable *SymbolTable) ([]DirectAccess, error) {

	// If any of the accesses are an Array, we switch to heap memory and have to negate the offset in memory!
	isHeapMemory := false

	for i, access := range directAccess {

		if t.t == TYPE_ARRAY {
			isHeapMemory = true
		}

		if access.indexed {
			if t.t != TYPE_ARRAY {
				row, col := access.indexExpression.startPos()
				return nil, fmt.Errorf("%w[%v:%v] - Indexing only works on arrays, ... got %v",
					ErrCritical, row, col, t.t,
				)
			}

			if t.subType == nil {
				row, col := access.indexExpression.startPos()
				return nil, fmt.Errorf("%w[%v:%v] - Indexing only works on arrays, got %v",
					ErrCritical, row, col, t.t,
				)
			}

			newIndexExpression, tmpE := analyzeExpression(access.indexExpression, symbolTable)
			if tmpE != nil {
				return nil, tmpE
			}
			directAccess[i].indexExpression = newIndexExpression

			ts := newIndexExpression.getExpressionTypes(symbolTable)
			if len(ts) != 1 || ts[0].getMemCount(symbolTable) != 1 {
				row, col := access.indexExpression.startPos()
				return nil, fmt.Errorf("%w[%v:%v] - Index expression can only have one value",
					ErrCritical, row, col,
				)
			}

			if ts[0].t != TYPE_INT {
				row, col := access.indexExpression.startPos()
				return nil, fmt.Errorf("%w[%v:%v] - Index expression must be int",
					ErrCritical, row, col,
				)
			}

			t = *t.subType
		} else {

			if t.t != TYPE_STRUCT {
				return nil, fmt.Errorf("%w[%v:%v] - Qualified name access only works on structs",
					ErrCritical, access.line, access.column,
				)
			}

			entry, ok := symbolTable.getType(t.tName)
			if !ok {
				return nil, fmt.Errorf("%w[%v:%v] - Struct type '%v' used before declaration",
					ErrCritical, access.line, access.column, t.tName,
				)
			}

			memberIndex := -1
			memberOffset := 0
			for j, m := range entry.members {
				if m.memName == access.accessName {
					memberIndex = j
					// For stack memory, the offset is negative!
					memberOffset = -m.offset
					break
				}
			}
			if memberIndex == -1 {
				return nil, fmt.Errorf("%w[%v:%v] - '%v' is not a member of struct '%v'",
					ErrCritical, access.line, access.column, access.accessName, t.tName,
				)
			}

			if isHeapMemory {
				memberOffset = -memberOffset
			}

			directAccess[i].structOffset = memberOffset

			t = entry.members[memberIndex].memType
		}
	}

	return directAccess, nil
}

// Sometimes we need a bit of special handling to work with system-side generic functions
// like 'append()', to keep our type system going on the programmer and semantic side ...
func analyzeFunCallReturnTypes(fun FunCall, args []ComplexType, symbolTable *SymbolTable) ([]ComplexType, error) {
	funEntry, _ := symbolTable.getFun(fun.funName, args, false)
	types := funEntry.returnTypes

	if len(types) == 1 && types[0].typeIsGeneric() {

		arrayType := ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}

		switch fun.funName {
		case "extend":

			if len(args) != 2 {
				return nil, fmt.Errorf("%w[%v:%v] - Function '%v' needs two parameters",
					ErrCritical, fun.line, fun.column, fun.funName,
				)
			}

			if !equalType(args[0], arrayType, false) {
				return nil, fmt.Errorf("%w[%v:%v] - Function '%v' needs array as first parameter",
					ErrCritical, fun.line, fun.column, fun.funName,
				)
			}

			// Both parameters to extend should be equal.
			if !equalType(args[0], args[1], true) {
				return nil, fmt.Errorf("%w[%v:%v] - Function '%v' can only be called with arrays (of the same type)",
					ErrCritical, fun.line, fun.column, fun.funName,
				)
			}

			// Whatever we dump into append, we get out again!
			return []ComplexType{args[0]}, nil
		case "append":

			if len(args) != 2 {
				return nil, fmt.Errorf("%w[%v:%v] - Function '%v' needs two parameters",
					ErrCritical, fun.line, fun.column, fun.funName,
				)
			}

			if !equalType(args[0], arrayType, false) {
				return nil, fmt.Errorf("%w[%v:%v] - Function '%v' needs array as first parameter",
					ErrCritical, fun.line, fun.column, fun.funName,
				)
			}

			// We expect the append element function
			// The element must have the same type as the array.
			if !equalType(*args[0].subType, args[1], true) {
				return nil, fmt.Errorf("%w[%v:%v] - Function '%v' type mismatch between array type and given element to append",
					ErrCritical, fun.line, fun.column, fun.funName,
				)
			}

			// Whatever we dump into append, we get out again!
			return []ComplexType{args[0]}, nil
		default:
			return nil, fmt.Errorf("%w[%v:%v] - Unknown generic function %v", ErrCritical, fun.line, fun.column, fun.funName)
		}
	}
	return types, nil
}

func structMemToTypeList(mems []StructMem) (types []ComplexType) {
	for _, m := range mems {
		types = append(types, m.memType)
	}
	return
}

func analyzeFunCall(fun FunCall, symbolTable *SymbolTable) (FunCall, error) {

	// The parameters must be analyzed before we query the symbol table for the actual
	// function because the argument types must be analyzed/determined by then!
	// Basically unpacking expression list before providing it to the function
	expressionTypes := []ComplexType{}
	for i, e := range fun.args {
		newE, parseErr := analyzeExpression(e, symbolTable)
		if parseErr != nil {
			return fun, parseErr
		}
		fun.args[i] = newE
		expressionTypes = append(expressionTypes, newE.getExpressionTypes(symbolTable)...)
	}

	funEntry, funOk := symbolTable.getFun(fun.funName, expressionsToTypes(fun.args, symbolTable), false)
	typeEntry, typeOk := symbolTable.getType(fun.funName)

	if funOk && typeOk {
		return fun, fmt.Errorf("%w[%v:%v] - Function name '%v' already in use as a Type", ErrCritical, fun.line, fun.column, fun.funName)
	}

	var paramTypes []ComplexType

	switch {
	case funOk:
		paramTypes = funEntry.paramTypes
		symbolTable.setFunIsUsed(fun.funName, expressionsToTypes(fun.args, symbolTable), true)

		returnTypes, err := analyzeFunCallReturnTypes(fun, expressionsToTypes(fun.args, symbolTable), symbolTable)
		if err != nil {
			return fun, err
		}
		fun.retTypes = returnTypes
		fun.createStruct = false
	case typeOk:
		paramTypes = structMemToTypeList(typeEntry.members)
		// One well defined return type. That is this struct we expect!
		fun.retTypes = []ComplexType{ComplexType{TYPE_STRUCT, fun.funName, nil}}
		fun.createStruct = true
	default:
		return fun, fmt.Errorf("%w[%v:%v] - Function '%v' called before declaration", ErrCritical, fun.line, fun.column, fun.funName)
	}

	if len(expressionTypes) != len(paramTypes) {
		return fun, fmt.Errorf("%w[%v:%v] - Function call to '%v' has %v parameters, but needs %v",
			ErrCritical, fun.line, fun.column, fun.funName, len(expressionTypes), len(funEntry.paramTypes),
		)
	}

	for i, t := range expressionTypes {

		if !equalType(t, paramTypes[i], false) {
			return fun, fmt.Errorf("%w[%v:%v] - Function call to '%v' got type %v as %v. parameter, but needs %v",
				ErrCritical, fun.line, fun.column, fun.funName, t, i+1, paramTypes[i],
			)
		}
	}

	if len(fun.directAccess) > 0 {
		if fun.getResultCount() != 1 {
			return fun, fmt.Errorf("%w[%v:%v] - Indexing is only allowed for single return functions",
				ErrCritical, fun.line, fun.column,
			)
		}
		newDirectAccess, err := analyzeDirectAccess(fun.retTypes[0], fun.directAccess, symbolTable)
		if err != nil {
			return fun, err
		}
		fun.directAccess = newDirectAccess
	}

	return fun, nil
}

func analyzeArrayDecl(a Array, symbolTable *SymbolTable) (Array, error) {

	var arrayType ComplexType = ComplexType{TYPE_UNKNOWN, "", nil}
	arraySize := 0
	for i, e := range a.aExpressions {
		newE, err := analyzeExpression(e, symbolTable)
		if err != nil {
			return a, err
		}
		a.aExpressions[i] = newE
		if i == 0 && a.aType.t == TYPE_UNKNOWN {
			arrayType = ComplexType{TYPE_ARRAY, "", &newE.getExpressionTypes(symbolTable)[0]}
		}
		if !equalType(newE.getExpressionTypes(symbolTable)[0], *arrayType.subType, true) {
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

	if e := analyzeType(a.aType, symbolTable); e != nil {
		return a, fmt.Errorf("%w[%v:%v] - %v", ErrCritical, a.line, a.column, e.Error())
	}

	if a.isDirectlyAccessed() {
		newDirectAccess, err := analyzeDirectAccess(a.aType, a.directAccess, symbolTable)
		if err != nil {
			return a, err
		}
		a.directAccess = newDirectAccess
	}

	return a, nil
}

func analyzeVariable(e Variable, symbolTable *SymbolTable) (Variable, error) {
	// Lookup variable type and annotate node.
	if vTable, ok := symbolTable.getVar(e.vName); ok {
		e.vType = vTable.sType

		if e.isDirectlyAccessed() {
			newDirectAccess, err := analyzeDirectAccess(e.vType, e.directAccess, symbolTable)
			if err != nil {
				return e, err
			}
			e.directAccess = newDirectAccess
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
		v, err := analyzeVariable(e, symbolTable)
		if err == nil && v.vShadow {
			err = fmt.Errorf("%w[%v:%v] - Variable used as expression can not use 'shadow' keyword", ErrCritical, v.line, v.column)
		}
		return v, err
	case UnaryOp:
		return analyzeUnaryOp(e, symbolTable)
	case BinaryOp:
		return analyzeBinaryOp(e, symbolTable)
	case FunCall:
		return analyzeFunCall(e, symbolTable)
	case Array:
		return analyzeArrayDecl(e, symbolTable)
	default:
		row, col := expression.startPos()
		return expression, fmt.Errorf("%w[%v:%v] - Unknown type for expression '%v'", ErrCritical, row, col, expression)
	}

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

		tmpTypes := expression.getExpressionTypes(symbolTable)
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

				variableType := getAccessedType(vTable.sType, v.directAccess, symbolTable)

				if !equalType(variableType, expressionType, true) {
					return assignment, fmt.Errorf(
						"%w[%v:%v] - Assignment type missmatch between variable %v and expression %v",
						ErrCritical, v.line, v.column, v.vType, expressionType,
					)
				}
			} else {
				if v.isDirectlyAccessed() {
					return assignment, fmt.Errorf("%w[%v:%v] - An indexed array write can not shadow its source",
						ErrCritical, v.line, v.column,
					)
				}

				symbolTable.setVar(v.vName, expressionType, false)
			}
		} else {
			symbolTable.setVar(v.vName, expressionType, v.isDirectlyAccessed())
		}

		assignment.variables[i].vType = expressionType

		if v.isDirectlyAccessed() {

			entry, _ := symbolTable.getVar(v.vName)
			newDirectAccess, err := analyzeDirectAccess(entry.sType, v.directAccess, symbolTable)
			if err != nil {
				return assignment, err
			}
			assignment.variables[i].directAccess = newDirectAccess
		}

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
	t := e.getExpressionTypes(symbolTable)[0]
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

func analyzeSwitch(sc Switch, symbolTable *SymbolTable) (Switch, error) {

	valueSwitch := sc.expression != nil
	if valueSwitch {
		e, err := analyzeExpression(sc.expression, symbolTable)
		if err != nil {
			return sc, err
		}
		sc.expression = e
	}

	for i, c := range sc.cases {
		// Empty case is permitted (behaves just like a default) as the last case only!
		// And only, if we have a value-switch! Otherwise the default will be 'true'
		if len(c.expressions) == 0 && i != len(sc.cases)-1 && sc.expression != nil {
			row, col := sc.startPos()
			return sc, fmt.Errorf("%w[%v:%v] - empty/default case must be the last case or have an expression to match",
				ErrCritical, row, col,
			)
		}
		for j, ce := range c.expressions {
			e, err := analyzeExpression(ce, symbolTable)
			if err != nil {
				return sc, err
			}
			if e.getResultCount() != 1 {
				row, col := e.startPos()
				return sc, fmt.Errorf("%w[%v:%v] - case expression can only have one result value",
					ErrCritical, row, col,
				)
			}
			if valueSwitch {
				if e.getExpressionTypes(symbolTable)[0].t != TYPE_INT {
					row, col := e.startPos()
					return sc, fmt.Errorf("%w[%v:%v] - value switch only accepts integers as cases",
						ErrCritical, row, col,
					)
				}
			} else {
				if e.getExpressionTypes(symbolTable)[0].t != TYPE_BOOL {
					row, col := e.startPos()
					return sc, fmt.Errorf("%w[%v:%v] - general switch only accepts bools as cases",
						ErrCritical, row, col,
					)
				}
			}
			c.expressions[j] = e
		}

		block, err := analyzeBlock(c.block, symbolTable, nil)
		if err != nil {
			return sc, err
		}
		c.block = block

		sc.cases[i] = c
	}

	return sc, nil
}

func analyzeLoop(loop Loop, symbolTable *SymbolTable) (Loop, error) {

	nextSymbolTable := &SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string][]SymbolFunEntry, 0),
		make(map[string]SymbolTypeEntry, 0),
		symbolTable.activeFunctionName,
		symbolTable.activeFunctionParams,
		symbolTable.activeFunctionReturn,
		true,
		"",
		"",
		symbolTable,
	}

	assignment, err := analyzeAssignment(loop.assignment, nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.assignment = assignment

	for i, e := range loop.expressions {
		expression, err := analyzeExpression(e, nextSymbolTable)
		if err != nil {
			return loop, err
		}

		for _, t := range expression.getExpressionTypes(symbolTable) {
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

	incrAssignment, err := analyzeAssignment(loop.incrAssignment, nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.incrAssignment = incrAssignment

	statements, err := analyzeBlock(loop.block, symbolTable, nextSymbolTable)
	if err != nil {
		return loop, err
	}
	loop.block = statements
	loop.block.symbolTable = nextSymbolTable

	return loop, nil
}

func analyzeRangedLoop(loop RangedLoop, symbolTable *SymbolTable) (RangedLoop, error) {

	nextSymbolTable := &SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string][]SymbolFunEntry, 0),
		make(map[string]SymbolTypeEntry, 0),
		symbolTable.activeFunctionName,
		symbolTable.activeFunctionParams,
		symbolTable.activeFunctionReturn,
		true,
		"",
		"",
		symbolTable,
	}

	rangeExpression, err := analyzeExpression(loop.rangeExpression, nextSymbolTable)
	if err != nil {
		return loop, err
	}

	if rangeExpression.getResultCount() != 1 {
		return loop, fmt.Errorf("%w[%v:%v] - RangedLoop can only iterate one array at a time", ErrCritical, loop.line, loop.column)
	}
	rangeType := rangeExpression.getExpressionTypes(symbolTable)[0]

	if !equalType(rangeType, ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}, false) {
		return loop, fmt.Errorf("%w[%v:%v] - RangedLoop can only iterate arrays", ErrCritical, loop.line, loop.column)
	}

	loop.elem.vType = *rangeType.subType
	// As we don't have an assignment (that does it for us), we have to register the variables ourself
	nextSymbolTable.setVar(loop.elem.vName, loop.elem.vType, false)
	nextSymbolTable.setVar(loop.counter.vName, loop.counter.vType, false)

	statements, err := analyzeBlock(loop.block, symbolTable, nextSymbolTable)
	if err != nil {
		return loop, err
	}

	loop.rangeExpression = rangeExpression
	loop.block = statements
	loop.block.symbolTable = nextSymbolTable

	return loop, nil
}

func analyzeFunction(fun Function, symbolTable *SymbolTable) (Function, error) {

	functionSymbolTable := &SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string][]SymbolFunEntry, 0),
		make(map[string]SymbolTypeEntry, 0),
		fun.fName,
		variablesToTypes(fun.parameters),
		fun.returnTypes,
		false,
		"",
		"",
		symbolTable,
	}

	for _, v := range fun.parameters {
		if v.vType.t == TYPE_UNKNOWN {
			return fun, fmt.Errorf("%w[%v:%v] - Function parameter %v has invalid type", ErrCritical, v.line, v.column, v)
		}
		if _, ok := functionSymbolTable.getVar(v.vName); ok {
			return fun, fmt.Errorf("%w[%v:%v] - Function parameter %v already exists", ErrCritical, v.line, v.column, v)
		}
		if e := analyzeType(v.vType, symbolTable); e != nil {
			return fun, fmt.Errorf("%w[%v:%v] - %v", ErrCritical, v.line, v.column, e.Error())
		}

		functionSymbolTable.setVar(v.vName, v.vType, false)
	}

	if symbolTable.isLocalFun(fun.fName, variablesToTypes(fun.parameters)) {
		return fun, fmt.Errorf("%w[%v:%v] - Function with the same name already exists in this scope", ErrCritical, fun.line, fun.column)
	}

	symbolTable.setFun(fun.fName, variablesToTypes(fun.parameters), fun.returnTypes, false)

	newBlock, err := analyzeBlock(fun.block, symbolTable, functionSymbolTable)
	if err != nil {
		return fun, err
	}
	fun.block = newBlock

	// Checks, that every single path has one return statement so we don't have undefined function returns,
	// unassigned registers or no expected values on stack
	if len(fun.returnTypes) > 0 {
		if err = fun.block.functionReturnAnalysis(); err != nil {
			return fun, fmt.Errorf("%w[%v:%v] - %v", ErrCritical, fun.line, fun.column, err.Error())
		}
	}

	return fun, nil
}

func analyzeReturn(ret Return, symbolTable *SymbolTable) (Return, error) {

	if symbolTable.activeFunctionName == "" {
		return ret, fmt.Errorf("%w[%v:%v] - Return called outside of a function",
			ErrCritical, ret.line, ret.column,
		)
	}

	typeIndex := 0
	for i, e := range ret.expressions {

		newE, err := analyzeExpression(e, symbolTable)
		if err != nil {
			return ret, err
		}
		row, col := e.startPos()
		for _, t := range newE.getExpressionTypes(symbolTable) {

			if typeIndex >= len(symbolTable.activeFunctionReturn) {
				return ret, fmt.Errorf("%w[%v:%v] - Too many expressions returned. Expected %v",
					ErrCritical, row, col, len(symbolTable.activeFunctionReturn),
				)
			}

			if !equalType(t, symbolTable.activeFunctionReturn[typeIndex], true) {
				return ret, fmt.Errorf("%w[%v:%v] - Function return type does not match definition. Expected %v, got %v",
					ErrCritical, row, col, symbolTable.activeFunctionReturn[typeIndex], t,
				)
			}
			typeIndex++
		}
		ret.expressions[i] = newE
	}
	return ret, nil
}

func analyzeBreak(br Break, symbolTable *SymbolTable) (Break, error) {
	if !symbolTable.activeLoop {
		return br, fmt.Errorf("%w[%v:%v] - Break statement outside of loop block", ErrCritical, br.line, br.column)
	}
	return br, nil
}

func analyzeContinue(cont Continue, symbolTable *SymbolTable) (Continue, error) {
	if !symbolTable.activeLoop {
		return cont, fmt.Errorf("%w[%v:%v] - Continue statement outside of loop block", ErrCritical, cont.line, cont.column)
	}
	return cont, nil
}

// analyzeType recursively goes through the type hierarchy and checks, if all types are well defined.
func analyzeType(t ComplexType, symbolTable *SymbolTable) error {

	// Needs to be defined!
	if t.t == TYPE_STRUCT {
		if _, ok := symbolTable.getType(t.tName); !ok {
			return fmt.Errorf("Struct name '%v' undefined", t.tName)
		}
		return nil
	}
	if t.t == TYPE_ARRAY {
		return analyzeType(*t.subType, symbolTable)
	}
	if t.t == TYPE_UNKNOWN {
		return fmt.Errorf("Type '%v' is unknown", t.t)
	}
	return nil
}

func analyzeStructDef(st StructDef, symbolTable *SymbolTable) (StructDef, error) {

	if _, ok := symbolTable.getType(st.name); ok {
		return st, fmt.Errorf("%w[%v:%v] - Struct name '%v' already used", ErrCritical, st.line, st.column, st.name)
	}

	if len(st.members) == 0 {
		return st, fmt.Errorf("%w[%v:%v] - A Struct can not have zero members", ErrCritical, st.line, st.column)
	}

	offset := 0
	// Go through the complex type of each member and check, that all types are well defined.
	for i, m := range st.members {
		if e := analyzeType(m.memType, symbolTable); e != nil {
			return st, fmt.Errorf("%w[%v:%v] - %v", ErrCritical, st.line, st.column, e.Error())
		}
		// Check against the remaining members, that each member name is unique!
		for j := i + 1; j < len(st.members); j++ {
			if m.memName == st.members[j].memName {
				return st, fmt.Errorf("%w[%v:%v] - The struct member '%v' is defined multiple times", ErrCritical, st.line, st.column, m.memName)
			}
		}

		st.members[i].offset = offset
		offset += m.memType.getMemCount(symbolTable)
	}

	symbolTable.setType(st.name, st.members)

	return st, nil
}

func analyzeStatement(statement Statement, symbolTable *SymbolTable) (Statement, error) {
	switch st := statement.(type) {
	case StructDef:
		return analyzeStructDef(st, symbolTable)
	case Condition:
		return analyzeCondition(st, symbolTable)
	case Switch:
		return analyzeSwitch(st, symbolTable)
	case Loop:
		return analyzeLoop(st, symbolTable)
	case RangedLoop:
		return analyzeRangedLoop(st, symbolTable)
	case Assignment:
		return analyzeAssignment(st, symbolTable)
	case Function:
		return analyzeFunction(st, symbolTable)
	case Return:
		return analyzeReturn(st, symbolTable)
	case FunCall:
		return analyzeFunCall(st, symbolTable)
	case Break:
		return analyzeBreak(st, symbolTable)
	case Continue:
		return analyzeContinue(st, symbolTable)
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
		block.symbolTable = newBlockSymbolTable
	} else {
		block.symbolTable = &SymbolTable{
			make(map[string]SymbolVarEntry, 0),
			make(map[string][]SymbolFunEntry, 0),
			make(map[string]SymbolTypeEntry, 0),
			symbolTable.activeFunctionName,
			symbolTable.activeFunctionParams,
			symbolTable.activeFunctionReturn,
			symbolTable.activeLoop,
			symbolTable.activeLoopBreakLabel,
			symbolTable.activeLoopContinueLabel,
			symbolTable,
		}
	}

	for i, s := range block.statements {
		statement, err := analyzeStatement(s, block.symbolTable)
		if err != nil {
			return block, err
		}
		block.statements[i] = statement
	}

	return block, nil
}

// Set usage status for system the functions above. This can be made nicer and general, but the effort is not
// worth it right now. Maybe at a later point.
func setSystemFunctionUsage(s *SymbolTable) {

	cArg := []ComplexType{ComplexType{TYPE_CHAR, "", nil}}
	iArg := []ComplexType{ComplexType{TYPE_INT, "", nil}}
	fArg := []ComplexType{ComplexType{TYPE_FLOAT, "", nil}}
	aArg := []ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}}
	aeArg := []ComplexType{
		ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}},
		ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}},
	}
	aaArg := []ComplexType{
		ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}},
		ComplexType{TYPE_WHATEVER, "", nil},
	}

	plnI := s.funIsUsed("println", iArg, true)
	pI := s.funIsUsed("print", iArg, true)

	plnF := s.funIsUsed("println", fArg, true)
	pF := s.funIsUsed("print", fArg, true)

	plnC := s.funIsUsed("println", cArg, true)
	pC := s.funIsUsed("print", cArg, true)

	s.setFunIsUsed("print", fArg, pF || plnF)
	s.setFunIsUsed("print", iArg, pI || plnI || s.funIsUsed("print", fArg, true))
	s.setFunIsUsed("print", cArg, pC || plnC)

	// We need to re-query them because we just possibly changed their state.
	s.setFunIsUsed("printChar", iArg, s.funIsUsed("print", iArg, true) || s.funIsUsed("print", fArg, true) || s.funIsUsed("print", cArg, true))

	s.setFunIsUsed("free", aArg,
		s.funIsUsed("free", aArg, false) ||
			s.funIsUsed("extend", aeArg, false) ||
			s.funIsUsed("append", aaArg, false),
	)

}

// analyzeTypes traverses the tree and analyzes variables with their corresponding type recursively from expressions!
// returns an error if we have a type missmatch anywhere!
func semanticAnalysis(ast AST) (AST, error) {

	ast.globalSymbolTable = &SymbolTable{
		make(map[string]SymbolVarEntry, 0),
		make(map[string][]SymbolFunEntry, 0),
		make(map[string]SymbolTypeEntry, 0),
		"",
		nil,
		nil,
		false,
		"",
		"",
		nil,
	}

	ast.globalSymbolTable.setFun("printChar", []ComplexType{ComplexType{TYPE_INT, "", nil}}, []ComplexType{}, false)
	ast.globalSymbolTable.setFun("print", []ComplexType{ComplexType{TYPE_INT, "", nil}}, []ComplexType{}, false)
	ast.globalSymbolTable.setFun("println", []ComplexType{ComplexType{TYPE_INT, "", nil}}, []ComplexType{}, false)
	ast.globalSymbolTable.setFun("print", []ComplexType{ComplexType{TYPE_FLOAT, "", nil}}, []ComplexType{}, false)
	ast.globalSymbolTable.setFun("println", []ComplexType{ComplexType{TYPE_FLOAT, "", nil}}, []ComplexType{}, false)
	ast.globalSymbolTable.setFun("print", []ComplexType{ComplexType{TYPE_CHAR, "", nil}}, []ComplexType{}, false)
	ast.globalSymbolTable.setFun("println", []ComplexType{ComplexType{TYPE_CHAR, "", nil}}, []ComplexType{}, false)

	ast.globalSymbolTable.setFun("char",
		[]ComplexType{ComplexType{TYPE_INT, "", nil}},
		[]ComplexType{ComplexType{TYPE_CHAR, "", nil}}, true,
	)

	ast.globalSymbolTable.setFun("int",
		[]ComplexType{ComplexType{TYPE_FLOAT, "", nil}},
		[]ComplexType{ComplexType{TYPE_INT, "", nil}}, true,
	)
	ast.globalSymbolTable.setFun("int",
		[]ComplexType{ComplexType{TYPE_CHAR, "", nil}},
		[]ComplexType{ComplexType{TYPE_INT, "", nil}}, true,
	)
	ast.globalSymbolTable.setFun("float",
		[]ComplexType{ComplexType{TYPE_INT, "", nil}},
		[]ComplexType{ComplexType{TYPE_FLOAT, "", nil}}, true,
	)

	ast.globalSymbolTable.setFun("cap",
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		[]ComplexType{ComplexType{TYPE_INT, "", nil}}, true,
	)
	ast.globalSymbolTable.setFun("len",
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		[]ComplexType{ComplexType{TYPE_INT, "", nil}}, true,
	)
	// free can not be set as inline, as we have to call it explicitely from assembly in append()
	ast.globalSymbolTable.setFun("free",
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		[]ComplexType{}, false,
	)
	ast.globalSymbolTable.setFun("reset",
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		[]ComplexType{}, true,
	)
	ast.globalSymbolTable.setFun("clear",
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		[]ComplexType{}, true,
	)

	ast.globalSymbolTable.setFun("extend",
		[]ComplexType{
			ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}},
			ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}},
		},
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		false,
	)

	ast.globalSymbolTable.setFun("append",
		[]ComplexType{
			ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}},
			ComplexType{TYPE_WHATEVER, "", nil},
		},
		[]ComplexType{ComplexType{TYPE_ARRAY, "", &ComplexType{TYPE_WHATEVER, "", nil}}},
		false,
	)

	block, err := analyzeBlock(ast.block, nil, ast.globalSymbolTable)
	if err != nil {
		ast.globalSymbolTable = &SymbolTable{}
		return ast, err
	}
	ast.block = block

	setSystemFunctionUsage(ast.globalSymbolTable)

	return ast, nil
}
