#! /bin/bash

currentDir="${PWD}"

output="./html"

gitbook install $currentDir

gitbook serve $currentDir
#gitbook build $currentDir $output
