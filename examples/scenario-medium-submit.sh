#!/bin/bash
# SCENARIO 2 - A submission of a long job
# Configuration : 5 schedulers, 2 workers, 1 client

sleep 2
echo "START"

sleep 5
echo "SUBMIT examples/job-medium-prime.cpp"

sleep 1
echo "STATUS"

sleep 2
echo "STATUS 1-1"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 5
echo "STATUS"

sleep 2
echo "STATUS 1-1"

sleep 1
echo "STOP"