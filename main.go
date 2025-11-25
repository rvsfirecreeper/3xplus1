package main

import (
	"fmt"     // Used For Printing
	"os"      // Used for exit codes
	"runtime" // Used for worker count
	"strconv" // Used for converting inputs to integers
	"time"    // Used for benchmarking
)

type collatz struct { // Struct for results
	seed  int
	steps int
}

func collatzworker(jobs <-chan int, resultchannel chan<- collatz) { // Defines the workers. If you're wondering how they're not slaves, they're paid in CPU Cycles
	for j := range jobs {
		resultchannel <- collatzcore(j)
	}
}

func collatzcore(seed int) collatz {
	var i int
	current := seed
	for i = 0; current != 1; i++ {
		if current%2 == 0 {
			current = current / 2
		} else {
			current = current*3 + 1
		}
	}
	return collatz{seed, i}
}

func main() {
	const numJobs = 10000                        // Number of jobs before the channel is flushed out
	workers := runtime.NumCPU() * 2              // Worker count
	var temp string                              // Temporary variable used when taking input from terminal
	valid := false                               // valid is used for input validatiom
	var innum int                                // maximum number to go up to
	var begin int                                // minimum number to be calculated
	var err error                                // err variable for input validation
	resultchannel := make(chan collatz, numJobs) // Where the workers send the work
	results := make([]int, numJobs)
	batchnum := 0
	jobchan := make(chan int, numJobs*2)

	for !valid { // this entire thing validates an input
		fmt.Println("Pick a number, we're gonna do some Collatz Wacky Stuff with it")
		_, err = fmt.Scanln(&temp)
		if err != nil {
			fmt.Println("That's an Error! Something went wrong")
			continue
		}
		innum, err = strconv.Atoi(temp)
		if err == nil && innum >= 1 {
			valid = true
		} else {
			fmt.Println("Pick something valid, buckeroo")
		}
	}
	valid = false
	for !valid { // this entire thing validates an input
		fmt.Println("Would you like single number mode, range mode, or full mode. Single number mode or range mode with a small range is required for very large numbers(s/r/f")
		_, err = fmt.Scanln(&temp)
		if err != nil {
			fmt.Println("Buckeroo How")
			continue
		}
		switch temp {
		case "s":
			fmt.Printf("%d took %d steps!\n", innum, collatzcore(innum).steps)
			os.Exit(0)
		case "f":
			valid = true
			begin = 0
		case "r":
			for !valid {
				fmt.Println("Where would you like to begin?")
				_, err = fmt.Scanln(&temp)
				if err != nil {
					fmt.Println("Buckeroo How")
					continue
				}
				valid = false
				begin, err = strconv.Atoi(temp)
				if err == nil && innum-begin >= 0 {
					valid = true
					begin-- // Just prevents random off by one fixes everywhere
				} else {
					fmt.Println("Pick. Something. Valid.")
				}
			}
		default:
			fmt.Println("Pick something valid, buckeroo")
		}
	}
	innum += 1

	fmt.Println("Starting Collatz Calculations!")
	start := time.Now()
	for range workers {
		go collatzworker(jobchan, resultchannel)
	}
	fmt.Println("Workers spawned! Now sending jobs", innum-begin)
	for num := begin + 1; num <= innum; num++ {
		jobchan <- num
		if (num-begin)%numJobs == 0 {

			for range numJobs {
				result := <-resultchannel
				results[(result.seed-begin)%numJobs] = result.steps
			}
			for i := 1; i < numJobs; i++ {
				fmt.Println(batchnum*numJobs+i+begin, " took ", results[i], " steps to get to 1.")
			}
			batchnum++
			fmt.Println(batchnum*numJobs+begin, " took ", results[0], " steps to get to 1.")

		}
	}
	close(jobchan)
	for i := 0; i < (innum-begin)%numJobs; i++ { // flush remaining numbers
		result := <-resultchannel
		results[(result.seed-begin)%numJobs] = result.steps
	}
	close(resultchannel)
	for i := 1; i < (innum-begin)%numJobs; i++ { // flush remaining numbers
		fmt.Println(batchnum*numJobs+i+begin, " took ", results[i], " steps to get to 1.")
	}
	elapsed := time.Since(start)
	fmt.Printf("All %d Calculations done in %s!", innum-begin-1, elapsed)
}
