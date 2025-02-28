#!/bin/bash

# set parrallel clients number
NUM_CLIENTS=1

PIDS=()

trap "echo 'Stopping all clients...'; kill 0; exit 1" SIGINT SIGTERM

for i in $(seq 1 $NUM_CLIENTS); do
    echo "Starting client $i..."
    ./client -mode c -case s -i 1000000 -userIndex $i&
    PIDS+=($!)
done

echo "All clients started!"
wait
