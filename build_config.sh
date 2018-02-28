#!/bin/bash

OUTPUT=$1
PWD=$(pwd)

if test -z "${OUTPUT}"; then
    echo "Usage: $0 <output_file>"
    exit 1
fi

rm -rf ${OUTPUT}
touch ${OUTPUT}

if [ ! -d "bin" ]; then
    $(mkdir bin)
fi

echo "export GOPATH=${GOPATH}:${PWD}" >> ${OUTPUT}
