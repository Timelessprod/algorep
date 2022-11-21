#!/bin/bash
# SCENARIO 3 - Multiple submissions of jobs
# Configuration : 5 schedulers, 2 workers, 1 client

sleep 2
echo "START"

sleep 5
echo "SUBMIT examples/job-medium-prime.cpp"
sleep 2
echo "SUBMIT examples/job-basic-hello.cpp"
echo "SUBMIT examples/job-basic-factorial.cpp"
echo "SUBMIT examples/job-hard-prime.cpp"
sleep 2
echo "SUBMIT examples/job-basic-hello.cpp"
echo "SUBMIT examples/job-basic-factorial.cpp"
echo "SUBMIT examples/job-basic-hello.cpp"

echo "STATUS"

sleep 2
echo "STATUS 2-1"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 1
echo "STOP"