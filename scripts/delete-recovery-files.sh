#!/bin/bash

DIR="../volumes"

if [ -d "$DIR" ]; then
    rm -rf "$DIR"/*
else
    mkdir "$DIR"
fi

python3 create-id-gen-csv.py