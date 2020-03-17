// lexer.go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func assemble(asm ASM, source, executable string) {

	var srcFile *os.File
	var err error

	if source == "" {
		srcFile, err = ioutil.TempFile("", "")
		if err != nil {
			fmt.Println("Creating temporary srcFile failed.")
			return
		}
		defer os.Remove(srcFile.Name())
	} else {
		srcFile, err = os.Create(source)
		if err != nil {
			fmt.Println("Creating srcFile failed.")
			return
		}
	}

	objectFile, err := ioutil.TempFile("", "")
	if err != nil {
		fmt.Println("Creating temporary objectFile failed.")
		return
	}
	objectFile.Close()

	// Write assembly into tmp source file
	defer os.Remove(objectFile.Name())

	for _, v := range asm.constants {
		fmt.Fprintln(srcFile, v)
	}
	for _, v := range asm.variables {
		fmt.Fprintln(srcFile, v)
	}
	for _, v := range asm.program {
		fmt.Fprintln(srcFile, v)
	}
	srcFile.Close()

	// Find yasm
	yasm, err := exec.LookPath("yasm")
	// Assemble
	yasmCmd := &exec.Cmd{
		Path:   yasm,
		Args:   []string{yasm, "-Worphan-labels", "-g", "dwarf2", "-f", "elf64", srcFile.Name(), "-o", objectFile.Name()},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := yasmCmd.Run(); err != nil {
		fmt.Println("Error while assembling the source code!")
		return
	}

	// Find ld
	ld, err := exec.LookPath("ld")
	// Link
	ldCmd := &exec.Cmd{
		Path:   ld,
		Args:   []string{ld, "-dynamic-linker", "/lib64/ld-linux-x86-64.so.2", "-o", executable, objectFile.Name(), "-lc"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := ldCmd.Run(); err != nil {
		fmt.Println("Error while linking the object file!")
		return
	}

}

func main() {
	var program []byte = []byte(`

value, b = 10 != 10, 5*5

`)

	//fmt.Println(string(program))

	tokenChan := make(chan Token, 1)
	go tokenize(program, tokenChan)

	ast, err := parse(tokenChan)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast, err = analyzeTypes(ast)
	if err != nil {
		fmt.Println(err)
		return
	}

	asm := ast.generateCode()

	assemble(asm, "source.asm", "executable")

}
