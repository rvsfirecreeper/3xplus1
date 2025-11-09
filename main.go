package main

import (
	"fmt"
	"os"
	"strconv"
)

func collatzworker(jobs <-chan int, resultchannel chan<- [2]int) {
	for j := range jobs {
		resultchannel <- collatzcore(j)
	}
}

func main() {
	const numJobs = 10000
	const workers = 5000
	var temp string
	valid := false
	var innum int
	var begin int
	var err error
	jobs := make(chan int, numJobs*2)
	for !valid {
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
	valid = false
	for !valid {
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

	fmt.Print("\033[H\033[2J")
	fmt.Println("Starting Collatz Calculations!")
	resultchannel := make(chan [2]int, workers)
	for num := 1; num <= workers; num++ {
		go collatzworker(jobs, resultchannel)
	}

	for num := 1; num <= innum-begin; num++ {
		jobs <- num
		if num%numJobs == 0 {
			for i := 1; i <= numJobs; i++ {
				result := <-resultchannel
				fmt.Println(result[0], " took ", result[1], " steps to make it to 1")
			}
		}
	}
	for len(resultchannel) != 0 {
		result := <-resultchannel
		fmt.Println(result[0], " took ", result[1], " steps to make it to 1")
	}
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
