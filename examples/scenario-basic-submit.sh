#!/bin/bash
# SCENARIO 1 - A simple submission of a job
# Configuration : 5 schedulers, 2 workers, 1 client

sleep 2
echo "START"

sleep 5
echo "SUBMIT examples/job-basic-hello.cpp"

sleep 1
echo "STATUS"

sleep 2
echo "STATUS 1-1"

sleep 1
echo "STOP"