package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func compileAndRun(program []byte) {
	if !compile(program, "", "executable") {
		return
	}
	absPath, _ := filepath.Abs("./executable")
	cmd := exec.Command(absPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// ExampleRecursion checks, that function stacks and recursion works as expected
func ExampleRecursion() {
	var program []byte = []byte(`
		fun sum(i int) int {
			if i == 0 {
				return 0
			}
			return i+sum(i-1)
		}
		println(sum(10))`,
	)
	compileAndRun(program)

	// Output:
	// 55
}

// ExampleArray1 checks, that handling arrays as parameters into functions works correctly
func ExampleArray1() {
	var program []byte = []byte(`
		fun sumArray(list []int) int {
			s = 0
			for i = 0; i < 5; i++ {
				s += list[i]
			}
			return s
		}
		list = [1, 2, 3, 4, 5]
		println(sumArray(list))`,
	)
	compileAndRun(program)

	// Output:
	// 15
}

// ExampleArray2 checks, that moving arrays, assigning them into other arrays and reading them out again is working correctly
func ExampleArray2() {
	var program []byte = []byte(`
		l1 = [1, 2, 3, 4, 5]
		l2 = [](int, 5)
		l2[3] = 4555
		l3 = [l1, l2]
		println(l3[1][3])`,
	)
	compileAndRun(program)

	// Output:
	// 4555
}

// ExampleArray3 checks, that arrays are passed correctly through a function (in and out)
func ExampleArray3() {
	var program []byte = []byte(`
		fun abc(list []int) []int {
			return list
		}
		println(abc([1,2,3,4,5])[2])`,
	)
	compileAndRun(program)

	// Output:
	// 3
}

// ExampleArrayLen checks, that the internal len() function works correctly on both array definition formats
func ExampleArrayLen() {
	var program []byte = []byte(`
		println(len([1,2,3,4,5,6,7,8]))
		println(len([](float, 134)))
		println(len([](int, 0)))`,
	)
	compileAndRun(program)

	// Output:
	// 8
	// 134
	// 0
}

// ExampleMultiAssignment checks, that multi-value assignments and automatic unpacking works correctly
func ExampleMultiAssignment() {
	var program []byte = []byte(`
		fun abc(v1 int, v2 int, v3 int) int, int, int {
			return v3, v2, v1
		}
		a,b,c,d,e = 0, abc(1,2,3), 4
		println(a)
		println(b)
		println(c)
		println(d)
		println(e)`,
	)
	compileAndRun(program)

	// Output:
	// 0
	// 3
	// 2
	// 1
	// 4
}

// ExampleMultiParameter checks, that given > 6 parameters, they will be passed to the function correctly.
// This is important, as with > 6 parameters, we need to pass them on the stack instead of registers, so
// there is some special case handling here
func ExampleMultiParameter() {
	var program []byte = []byte(`
		fun sum(a int, b int, c int, d int, e int, f int, g int, h int) int {
			return a+b+c+d+e+f+g+h
		}
		println(sum(1,2,3,4,5,6,7,8))`,
	)
	compileAndRun(program)

	// Output:
	// 36
}

// ExampleMultiReturn checks, that returning multiple values works correctly. This is important, as
// we need to make space on the stack before calling the function and have special handling with the
// stack pointer, as the calling conventions do not cover multi-return functions.
func ExampleMultiReturn() {
	var program []byte = []byte(`
		fun abc() int, int, int, int, int, int {
			return 1,2,3,4,5,6
		}
		a,b,c,d,e,f = abc()
		println(a)
		println(b)
		println(c)
		println(d)
		println(e)
		println(f)`,
	)
	compileAndRun(program)

	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
}

// ExampleFunctionOverload checks, if overloaded function names work as expected. Both for the compiler
// println functions and the new abc() functions
func ExampleFunctionOverload() {
	var program []byte = []byte(`
		fun abc(i int) {
			println(i)
		}
		fun abc(i int, j int) {
			println(i)
			println(j)
		}
		fun abc(i float) {
			println(i)
		}
		fun abc(i float, j float) {
			println(i)
			println(j)
		}
		abc(5)
		abc(6, 7)
		abc(5.5)
		abc(5.5, 6.5)
		`,
	)
	compileAndRun(program)

	// Output:
	// 5
	// 6
	// 7
	// 5.500
	// 5.500
	// 6.500
}
