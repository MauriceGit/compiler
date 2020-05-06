package main

import (
	"fmt"
)

func (s Switch) allPathsReturn() bool {

	allPathsReturn := true
	hasDefault := false

	for _, c := range s.cases {

		allPathsReturn = allPathsReturn && c.block.allPathsReturn()

		if len(c.expressions) == 0 {
			hasDefault = true
		} else {
			if len(c.expressions) == 1 {
				if constant, ok := c.expressions[0].(Constant); ok {
					if constant.cType == TYPE_BOOL && constant.cValue == "true" {
						hasDefault = true
					}
				}
			}
		}
	}
	return allPathsReturn && hasDefault
}

func (b Block) allPathsReturn() bool {

	for _, s := range b.statements {
		switch st := s.(type) {
		case Block:
			if st.allPathsReturn() {
				return true
			}
		case Condition:
			if st.block.allPathsReturn() && st.elseBlock.allPathsReturn() {
				return true
			}
		case Loop:
			// We cannot guarantee, that the loop body is executed at all. So analysis on its block doesn't matter.
		case RangedLoop:
		case Switch:
			if st.allPathsReturn() {
				return true
			}
		case StructDef:
		case FunCall:
		case Assignment:
		case Break:
			return false
		case Continue:
			return false
		case Return:
			return true
		default:
			panic(fmt.Sprintf("Unknown statement '%v' for return analysis", st))
		}
	}
	return false
}

func (b Block) functionReturnAnalysis() error {

	if !b.allPathsReturn() {
		return fmt.Errorf("%wNot all code paths have a return statement", ErrCritical)
	}
	return nil
}
