
# Compiler

Still incomplete compiler for a small little language into X86-64 Assembly.

Assembling and linking is currently done with yasm and ld.

## Current (incomplete) language grammar:

```
stat        ::= assign | if | for | funDecl | ret | funCall
statlist    ::= stat [statlist]

if          ::= 'if' exp '{' [statlist] '}' [else '{' [statlist] '}']
for         ::= 'for' [assign] ';' [explist] ';' [assign] '{' [statlist] '}'

type        ::= 'int' | 'float' | 'bool'
typelist    ::= type [',' typelist]
paramlist   ::= var type [',' paramlist]
funDecl     ::= 'fun' Name '(' [paramlist] ')' [typelist] '{' [statlist] '}'

ret         ::= 'return' [explist]

assign      ::= varlist ‘=’ explist | postIncr | postDecr
postIncr    ::= varDecl '++'
postDecr    ::= varDecl '--'
varlist     ::= varDecl [‘,’ varlist]
explist     ::= exp [‘,’ explist]
exp         ::= Numeral | String | var | '(' exp ')' | exp binop exp | unop exp | funCall
funCall     ::= Name '(' [explist] ')'
varDecl     ::= [shadow] var
var         ::= Name
binop       ::= '+' | '-' | '*' | '/' | '%' | '==' | '!=' | '<=' | '>=' | '<' | '>' | '&&' | '||'
unop        ::= '-' | '!'


Operator priority (Descending priority!):

0:  '-', '!'
1:  '*', '/', '%'
2:  '+', '-'
3:  '==', '!=', '<=', '>=', '<', '>'
4:  '&&', '||'
```

## Currently working example code:


```
fun sum(i int) int {
    if i == 0 {
        return 0
    }
    return i+sum(i-1)
}

fun abc(list []int) int {
    return list[2]
}

fun sumArray(list []int) int {
    s = 0
    for i = 0; i < 5; i++ {
        s += list[i]
    }
    return s
}

list = [1, 2, 3, 4, 5]
//list = [](int, 5)

printInt(sum(10))
printInt(abc(list))
printInt(sumArray(list))
```
