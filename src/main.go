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
	argCount := len(os.Args)

	if argCount < 2 || argCount > 2 {
		fmt.Println(`
	    ██╗        ██╗    ██╗
	   ███████╗    ╚═╝ ████████╗
	   ██╔═══██╗  ██╗  ╚═██╔═══╝
	  ██╔╝  ██╔╝ ██╔╝   ██╔╝
	  ███████╔╝  ██║    ╚█████╗
	  ╚══════╝   ╚═╝     ╚════╝
 
 
            RUN || $ bit <fileName.bit>
         GITHUB || https://github.com/someiro/bit
 
     COPYRIGHT || Copyright (c) 2026 someiro All right reserved.
	 
	 `)
		os.Exit(0)
	}

	fileName := os.Args[1]
	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	defer file.Close()

	variables := make(map[string]any)
	blocks := make(map[int]int)
	elses := make(map[int]int)
	stack := []int{}
	lines := []string{}

	elseEnds := make(map[int]int)
	loopStarts := make(map[int]int)
	funcs := make(map[string]int)
	funcEnds := make(map[int]bool)
	kinds := make(map[int]string)
	callStack := []int{}

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		cmd := strings.TrimSpace(line)

		lines = append(lines, cmd)
		lineCount++

		// fmt.Printf("%d | %s\n", lineCount, cmd) // test

		if strings.HasPrefix(cmd, "if ") || strings.HasPrefix(cmd, "for ") || strings.HasPrefix(cmd, "func ") {
			stack = append(stack, lineCount)
			//fmt.Println(stack) // test
		}

		if strings.HasPrefix(cmd, "if ") || strings.HasPrefix(cmd, "for ") || strings.HasPrefix(cmd, "func ") {
			kinds[lineCount] = strings.SplitN(cmd, " ", 2)[0]
		}

		if strings.HasPrefix(cmd, "func ") {
			funcName := strings.TrimSpace(cmd[5:])
			if funcName == "" {
				fmt.Println("Function name required:", cmd)
				os.Exit(1)
			}
			funcs[funcName] = lineCount
		}

		if cmd == "else" {
			if len(stack) == 0 {
				fmt.Println("Unexpected else")
				os.Exit(1)
			}

			start := stack[len(stack)-1]
			elses[start] = lineCount

			// fmt.Println(elses) // test
		}

		if cmd == "else" && kinds[stack[len(stack)-1]] != "if" {
			fmt.Println("else without if at line:", lineCount)
			os.Exit(1)
		}

		if cmd == "." {
			start := 0
			if len(stack) > 0 {
				start = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}

			blocks[start] = lineCount
			//fmt.Println(blocks) // test

			switch kinds[start] {
			case "if":
				// elseがあれば「elseの行 → この . の行」を記録
				if elseLine, ok := elses[start]; ok {
					elseEnds[elseLine] = lineCount
				}
			case "for":
				// 「この . の行 → for の行」を記録
				loopStarts[lineCount] = start
			case "func":
				// この . は関数の終わり
				funcEnds[lineCount] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	for i := 0; i < len(lines); i++ {
		cmd := lines[i]

		if strings.HasPrefix(cmd, "if ") {
			end, ok := blocks[i+1]

			if !ok {
				fmt.Println("Block not closed:", cmd)
				os.Exit(1)
			}

			expr := strings.TrimSpace(cmd[3:])
			value, err := eval(expr, variables)

			if err != nil {
				fmt.Println("Error evaluating condition:", err)
				os.Exit(1)
			}

			condition, ok := value.(bool)

			if !ok {
				fmt.Println("Condition must be boolean:", expr)
				os.Exit(1)
			}

			elseLine, hasElse := elses[i+1]

			if !condition {
				if hasElse {
					i = elseLine - 1
				} else {
					i = end - 1
				}
			}

			continue

		}

		if strings.HasPrefix(cmd, "print ") {
			expr := strings.TrimSpace(cmd[6:])

			value, err := eval(expr, variables)

			if err != nil {
				fmt.Println("Error evaluating expression:", err)
				os.Exit(1)
			}
			fmt.Println(value)
		}

		if strings.HasPrefix(cmd, "@ ") {
			// @ 以降の文字列を取得し、"="で分割する
			parts := strings.SplitN(cmd[2:], "=", 2)

			// fmt.Println(parts) // test

			if len(parts) != 2 {
				fmt.Println("Invalid variable assignment:", cmd)
				os.Exit(1)
			}

			name := strings.TrimSpace(parts[0])
			expr := strings.TrimSpace(parts[1])

			if name == "" || expr == "" {
				fmt.Println("Invalid variable assignment:", cmd)
				os.Exit(1)
			}

			var value any

			if strings.HasPrefix(expr, "input ") {
				prompt := strings.TrimSpace(expr[6:])

				message, err := eval(prompt, variables)

				if err != nil {
					fmt.Println("Error evaluating input prompt:", err)
					os.Exit(1)
				}

				text, ok := message.(string)

				if !ok {
					fmt.Println("Input prompt must be a string")
					os.Exit(1)
				}

				value, err = readInput(text)

				if err != nil {
					fmt.Println("Input error:", err)
					return
				}
			} else {
				value, err = eval(expr, variables)

				if err != nil {
					fmt.Println("Error evaluating expression:", err)
					os.Exit(1)
				}
			}

			if err != nil {
				fmt.Println("Error evaluating expression:", err)
				os.Exit(1)
			}

			variables[name] = value

			// fmt.Printf("Variable '%s' set to: %v\n", name, value) // test
		}

		if strings.HasPrefix(cmd, "$ ") {
			// $ 以降の文字列を取得し、"="で分割する
			parts := strings.SplitN(cmd[2:], "=", 2)

			// fmt.Println(parts) // test

			if len(parts) != 2 {
				fmt.Println("Invalid variable assignment:", cmd)
				os.Exit(1)
			}

			name := strings.TrimSpace(parts[0])
			expr := strings.TrimSpace(parts[1])

			if name == "" || expr == "" {
				fmt.Println("Invalid variable assignment:", cmd)
				os.Exit(1)
			}

			_, ok := variables[name]
			if !ok {
				fmt.Println("Variable not defined:", name)
				os.Exit(1)
			}

			var value any

			if strings.HasPrefix(expr, "input ") {
				prompt := strings.TrimSpace(expr[6:])

				message, err := eval(prompt, variables)

				if err != nil {
					fmt.Println("Error evaluating input prompt:", err)
					os.Exit(1)
				}

				text, ok := message.(string)

				if !ok {
					fmt.Println("Input prompt must be a string")
					os.Exit(1)
				}

				value, err = readInput(text)

				if err != nil {
					fmt.Println("Input error:", err)
					return
				}
			} else {
				value, err = eval(expr, variables)

				if err != nil {
					fmt.Println("Error evaluating expression:", err)
					os.Exit(1)
				}
			}
			if err != nil {
				fmt.Println("Error evaluating expression:", err)
				os.Exit(1)
			}

			variables[name] = value

			// fmt.Printf("Variable '%s' updated to: %v\n", name, value) // test
		}

		// i を書き換えたら、ループの i++ で「飛び先の次の行」から再開
		// else に着いた = ifがtrueで中身を実行し終わった。残りのelseは飛ばして . の次へ
		if cmd == "else" {
			i = elseEnds[i+1] - 1
		}

		// for: 条件がfalseなら . の次へ抜ける。trueならそのまま中身へ
		if strings.HasPrefix(cmd, "for ") {
			end, ok := blocks[i+1]

			if !ok {
				fmt.Println("Block not closed:", cmd)
				os.Exit(1)
			}

			expr := strings.TrimSpace(cmd[4:])
			value, err := eval(expr, variables)

			if err != nil {
				fmt.Println("Error evaluating condition:", err)
				os.Exit(1)
			}

			condition, ok := value.(bool)

			if !ok {
				fmt.Println("Condition must be boolean:", expr)
				os.Exit(1)
			}

			if !condition {
				i = end - 1
			}
		}

		// func: 定義の行。呼ばれたときだけ実行したいので、中身を飛ばして . の次へ
		if strings.HasPrefix(cmd, "func ") {
			end, ok := blocks[i+1]

			if !ok {
				fmt.Println("Block not closed:", cmd)
				os.Exit(1)
			}

			i = end - 1
		}

		// call: 今の行を積んでから、関数の1行目へ飛ぶ
		if strings.HasPrefix(cmd, "call ") {
			funcName := strings.TrimSpace(cmd[5:])
			funcLine, ok := funcs[funcName]

			if !ok {
				fmt.Println("Function not defined:", funcName)
				os.Exit(1)
			}

			callStack = append(callStack, i)
			i = funcLine - 1
		}

		// . に着いたとき: forなら for の行に戻る。関数の終わりなら call の次の行に戻る
		if cmd == "." {
			if start, ok := loopStarts[i+1]; ok {
				i = start - 2
			} else if funcEnds[i+1] && len(callStack) > 0 {
				last := len(callStack) - 1
				i = callStack[last]
				callStack = callStack[:last]
			}
		}

	}

}
