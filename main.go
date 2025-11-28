package main

import (
	"fmt"     // Used For Printing
	"os"      // Used for exit codes
	"runtime" // Used for worker count
	"runtime/pprof"
	"strconv" // Used for converting inputs to integers
	"time"    // Used for benchmarking

	flag "github.com/spf13/pflag" // Used for quiet mode
)

type collatz struct { // Struct for results
	seed  int
	steps int
}

type intconstraint struct {
	min   *int
	equal *int
	max   *int
}

func mkpoint[T any](v T) *T {
	return &v
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
	quiet  = flag.BoolP("quiet", "q", false, "disable printing of individual results")
	record = flag.BoolP("record", "r", false, "disable printing of individual results, but print records")
)

func intInput(prompt string, conditions intconstraint) int {
	valid := false
	var flag bool
	var temp string
	var err error
	var num int
	for !valid {
		flag = false
		fmt.Println(prompt)
		_, err = fmt.Scanln(&temp)
		if err != nil {
			fmt.Println("Hmm, you seem to have entered an invalid value.")
			fmt.Println("Hint: Cannot be blank")
			continue
		}
		num, err = strconv.Atoi(temp)
		if err != nil {
			fmt.Println("Hmm, you seem to have entered an invalid value.")
			fmt.Println("Hint: Must be integer")
			continue
		}
		if conditions.equal != nil && num != *conditions.equal {
			fmt.Println("Hmm, you seem to have entered an invalid value.")
			fmt.Printf("Hint: Must be equal to %d\n", *conditions.equal)
			flag = true
		}
		if conditions.max != nil && num > *conditions.max {
			fmt.Println("Hmm, you seem to have entered an invalid value.")
			fmt.Printf("Hint: Must be less than %d\n", *conditions.max+1)
			flag = true
		}
		if conditions.min != nil && num < *conditions.min {
			fmt.Println("Hmm, you seem to have entered an invalid value.")
			fmt.Printf("Hint: Must be more than %d\n", *conditions.min-1)
			flag = true
		}
		if flag {
			continue
		}
		valid = true
	}
	return num
}

func main() {
	flag.Parse()
	if *quiet && *record {
		fmt.Fprintln(os.Stderr, "Error: quiet and record mode are conflicting.")
		os.Exit(1)
	}
	f, _ := os.Create("cpu.prof")
	if *quiet {
		fmt.Println("Shhhhh.... it's quiet mode.")
	}
	if *record {
		fmt.Println("Who will be the best? It's record mode! ")
	}
	const numJobs = 10000                           // Number of jobs before the channel is flushed out
	workers := runtime.GOMAXPROCS(runtime.NumCPU()) // Worker count
	var end int                                     // maximum number to go up to
	var begin int                                   // minimum number to be calculated
	var recseq collatz
	var result collatz
	resultchannel := make(chan collatz, numJobs) // Where the workers send the work
	results := make([]int, numJobs)
	batchnum := 0
	jobchan := make(chan int, numJobs*2)
	end = intInput("Pick a number, and this program will calculate a lot of Collatz Sequences.", intconstraint{min: mkpoint(1)})
	begin = intInput("Would you like single number mode, range mode, or full mode.(1 for single, 2 for full, 3 for range)", intconstraint{min: mkpoint(1), max: mkpoint(3)})
	switch begin {
	case 1:
		fmt.Printf("%d took %d steps!\n", end, collatzcore(end).steps)
		os.Exit(0)
	case 2:
		begin = 1
	case 3:
		begin = intInput("Where would you like to begin?", intconstraint{min: mkpoint(1), max: &end})

		end-- // Fixes bug where one last number tries to send after the resultchannel closes.
	}
	end += 1
	fmt.Println("Initializing...")
	start := time.Now()
	err := pprof.StartCPUProfile(f)
	if err != nil {
		fmt.Println("Aborting, pprof error.")
		os.Exit(1)
	}
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
		for num := begin + 1; num <= end; num++ {
			jobchan <- num
			if (num-begin)%numJobs == 0 {
				for range numJobs {
					result = <-resultchannel
					if result.steps > recseq.steps {
						fmt.Printf("A new Record! %d broke the old record of %d steps with %d steps!\n", result.seed, recseq.steps, result.steps)
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
					result = <-resultchannel
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
	if *quiet {
		for i := 1; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			<-resultchannel
		}
	} else if *record {
		for i := 1; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			result = <-resultchannel
			if result.steps > recseq.steps {
				fmt.Printf("A new Record! %d broke the old record of %d steps with %d steps!\n", result.seed, recseq.steps, result.steps)
				recseq = result
			}
		}
	} else {
		for i := 1; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			result = <-resultchannel
			results[(result.seed-begin)%numJobs] = result.steps
		}
		for i := 1; i < (end-begin)%numJobs; i++ { // flush remaining numbers
			fmt.Println(batchnum*numJobs+i+begin, " took ", results[i], " steps to get to 1.")
		}
	}
	close(resultchannel)
	elapsed := time.Since(start)
	fmt.Printf("All %d Calculations done in %s!", end-begin, elapsed)
	defer pprof.StopCPUProfile()
}
