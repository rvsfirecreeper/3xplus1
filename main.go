package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	var temp string
	valid := false
	var innum int
	var err error
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
		fmt.Println("Would you like single number mode or full mode. Single number mode is required for numbers above [your memory in bytes]/2000 for memory allocation reasons and goroutine reasons(s/f)")
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
		default:
			fmt.Println("Pick something valid, buckeroo")
		}
	}
	results := make([]int, innum+1)
	resultchannel := make(chan [2]int, innum+1)
	for num := 1; num <= innum; num++ {
		go func(n int) {
			resultchannel <- collatzcore(n)
		}(num)
	}
	for i := 1; i <= innum; i++ {
		result := <-resultchannel
		results[result[0]] = result[1]
	}
	for i := 1; i < len(results); i++ {
		fmt.Printf("%d took %d steps!\n", i, results[i])
	}
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
