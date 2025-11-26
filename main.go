package main

import (
	"flag"    // Used for quiet mode
	"fmt"     // Used For Printing
	"os"      // Used for exit codes
	"runtime" // Used for worker count
	"runtime/pprof"
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

func collatzcore(seed int) collatz { // BRANCHLESS
	var i int
	current := seed
	for i = 0; current != 1; i++ {
		evenvalue := current >> 1    // Right Shift = divide by 2
		oddvalue := current*3 + 1    // CLASSIC COLLATZ
		evenmask := -(current&1 ^ 1) // Bitmask such that when its even it does nothing but when odd it cancels
		oddmask := -(current & 1)    // opposite of evenmask
		current = (evenvalue & evenmask) | (oddvalue & oddmask)
	}
	return collatz{seed, i}
}

var (
	quiet  = flag.Bool("quiet", false, "disable printing of individual results")
	record = flag.Bool("record", false, "disable printing of individual results, but print records")
)

func main() {
	f, _ := os.Create("cpu.prof")
	err := pprof.StartCPUProfile(f)
	if err != nil {
		fmt.Println("Aborting, pprof error.")
		os.Exit(1)
	}
	flag.Parse()
	if *quiet {
		fmt.Println("Shhhhh....")
	}
	const numJobs = 10000                           // Number of jobs before the channel is flushed out
	workers := runtime.GOMAXPROCS(runtime.NumCPU()) // Worker count
	var temp string                                 // Temporary variable used when taking input from terminal
	valid := false                                  // valid is used for input validatiom
	var end int                                     // maximum number to go up to
	var begin int                                   // minimum number to be calculated
	resultchannel := make(chan collatz, numJobs)    // Where the workers send the work
	results := make([]int, numJobs)
	batchnum := 0
	jobchan := make(chan int, numJobs*2)

	for !valid { // this entire thing validates an input
		fmt.Println("Pick a number, and this program will calculate a lot of Collatz Sequences.")
		_, err = fmt.Scanln(&temp)
		if err != nil {
			fmt.Println("Error receiving input, please try again.")
			continue
		}
		end, err = strconv.Atoi(temp)
		if err == nil && end >= 1 {
			valid = true
		} else {
			fmt.Println("Pick a valid psoitive integer.")
		}
	}
	valid = false
	for !valid { // this entire thing validates an input
		fmt.Println("Would you like single number mode, range mode, or full mode. Single number mode or range mode with a small range is required for very large numbers(s/r/f")
		_, err = fmt.Scanln(&temp)
		if err != nil {
			fmt.Println("Error receiving input, please try again.")
			continue
		}
		switch temp {
		case "s":
			fmt.Printf("%d took %d steps!\n", end, collatzcore(end).steps)
			os.Exit(0)
		case "f":
			valid = true
			begin = 0
		case "r":
			for !valid {
				fmt.Println("Where would you like to begin?")
				_, err = fmt.Scanln(&temp)
				if err != nil {
					fmt.Println("Error receiving input, please try again.")
					continue
				}
				valid = false
				begin, err = strconv.Atoi(temp)
				if err == nil && end-begin >= 0 {
					valid = true
					begin-- // Just prevents random off by one fixes everywhere
				} else {
					fmt.Println("Pick a valid integer less than the number you selected earlier.")
				}
			}
		default:
			fmt.Println("Pick a valid option, please.")
		}
	}
	end += 1

	fmt.Println("Initializing...")
	start := time.Now()
	for range workers {
		go collatzworker(jobchan, resultchannel)
	}
	fmt.Println("Now starting calculations!", end-begin)
	if *quiet { // i Know putting it here is ugly. But ugly code is sometimes fast code, and that's what matters more
		for num := begin + 1; num <= end; num++ {
			jobchan <- num
			if (num-begin)%numJobs == 0 {
				for range numJobs {
					<-resultchannel
				}
			}
		}
	} else if *record {
		recseq := collatz{steps: 0}
		for num := begin + 1; num <= end; num++ {
			jobchan <- num
			if (num-begin)%numJobs == 0 {
				for range numJobs {
					result := <-resultchannel
					if result.steps > recseq.steps {
						fmt.Printf("A new Record! %d broke the old record of %d steps with %d steps!", result.seed, recseq.steps, result.steps)
						recseq = result
					}
				}
			}
		}
	} else {
		for num := begin + 1; num <= end; num++ {
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
	}
	close(jobchan)
	if !*quiet {
		for i := 0; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			result := <-resultchannel
			results[(result.seed-begin)%numJobs] = result.steps
		}
		close(resultchannel)
		for i := 1; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			fmt.Println(batchnum*numJobs+i+begin, " took ", results[i], " steps to get to 1.")
		}
	} else {
		for i := 0; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			<-resultchannel
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("All %d Calculations done in %s!", end-begin-1, elapsed)
	defer pprof.StopCPUProfile()
}
