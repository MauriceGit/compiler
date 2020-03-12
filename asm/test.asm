; Console output

section     .data

; ----------------------------------------------------------------------
; Define constants

LF              equ     10      ; line feed
NULL            equ     0       ; end of string
TRUE            equ     1
FALSE           equ     0

EXIT_SUCCESS    equ     0       ; successful operation

STDIN           equ     0
STDOUT          equ     1
STDERR          equ     2

SYS_read        equ     0       ; read
SYS_write       equ     1       ; write
SYS_open        equ     2       ; file open
SYS_close       equ     3       ; file close
SYS_fork        equ     57      ; fork
SYS_exit        equ     60      ; terminate
SYS_create      equ     85      ; file open/create
SYS_time        equ     201     ; get time

; ----------------------------------------------------------------------
; Define variables/strings

message1        db      "Test string", LF, NULL
message2        db      "Enter answer:", NULL
newLine         db      LF, NULL

; ----------------------------------------------------------------------
section         .text
global _start
_start:

; ----------------------------------------------------------------------
; Display first message

    mov     rdi, message1
    call    printString

; ----------------------------------------------------------------------
; Display second message and newline

    mov     rdi, message2
    call    printString

    mov     rdi, newLine
    call    printString

; ----------------------------------------------------------------------
; Example program done.

exampleDone:
    mov     rax, SYS_exit
    mov     rdi, EXIT_SUCCESS
    syscall

; ----------------------------------------------------------------------
; Generic function to display a string to the screen.
; String must be NULL terminated.
; Algorithm:
;   Count characters in string (excluding NULL)
;   Use syscall to output characters
; Arguments:
;   1) address, string
; Returns:
;   nothing
global printString
printString:
    push        rbx

; ----------------------------------------------------------------------
; Count characters in string
    mov     rbx, rdi
    mov     rdx, 0
strCountLoop:
    cmp     byte [rbx], NULL
    je      strCountDone
    inc     rdx
    inc     rbx
    jmp     strCountLoop
strCountDone:
    cmp     rdx, 0
    je      prtDone

; ----------------------------------------------------------------------
; Call OS to output string
    mov     rax, SYS_write      ; system code for write()
    mov     rsi, rdi            ; address of chars to write
    mov     rdi, STDOUT         ; stdout. RDX=count to write, set above
    syscall

; ----------------------------------------------------------------------
; String printed, return to calling routine

prtDone:
    pop     rbx
    ret
