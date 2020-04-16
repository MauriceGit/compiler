#!/usr/bin/env bash
set -e
basename="${1%.*}"
extension="${1##*.}"
yasm -f elf64 "$basename.$extension"
ld -o "$basename" "$basename.o"
rm -f "$basename.o"
"./$basename"
