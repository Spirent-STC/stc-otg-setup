#!/usr/bin/bash

TESTBED="-testbed ../../testbed/b2b_1ate_1link.testbed"
BINDING="-binding ../../testbed/b2b_1ate_1link.binding"
OUTPUT="-outputs_dir ./fpLogs"

if [ -z "$1" ]; then
  CASES=`cat *.go | grep -o ' Test\w*'`
  CASES="$(grep -v "TestMain" <<< "$CASES")"
  echo Usage:
  echo   $0 all
  for i in $CASES; do
    echo "$0 $i"
  done
  exit
fi

if [ "$1" = "all" ]; then
  TESTCASE=
else
  TESTCASE="-test.run $1"
fi

if [ ! -d "fpLogs" ]; then
  mkdir fpLogs
fi

if [ -f *.test ]; then
  mv *.test test.exe
fi

echo -----------------------------------------------
echo -----------------------------------------------
echo -e Test `realpath . --relative-to=../` begin at: `date`
echo -----------------------------------------------
echo -----------------------------------------------

time ./test.exe $TESTBED $BINDING $OUTPUT $TESTCASE

echo -----------------------------------------------
echo -----------------------------------------------
echo -e Test end at: `date`
echo -----------------------------------------------
echo -----------------------------------------------