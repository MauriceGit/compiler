// lexer.go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func assemble(asm ASM, source, executable string) (err error) {

	var srcFile *os.File
	var e error

	if source == "" {
		srcFile, e = ioutil.TempFile("", "")
		if e != nil {
			err = fmt.Errorf("Creating temporary srcFile failed - %w", e)
			return
		}
		defer os.Remove(srcFile.Name())
	} else {
		srcFile, e = os.Create(source)
		if e != nil {
			err = fmt.Errorf("Creating srcFile failed - %w", e)
			return
		}
	}

	objectFile, e := ioutil.TempFile("", "")
	if e != nil {
		err = fmt.Errorf("Creating temporary objectFile failed - %w", e)
		return
	}
	objectFile.Close()

	// Write assembly into tmp source file
	defer os.Remove(objectFile.Name())

	for _, v := range asm.header {
		fmt.Fprintf(srcFile, "%v\n", v)
	}
	for k, v := range asm.constants {
		fmt.Fprintf(srcFile, "%-12v%-10v%-15v\n", v, "equ", k)
	}
	for k, v := range asm.sysConstants {
		fmt.Fprintf(srcFile, "%-12v%-10v%-15v\n", v, "equ", k)
	}
	for _, v := range asm.variables {
		fmt.Fprintf(srcFile, "%-12v%-10v%-15v\n", v[0], v[1], v[2])
	}
	for _, v := range asm.sectionText {
		fmt.Fprintf(srcFile, "%v\n", v)
	}
	for k, f := range asm.functions {
		if !f.inline && f.used {
			fmt.Fprintf(srcFile, "global %v\n", k)
			fmt.Fprintf(srcFile, "%v:\n", k)

			for _, v := range f.code {
				fmt.Fprintf(srcFile, "%v%-10v%-10v\n", v[0], v[1], v[2])
			}
		}
	}
	for _, v := range asm.program {
		fmt.Fprintf(srcFile, "%v%-10v%-10v\n", v[0], v[1], v[2])
	}
	srcFile.Close()

	// Find yasm
	yasm, e := exec.LookPath("yasm")
	if e != nil {
		err = fmt.Errorf("'yasm' not found. Please install - %w", e)
		return
	}
	// Assemble
	yasmCmd := &exec.Cmd{
		Path:   yasm,
		Args:   []string{yasm, "-f", "elf64", srcFile.Name(), "-o", objectFile.Name()},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if e := yasmCmd.Run(); e != nil {
		err = fmt.Errorf("Error while assembling the source code - %w", e)
		return
	}

	// Find ld
	ld, e := exec.LookPath("ld")
	if e != nil {
		err = fmt.Errorf("'ld' not found. Please install - %w", e)
		return
	}
	// Link
	ldCmd := &exec.Cmd{
		Path:   ld,
		Args:   []string{ld, "-o", executable, objectFile.Name()},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if e := ldCmd.Run(); e != nil {
		err = fmt.Errorf("Error while linking object file - %w", e)
		return
	}
	return
}

func compile(program []byte, sourceFile, binFile string) bool {
	tokenChan := make(chan Token, 1)
	lexerErr := make(chan error, 1)
	go tokenize(program, tokenChan, lexerErr)

	ast, parseErr := parse(tokenChan)

	// check error channel on incoming errors
	// As we lex and parse simultaneously, there is most likely a parser error as well. But that should be ignored
	// as long as we have token errors before!
	select {
	case e := <-lexerErr:
		fmt.Println(e)
		return false
	default:
	}

	if parseErr != nil {
		fmt.Println(parseErr)
		return false
	}

	ast, semanticErr := semanticAnalysis(ast)
	if semanticErr != nil {
		fmt.Println(semanticErr)
		return false
	}

	asm := ast.generateCode()

	if asmErr := assemble(asm, sourceFile, binFile); asmErr != nil {
		fmt.Println(asmErr)
		return false
	}
	return true
}

func main() {
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

		b[2] = Blubb(9, Blubb2(10, 11))
		b[1] = a

		a.j.j = 100

		println(b[0].i)
		println(b[0].j.j)

		println(a.i)
		println(a.j.j)




	`)

	if !compile(program, "source.asm", "executable") {
		os.Exit(1)
	}
}
