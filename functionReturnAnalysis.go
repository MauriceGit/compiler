package main

import (
	"fmt"
)

func (s Condition) functionReturnAnalysis() error {
	if err := s.block.functionReturnAnalysis(); err != nil {
		return err
	}
	if err := s.elseBlock.functionReturnAnalysis(); err != nil {
		return err
	}
	return nil
}

func (b Block) functionReturnAnalysis() error {

	// Shallow check so we don't need a recursive descent.
	for _, s := range b.statements {
		switch s.(type) {
		case Return:
			return nil
		}
	}

	// That means, that we are not in a leaf-node. If we are then we can return an error anyway.
	foundReturns := 0
	for _, s := range b.statements {

		switch st := s.(type) {
		case Block:
			if err := st.functionReturnAnalysis(); err != nil {
				return err
			}
			foundReturns++
		case Condition:
			if err := st.functionReturnAnalysis(); err != nil {
				return err
			}
			foundReturns++
		case Loop:
			if err := st.block.functionReturnAnalysis(); err != nil {
				return err
			}
			foundReturns++
		case RangedLoop:
			if err := st.block.functionReturnAnalysis(); err != nil {
				return err
			}
			foundReturns++
		case Switch:
			for _, m := range st.cases {
				if err := m.block.functionReturnAnalysis(); err != nil {
					return err
				}
				foundReturns++
			}
		case StructDef:
		case FunCall:
		case Assignment:
		case Break:
		case Continue:
		default:
			panic(fmt.Sprintf("Unknown statement '%v' for return analysis", st))
		}
	}

	if foundReturns == 0 {
		return fmt.Errorf("%wNot all code paths have a return statement", ErrCritical)
	}

	return nil
}
