#!/bin/bash

if [ "$#" -lt 2 ]; then
    echo "Usage: ./run_comparison.sh <file1> <file2> [<file3> ...]"
    exit 1
fi

python3 compare-results.py "$@"
