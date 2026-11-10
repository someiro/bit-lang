/// bit-lang
/// Copyright (c) 2026 someiro All right reserved.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	var argCount int = len(os.Args)

	if argCount < 2 || argCount > 2 {
		fmt.Println(`
	    ██╗        ██╗    ██╗
	   ███████╗    ╚═╝ ████████╗
	   ██╔═══██╗  ██╗  ╚═██╔═══╝
	  ██╔╝  ██╔╝ ██╔╝   ██╔╝
	  ███████╔╝  ██║    ╚█████╗
	  ╚══════╝   ╚═╝     ╚════╝


            RUN || $ bit <fileName.bit>
         GITHUB || https://gihub.com/someiro/bit

     COPYRIGHT || Copyright (c) 2026 someiro All right reserved.
	 
	 `)
		os.Exit(0)
	}

	var fileName string = os.Args[1]
	var file, err = os.Open(fileName)

	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	defer file.Close()

	//var variables map[string]any = make(map[string]any)
	var blocks map[int]int = make(map[int]int)
	var stack []int = []int{}

	var scanner = bufio.NewScanner(file)
	var lineCount int = 0

	for scanner.Scan() {
		var line string = scanner.Text()
		var cmd string = strings.TrimSpace(line)
		lineCount++

		fmt.Printf("%d | %s\n", lineCount, cmd) // test

		if strings.HasPrefix(cmd, "if ") || strings.HasPrefix(cmd, "loop ") {
			stack = append(stack, lineCount)
			fmt.Println(stack) // test
		}

		if cmd == "." {
			var start int = 0
			if len(stack) > 0 {
				start = stack[len(stack)-1] 
				stack = stack[:len(stack)-1]
			}

			blocks[start] = lineCount
			fmt.Println(blocks) // test
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

}
