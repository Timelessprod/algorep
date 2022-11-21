#!/bin/bash
# SCENARIO 4 - Multiple submissions of jobs with crash and recover
# Configuration : 5 schedulers, 2 workers, 1 client

sleep 2
echo "START"

sleep 5
echo "SUBMIT examples/job-medium-prime.cpp"
echo "SUBMIT examples/job-basic-hello.cpp"
sleep 3
echo "SUBMIT examples/job-basic-hello.cpp"
echo "SUBMIT examples/job-basic-factorial.cpp"
echo "SUBMIT examples/job-hard-prime.cpp"

sleep 1
echo "STATUS"

sleep 1
echo "CRASH 4"

sleep 1
echo "STATUS"

sleep 1
echo "SUBMIT examples/job-basic-hello.cpp"
echo "SUBMIT examples/job-basic-factorial.cpp"
echo "SUBMIT examples/job-basic-hello.cpp"

sleep 1
echo "CRASH 3"

sleep 1
echo "SUBMIT examples/job-basic-hello.cpp"
echo "SUBMIT examples/job-medium-prime.cpp"
echo "SUBMIT examples/job-basic-hello.cpp"

sleep 1
echo "STATUS"

sleep 1
echo "RECOVER 4"

sleep 1
echo "STATUS"

sleep 1
echo "SUBMIT examples/job-basic-hello.cpp"
echo "SUBMIT examples/job-medium-prime.cpp"
echo "SUBMIT examples/job-basic-hello.cpp"

sleep 1
echo "RECOVER 3"

sleep 2
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

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 1
echo "STOP"