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
	// 555.66600
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
	// 1.50000
	// 2.50000
	// 3.50000
	// 4.50000
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

// ExampleRangedFor checks, that the ranged-for loop works fine on more complex examples, such as arrays in arrays
func ExampleRangedFor() {
	var program []byte = []byte(`
		a = [1,2,3,4,5]
		b = [[1], [1,2,3], a]

	  	for i, aa : b {
			for j, e : aa {
				println(e)
			}
		}
		`,
	)
	compileAndRun(program)

	// Output:
	// 1
	// 1
	// 2
	// 3
	// 1
	// 2
	// 3
	// 4
	// 5
}

// ExampleBreakContinue checks, that break and continue statements work as expected.
func ExampleBreakContinue() {
	var program []byte = []byte(`
		for i, e : [1,2,3,4,5,6,7,8,9,10] {
			if e == 3 {
				continue
			}
			if i >= 6 {
				break
			}
			println(e)
		}
		`,
	)
	compileAndRun(program)

	// Output:
	// 1
	// 2
	// 4
	// 5
	// 6
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
		fun sumI(a int, b int, c int, d int, e int, f int, g int, h int) int {
			return a+b+c+d+e+f+g+h
		}
		println(sumI(1,2,3,4,5,6,7,8))

		fun sumF(a float, b float, c float, d float, e float, f float, g float, h float, i float, j float, k float) float {
			return a+b+c+d+e+f+g+h+i+j+k
		}
		println(sumF(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0))
		`,
	)
	compileAndRun(program)

	// Output:
	// 36
	// 66.00000
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
	// 5.50000
	// 5.50000
	// 6.50000
}

func ExampleTypeCast() {
	var program []byte = []byte(`
		f = 5.123
	   	println(f)
		println(int(f))
		println(float(int(f)))
		println(int(float(int(f))))
		`,
	)
	compileAndRun(program)

	// Output:
	// 5.12300
	// 5
	// 5.00000
	// 5
}

// ExamplePrintFloat exists because we had an issue with incorrect printing of some float values. Just to make sure
// it works from now on.
func ExamplePrintFloat() {
	var program []byte = []byte(`
	   	println(5.1)
	   	println(0.0123)
		`,
	)
	compileAndRun(program)

	// Output:
	// 5.10000
	// 0.01230
}

// ExampleSwitch checks the two variants of switch statements. Value-match switch and general switch.
func ExampleSwitch() {
	var program []byte = []byte(`
	   	a = 4

		switch 4 {
		case 1:
			println(1)
		case 2, 3, a:
			println(4)
		case 5:
			println(5)
		}

		switch {
		case 7 < 5:
			println(1)
		case false, 4 > 5, a < 6:
			println(35)
		case 6 < 7:
			println(6)
		}
		`,
	)
	compileAndRun(program)

	// Output:
	// 4
	// 35
}

// ExampleSwitch2 checks the default cases for both switches
func ExampleSwitch2() {
	var program []byte = []byte(`
	   	switch 4 {
		case 1:
			println(1)
		case 2, 3, 6:
			println(4)
		case 5:
			println(5)
		case:
			println(999)
		}

		switch {
		case 2 > 3:
			println(1)
		case 2 == 3:
			println(4)
		default:
			println(888)
		}
		`,
	)
	compileAndRun(program)

	// Output:
	// 999
	// 888
}

// ExampleStruct1 checks, that structs work in general, within other structs, and within arrays.
func ExampleStruct1() {
	var program []byte = []byte(`
	   	struct Blubb2 {
			i int
			j int
		}
		struct Blubb {
			i int
			j Blubb2
		}

		a = Blubb(1, Blubb2(3, 4))
		b = [](Blubb, 5)

		b[0] = Blubb(6, Blubb2(7, 8))
		b[1] = a
		b[2] = Blubb(9, Blubb2(10, 11))

		a.j.j = 100

		println(b[1].i)
		println(b[1].j.j)

		println(a.i)
		println(a.j.j)
		`,
	)
	compileAndRun(program)

	// Output:
	// 1
	// 4
	// 1
	// 100
}

// ExampleStruct1 checks, that also work with arrays inside
func ExampleStruct2() {
	var program []byte = []byte(`
	   	struct Blubb2 {
			i int
			j int
		}
		struct Blubb {
			i int
			j []Blubb2
		}

        b = Blubb2(1,2)
		array = [b, Blubb2(5,6), b]

		array[1].j = 234

		println(array[1].i)
		println(array[1].j)
		`,
	)
	compileAndRun(program)

	// Output:
	// 5
	// 234
}
