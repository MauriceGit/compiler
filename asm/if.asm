; Console output

section     .data

; ----------------------------------------------------------------------
; Define constants

SYS_exit        equ     60      ; call code for terminate

; ----------------------------------------------------------------------
; Define variables/strings

var1            dq      17
var2            dq      10

; ----------------------------------------------------------------------
section         .text
global _start
_start:

; ----------------------------------------------------------------------
; Integer comparison
; if (var1 < var2) {
;     return 5
; } else {
;     return 3
; }

    mov     rdi, 5              ; Set default return code before the whole comparison thing.
    mov     rax, qword [var1]
    cmp     rax, qword [var2]
    jl      lblSuccess          ; jump less, depending on the value in from cmp. The jump must be immediately after the cmp!
    mov     rdi, 3              ; rdi is the first parameter for a function. See page: 172
    jmp     lblFail

; ----------------------------------------------------------------------
;

lblSuccess:
    mov     rax, SYS_exit
    syscall

lblFail:
    mov     rax, SYS_exit
    syscall
