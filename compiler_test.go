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

// ExampleArrayCap checks, that the internal cap() function works correctly on both array definition formats
func ExampleArrayCap() {
	var program []byte = []byte(`
		println(cap([1,2,3,4,5,6,7,8]))
		println(cap([](float, 134)))
		println(cap([](int, 0)))`,
	)
	compileAndRun(program)

	// Output:
	// 8
	// 134
	// 0
}

// ExampleArrayFree runs the internal free() function on some arrays and prints something afterwards to verify, that
// nothing crashed.
func ExampleArrayFree() {
	var program []byte = []byte(`

		a = [1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]
		b = [](int, 1234565)
		c = [](float, 1234565)
		d = [](bool, 1234565)

		free(a)
		free(b)
		free(c)
		free(d)

		println(555.666)
		`,
	)
	compileAndRun(program)

	// Output:
	// 555.666
}

// ExampleArrayReset runs the internal reset() function and checks, if the length has been reset.
func ExampleArrayReset() {
	var program []byte = []byte(`
		a = [1,2,3,4,5]
		println(len(a))
		println(cap(a))
		println(a[2])
		reset(a)
		println(a[2])
		println(len(a))
		println(cap(a))
		`,
	)
	compileAndRun(program)

	// Output:
	// 5
	// 5
	// 3
	// 3
	// 0
	// 5
}

// ExampleArrayClear runs the internal clear() function and checks, if the memory has been cleared.
func ExampleArrayClear() {
	var program []byte = []byte(`
		a = [1,2,3,4,5]
		println(len(a))
		println(cap(a))
		println(a[2])
		clear(a)
		println(a[2])
		println(len(a))
		println(cap(a))
		`,
	)
	compileAndRun(program)

	// Output:
	// 5
	// 5
	// 3
	// 0
	// 0
	// 5
}

// ExampleArrayAppend tests the internal extend() function and checks the length of the resulting arrays
// for both floats and ints
func ExampleArrayExtend() {
	var program []byte = []byte(`
		a = [1]
		b = [7,8,9]

		c = [1.5, 2.5]
		d = [3.5, 4.5]

		a = extend(a, b)
		c = extend(c, d)

		println(a[0])
		println(a[1])
		println(a[2])
		println(a[3])

		println(c[0])
		println(c[1])
		println(c[2])
		println(c[3])

		println(len(a))
		println(len(c))
		`,
	)
	compileAndRun(program)

	// Output:
	// 1
	// 7
	// 8
	// 9
	// 1.500
	// 2.500
	// 3.500
	// 4.500
	// 4
	// 4
}

// ExampleArrayAppend2 checks, if the capacity is adjusted as expected (factor 2 for new allocations) and if we re-use
// existing memory, if it is available!
func ExampleArrayExtend2() {
	var program []byte = []byte(`
		a = [](int, 10)
		b = [](int, 5)

		a = extend(a, b)

		println(len(a))
		println(cap(a))

		// Memory reuse. Len should be 5*len(b), capacity should stay the same, as it is large enough!
		reset(a)
		a = extend(a, b)
		a = extend(a, b)
		a = extend(a, b)
		a = extend(a, b)
		a = extend(a, b)

		println(len(a))
		println(cap(a))

		a = extend(a, b)

		println(len(a))
		println(cap(a))

		// We now exeed the internal capacity and re-allocate more memory!
		a = extend(a, b)

		println(len(a))
		println(cap(a))

		`,
	)
	compileAndRun(program)

	// Output:
	// 15
	// 30
	// 25
	// 30
	// 30
	// 30
	// 35
	// 70
}

// ExampleArrayAppend2 checks, if the capacity is adjusted as expected (factor 2 for new allocations) and if we re-use
// existing memory, if it is available!
func ExampleArrayAppend() {
	var program []byte = []byte(`
		a = [](int, 10)

		println(cap(a))
		println(len(a))

		a = append(a, 8)

		println(cap(a))
		println(len(a))

		println(a[10])

		`,
	)
	compileAndRun(program)

	// Output:
	// 10
	// 10
	// 22
	// 11
	// 8
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
