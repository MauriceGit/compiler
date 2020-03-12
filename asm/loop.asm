; Simple example program to compute the
; sum of squares from 1 to n.
; **********************************************
; Data declarations

section     .data

; -----
; Define constants

SUCCESS         equ     3       ; Successful operation
SYS_exit        equ     60      ; call code for terminate

; Define Data.
n               dd      10
sumOfSquares    dq      0

; *******************************************************

section     .text
global      _start
_start:

; -----
; Compute sum of squares from 1 to n (inclusive).
; Approach:
; for (i=1; i<=n; i++)
;     sumOfSquares += i²;

    mov     rbx, 1      ;i
    mov     ecx, dword[n]
sumLoop:
    mov     rax, rbx    ; get i
    mul     rax         ; i²
    add     qword [sumOfSquares], rax
    inc     rbx
    loop    sumLoop

; -----
; Done, terminate program.
last:
    mov     rax, SYS_exit       ; call code for exit
    mov     rdi, SUCCESS        ; exit with success
    syscall

