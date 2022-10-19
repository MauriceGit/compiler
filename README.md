
# Compiler

This project is a small compiler, that compiles my own little language into X86-64 Assembly.
It then uses `yasm` and `ld` to assemble and link into a Linux X86-64 executable.

## But why?

I've always wanted to write a compiler myself! But just never got around to do it. So the current Coronavirus quarantine
situation finally gives me enough time to tackle it myself.

And I was really impressed by people, that wrote solutions for last years [adventofcode.com](https://adventofcode.com)
in their own language. So that is something I'd like to achieve :)

I like to work on challenging problems and found compilers intriguing.

## How to run

- `go build`
- `./compiler <source_file>`
- `./executable`

The compiler will always create an executable called `executable`. Additionally, it will create a file `source.asm`
that contains the generated (not optimized) assembly.

## Dependencies

Everything is written from scratch, there are no code dependencies.
But to assemble and link the program into an executable, you need:
- `yasm`
- `ld`

The resulting Assembly also has no external dependencies (No C std lib, printing is implemented in Assembly directly).

## Language influences

- Go
- C
- Python
- Lua

## Features

- Strong and static type system
- Multiple return values from functions
- Automatic packing/unpacking of multiple arguments and/or function returns
- Function overloading
- Function inlining (only for system functions right now)
- Dynamic Arrays with an internal capacity, so not every `append` needs a new memory allocation
- Int, Char, String and Float types are always 64bit
- Very Python-like array creation
- Switch expressions match either values or general boolean expressions
- Range-based Loops with index and element
- Structs

## Examples

See the `compiler_test.go` file for a lot more working examples :)

### Print
```C
// There are overloaded functions: print, println that work on int, float, char, bool and string
println(5)
println(6.543)
println('b')
println("abc")
println(true)
```

### Assignment
```C
// Types are derived from the expressions!
a = 4
b = 5.6
c = true
d = 'f'
e = "abc"
f = a + 5 + int(b) // 14
```

### Functions
```C
fun abc(i int, j float) int, float, char, string {
    return i, j, 'b', "abc"
}
// Can be overloaded
fun abc(i int, j int) int, float, char, string {
    return i, 5.5, 'a', "abc"
}
// ...
a, b, c, d = abc(5, 6.5)
```

### Lists
```C
// List of integers. Type derived from the expressions. Every values must have the exact same type.
list = [1, 2, 3, 4, 5]
// Strings can be explicitely defined as a list of chars or implicitely by using string literals
// Internally, they are handled the same way.
chars = ['a', 'b', 'c', 'd', 'e', 'f', 'g']
str = "abcdefg"

// Empty list of integers with length 10. Type explicitely set
list2 = [](int, 10)
// Lists can naturally contain other lists
list3 = [list, list2]
// There are build-in functions, to get the length and capacity
println(len(list3))
println(cap(list3))

// You have to free them yourself
free(list)

// And some convenience functions to clear/reset the list without deallocating the memory:
// reset only resets the length, while clear overwrites the memory with 0s
reset(list)
clear(list)

// Build-in append/extend function, similar to the one in Go
// Careful ! append/extend works on the first argument. Depending on the available capacity, it will
// extend list or free list and create a completely new memory block, copy list and list2 over and return
// the new pointer!
list = append(list, 6)
list = extend(list, list2)

// Lists in functions/structs
fun abc(a []int) {
    // ...
}
```

### Loops
```C
list = [1,2,3,4,5]

for i = 0; i < len(list); i++ {
    // ...
}

for i,e : list {
    // i is the current index
    // e is the actual element: list[i]
}

for i, c : "abcdefg" {
    print(c)
}
// will print in: abcdefg
```

### Conditions
```C
a = true
b = 3 > 4
if a && !b {
    println("abc")
}
```

### Switch
```C
// We have no fallthrough
switch 4 {
case 1:
    println("case 1")
case 2, 3, 6:
    println("case 2")
case 5:
    println("case 3")
default:
    println("default case")
}

// switches do not need a value to check against.
// With no value provided, each case will be checked after each other for an expression that evaluates to true.
switch {
case 2 > 3:
    println("case 1")
case 2 == 3:
    println("case 2")
default:
    println("default case")
}
```

### Structs
```C
struct B {
    i int
    j int
}
struct A {
    i int
    j B
}

// Structs are created by calling a function with the same name and an exact match of parameters
// that match the expected types of the struct.
// Internally, this is just syntax, not a function. So there is no overhead!
a = A(1, B(3, 4))
a.j.j = 100
println(a.i)
println(a.j.j)
```

### Type conversions
```C
// Build-in (inline) functions: int(), float()
println(int(5.5))
println(float(5))

// Chars can be converted from and to int (ascii value) by using char() or int()
println(int('a'))
println(char(98))
```
