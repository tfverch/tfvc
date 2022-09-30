#! /bin/sh

cur=v0.1.1
next=$(svu next)

if [[ $cur != $next ]]; then
    echo "Hello"
fi