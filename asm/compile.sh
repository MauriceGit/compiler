#!/usr/bin/env bash
set -e
basename="${1%.*}"
extension="${1##*.}"
yasm -Worphan-labels -g dwarf2 -f elf64 "$basename.$extension"
ld -g -o "$basename" "$basename.o"
rm -f "$basename.o"
