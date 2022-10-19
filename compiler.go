// lexer.go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

func getConvenienceFunctions() []byte {

	// the print and println functions for strings can be written in code instead of assembly
	return []byte(`
		fun print(s string) {
		    for i, c : s {
		        print(c)
		    }
		}

		fun println(s string) {
		    print(s)
		    print(char(10))
		}

		fun print(b bool) {
		    if b {
		        print("true")
		    } else {
		        print("false")
		    }
		}

		fun println(b bool) {
		    print(b)
		    print(char(10))
		}
	`)
}

func preprocess(program []byte) (out []byte) {

	//out = append(getConvenienceFunctions(), program...)

	out = program

	out = bytes.ReplaceAll(out, []byte("string"), []byte("[]char"))

	re := regexp.MustCompile(`(\".*\")`)
	out = re.ReplaceAllFunc(out, func(s []byte) []byte {
		tmpS := string(s)
		if tmpS == "\"\"" {
			return []byte("[](char, 0)")
		}
		return []byte("['" + strings.Join(strings.Split(tmpS[1:][:len(tmpS)-2], ""), "','") + "']")
	})

	return
}

func compile(program []byte, sourceFile, binFile string) bool {
	tokenChan := make(chan Token, 1)
	lexerErr := make(chan error, 1)

	program = preprocess(program)

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
		fun test () int {
		    for ;; {
		        break
		        return 0
		    }
			return 1
		}
		fun test (x int) int {
		    switch x {
				case 1:
					return 1
				case 3:
					return 32
				default:
					return 3
			}
		}
	`)

	if len(os.Args) > 1 {
		var err error
		program, err = ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if !compile(program, "source.asm", "executable") {
		os.Exit(1)
	}
}
