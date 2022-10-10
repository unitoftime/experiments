#!/bin/bash

mkdir -p results

# Note: ecs-slow is go only
#    for collisions in 0 10; do
for collisions in 0 1 2 3 4 5 6 7 8 9 10 20 30 40 50 60 70 80 90 100; do
#    for size in 1000 5000; do
    for size in 1000 2000 3000 4000 5000 6000 7000 8000 9000 10000 20000 30000 40000 50000 60000 70000 80000 90000 100000 1000000; do
        for program in "native" "nativeSplit" "ecs" "ecs-slow"; do
#        for program in "nativeSplit" ; do
            # Rust Release Benchmark
            file=results/rust/release/${program}/${size}_${collisions}.txt
            echo ${file}
            mkdir -p results/rust/release/${program}/
            ./rust-bin/release/rust_benchmarks ${program} ${size} ${collisions} &> ${file}
            echo DONE
            sleep 10;

            # Rust Debug Benchmark
            # Warning: This is incredibly slow
#            file=results/rust/debug/${program}/${size}_${collisions}.txt
#            echo $file
#            mkdir -p results/rust/debug/${program}/
#            ./rust-bin/debug/rust_benchmarks ${program} ${size} ${collisions} &> ${file}
#            echo DONE
#            sleep 10;

            # Go Benchmark
            file=results/go/release/${program}/${size}_${collisions}.txt
            echo ${file}
            mkdir -p results/go/release/${program}/
            ./go-bin/bench ${program} ${size} ${collisions} &> ${file}
            echo DONE
            sleep 10;
        done
    done
done
