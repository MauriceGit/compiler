package main

import (
	"fmt"
)

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
			if st.block.allPathsReturn() {
				return true
			}
		case RangedLoop:
			if st.block.allPathsReturn() {
				return true
			}
		case Switch:
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
