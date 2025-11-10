package main

import (
	"fmt"
	"os"
	"strconv"
)

func collatzworker(jobs <-chan int, resultchannel chan<- [2]int) { // Defines the workers. If you're wondering how they're not slaves, they're paid in CPU Cycles
	for j := range jobs {
		resultchannel <- collatzcore(j)
	}
}

func main() {
	const numJobs = 10000                       // Number of jobs before the channel is flushed out
	const workers = 10000                       // Worker count
	var temp string                             // Temporary variable used when taking input from terminal
	valid := false                              // valid is used for input validatiom
	var innum int                               // maximum number to go up to
	var begin int                               // minimum number to be calculated
	var err error                               // err variable for input validation
	resultchannel := make(chan [2]int, workers) // Where the workers send the work
	results := make([]int, numJobs)
	index := 0
	jobs := make(chan int, numJobs*2)

	for !valid { // this entire thing validates an input
		fmt.Println("Pick a number, we're gonna do some Collatz Wacky Stuff with it")
		_, err = fmt.Scanln(&temp)
		if err != nil {
			fmt.Println("That's an Error! Something went wrong")
			continue
		}
		innum, err = strconv.Atoi(temp)
		if err == nil {
			valid = true
		} else {
			fmt.Println("Pick something valid, buckeroo")
		}
	}
	innum += 1
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
			fmt.Printf("%d took %d steps!\n", innum, collatzcore(innum)[1])
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
				if err == nil {
					valid = true
					begin -= 1
				} else {
					fmt.Println("Pick. Something. Valid.")
				}
			}
		default:
			fmt.Println("Pick something valid, buckeroo")
		}
	}

	fmt.Print("\033[H\033[2J") // ANSI escape code to clear terminal
	fmt.Println("Starting Collatz Calculations!")
	fmt.Println("Spawning Workers...")
	for num := range workers {
		go collatzworker(jobs, resultchannel)
		_ = num // hacky way to get num as used
	}
	fmt.Println("Workers spawned! Now sending jobs", innum-begin)
	for num := 1; num <= innum-begin; num++ {
		jobs <- num
		if num%numJobs == 0 {
			index++

			for i := range numJobs {
				result := <-resultchannel
				results[result[0]%numJobs] = result[1]
				fmt.Println(index*numJobs+i, " took ", results[i], " steps to get to 1.")
			}

		}
	}

	for i := 0; i < (innum-begin)%numJobs; i++ { // flush remaining numbers
		result := <-resultchannel
		results[result[0]%numJobs] = result[1]
		fmt.Println(index*numJobs+i, " took ", results[i], " steps to get to 1.")
	}
	fmt.Println("All Calculations done!")
	close(jobs)
}

func collatzcore(seed int) [2]int {
	var i int
	current := seed
	for i = 0; current != 1; i++ {
		if current%2 == 0 {
			current = current / 2
		} else {
			current = current*3 + 1
		}
	}
	return [2]int{seed, i}
}
