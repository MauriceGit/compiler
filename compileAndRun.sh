#!/usr/bin/env bash
set -e
basename="${1%.*}"
extension="${1##*.}"
yasm -Worphan-labels -g dwarf2 -f elf64 "$basename.$extension"
ld -dynamic-linker /lib64/ld-linux-x86-64.so.2 -o "$basename" "$basename.o" -lc
rm -f "$basename.o"
"./$basename"
