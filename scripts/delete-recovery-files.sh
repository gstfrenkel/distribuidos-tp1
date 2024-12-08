#!/bin/bash

DIR="../volumes"

if [ -d "$DIR" ]; then
    rm -rf "$DIR"/*
else
    echo "$DIR does not exist."
fi
