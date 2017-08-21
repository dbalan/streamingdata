#!/bin/bash

set -e -x -o pipefail
go build


COUNT=5
CID=$( hexdump -n 4 /dev/urandom |head -1|tr -d ' ')


VAL=$(./recon_test -cid $CID -fail -count $COUNT -port 8000)

# server would explicitly fail, because storage is not fast and server holds a
# session lock
sleep 1s
VAL2=$(./recon_test -cid $CID -count $COUNT -port 8000)

if [[ -z $VAL ]]; then
    echo "error: empty first value"
    exit -1
fi

if [[ $VAL2 != $(echo $COUNT-1|bc) ]];then
    echo $VAL2
    echo $COUNT
    echo "counter mismatch, error"
fi

echo "PASS"

CID=$( hexdump -n 4 /dev/urandom |head -1|tr -d ' ')
COUNT=3
./recon_test -cid $CID -fail -count $COUNT -port 8000

sleep 40s
set +e
echo "this should panic"
./recon_test -cid $CID -fail -count $COUNT -port 8000
if [[ $? == 0  ]]; then
    echo "program executed properly when key should've expired, FAIL"
    exit -1
fi
set -e
echo PASS

