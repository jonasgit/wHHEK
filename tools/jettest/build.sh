#!/bin/sh
# Bygger JETTEST.EXE (32-bitars Windows-konsolprogram, Jet 4.0 finns bara i 32 bitar)
# med MinGW (i686-w64-mingw32-g++).
set -e
cd "$(dirname "$0")"
i686-w64-mingw32-g++ -O2 -Wall -static -o JETTEST.EXE jet_test.cpp -lole32 -loleaut32 -luuid -ladvapi32 -lodbc32
echo "byggde $(pwd)/JETTEST.EXE"
