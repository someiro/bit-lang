package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

func evalNode(node ast.Expr, variables map[string]any) (any, error) {
	switch n := node.(type) {
	// 数値や文字列などの基本リテラルを評価する
	case *ast.BasicLit:
		// リテラルの種類に応じて値を返す
		switch n.Kind {
		case token.INT:
			return strconv.ParseInt(n.Value, 0, 64)

		// 文字列リテラルの場合
		case token.STRING:
			// 文字列リテラルの場合は、引用符を取り除いて返す
			return strconv.Unquote(n.Value)

		// 浮動小数点リテラルの場合
		case token.FLOAT:
			return strconv.ParseFloat(n.Value, 64)

		// その他のリテラルの場合
		default:
			return nil, fmt.Errorf("unsupported literal type: %v", n.Kind)
		}

	// 変数と真偽値を評価する
	case *ast.Ident:
		// 変数名が "true" または "false" の場合は、それぞれ true または false を返す
		if n.Name == "true" {
			return true, nil
		}
		if n.Name == "false" {
			return false, nil
		}

		// 変数名が "nil" の場合は nil を返す
		if n.Name == "nil" {
			return nil, nil
		}

		// 変数の値を取得する
		value, ok := variables[n.Name]

		// 変数が定義されていない場合はエラーを返す
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", n.Name)
		}

		// 変数の値を返す
		return value, nil

	// 括弧
	case *ast.ParenExpr:
		// 括弧内の式を評価する
		return evalNode(n.X, variables)

	// 単項演算子
	case *ast.UnaryExpr:
		// 単項演算子のオペランドを評価する
		value, err := evalNode(n.X, variables)
		// オペランドの評価に失敗した場合はエラーを返す
		if err != nil {
			return nil, err
		}

		// 単項演算子の種類に応じて値を返す
		switch n.Op {
		// 負の数
		case token.SUB:
			// オペランドが整数であることを確認する
			switch number := value.(type) {
			case int64:
				return -number, nil
			case float64:
				return -number, nil
			default:
				return nil, fmt.Errorf("invalid operand for -")
			}

			// 真偽値の否定
		case token.NOT:
			// オペランドが真偽値であることを確認する
			boolean, ok := value.(bool)

			// オペランドが真偽値でない場合はエラーを返す
			if !ok {
				return nil, fmt.Errorf("invalid operand for !")
			}

			// 真偽値の否定を返す
			return !boolean, nil

		// その他の単項演算子はサポートされていないため、エラーを返す
		default:
			return nil, fmt.Errorf("unsupported unary operator: %v", n.Op)
		}

	// 二項演算子
	case *ast.BinaryExpr:
		// 左辺の評価
		left, err := evalNode(n.X, variables)
		// 左辺の評価に失敗した場合はエラーを返す
		if err != nil {
			return nil, err
		}
		// 右辺の評価
		right, err := evalNode(n.Y, variables)
		// 右辺の評価に失敗した場合はエラーを返す
		if err != nil {
			return nil, err
		}

		// 左辺が整数であることを確認する
		l, lok := left.(int64)
		// 右辺が整数であることを確認する
		r, rok := right.(int64)

		// 整数演算
		if lok && rok {
			// 二項演算子の種類に応じて値を返す
			switch n.Op {
			// 加算
			case token.ADD:
				return l + r, nil
			// 減算
			case token.SUB:
				return l - r, nil
			// 乗算
			case token.MUL:
				return l * r, nil
			// 除算
			case token.QUO:
				// 除算の際にゼロ除算をチェックする
				if r == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				// 除算の結果を返す
				return l / r, nil
			// 剰余
			case token.REM:
				// 剰余の際にゼロ除算をチェックする
				if r == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				// 剰余の結果を返す
				return l % r, nil
			// 比較演算
			case token.EQL:
				// 等価比較の結果を返す
				return l == r, nil
			// 不等価比較の結果を返す
			case token.NEQ:
				// 不等価比較の結果を返す
				return l != r, nil
			// 小なり
			case token.LSS:
				// 小なりの結果を返す
				return l < r, nil
			// 小なりイコール
			case token.LEQ:
				// 小なりイコールの結果を返す
				return l <= r, nil
			// 大なり
			case token.GTR:
				// 大なりの結果を返す
				return l > r, nil
			// 大なりイコール
			case token.GEQ:
				// 大なりイコールの結果を返す
				return l >= r, nil
			// その他の演算子はサポートされていないため、エラーを返す
			default:
				// サポートされていない二項演算子の場合はエラーを返す
				return nil, fmt.Errorf("invalid operation: %v", n.Op)
			}
		}

		// 浮動小数点数演算
		lf, lfok := left.(float64)
		rf, rfok := right.(float64)

		if lfok && rfok {
			switch n.Op {
			case token.ADD:
				return lf + rf, nil
			case token.SUB:
				return lf - rf, nil
			case token.MUL:
				return lf * rf, nil
			case token.QUO:
				if rf == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lf / rf, nil
			case token.EQL:
				return lf == rf, nil
			case token.NEQ:
				return lf != rf, nil
			case token.LSS:
				return lf < rf, nil
			case token.LEQ:
				return lf <= rf, nil
			case token.GTR:
				return lf > rf, nil
			case token.GEQ:
				return lf >= rf, nil
			default:
				return nil, fmt.Errorf("invalid operation: %v", n.Op)
			}
		}

		// 整数と小数の混合演算
		if (lok || lfok) && (rok || rfok) {
			// 整数を小数に変換する
			if lok {
				lf = float64(l)
			}

			if rok {
				rf = float64(r)
			}

			switch n.Op {
			case token.ADD:
				return lf + rf, nil
			case token.SUB:
				return lf - rf, nil
			case token.MUL:
				return lf * rf, nil
			case token.QUO:
				if rf == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lf / rf, nil
			case token.EQL:
				return lf == rf, nil
			case token.NEQ:
				return lf != rf, nil
			case token.LSS:
				return lf < rf, nil
			case token.LEQ:
				return lf <= rf, nil
			case token.GTR:
				return lf > rf, nil
			case token.GEQ:
				return lf >= rf, nil
			default:
				return nil, fmt.Errorf("invalid operation: %v", n.Op)
			}
		}

		// 左辺が文字列であることを確認する
		ls, lsok := left.(string)
		// 右辺が文字列であることを確認する
		rs, rsok := right.(string)

		// 文字列演算
		// 文字列演算
		if lsok || rsok {
			switch n.Op {
			case token.ADD:
				return fmt.Sprint(left) + fmt.Sprint(right), nil

			case token.EQL:
				return lsok && rsok && ls == rs, nil

			case token.NEQ:
				return !(lsok && rsok && ls == rs), nil

			default:
				return nil, fmt.Errorf("invalid operation: %v", n.Op)
			}
		}

		// 真偽値演算
		lb, lbok := left.(bool)
		rb, rbok := right.(bool)

		if lbok && rbok {
			switch n.Op {
			case token.LAND:
				return lb && rb, nil
			case token.LOR:
				return lb || rb, nil
			case token.EQL:
				return lb == rb, nil
			case token.NEQ:
				return lb != rb, nil
			default:
				return nil, fmt.Errorf("invalid operation: %v", n.Op)
			}
		}

		return nil, fmt.Errorf("invalid operation: %v", n.Op)

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", n)
	}
}

func eval(expr string, variables map[string]any) (any, error) {
	node, err := parser.ParseExpr(expr)

	if err != nil {
		return nil, err
	}

	// fmt.Println(node) // test

	// ast.Print(token.NewFileSet(), node) // test

	return evalNode(node, variables)
}

func parseInput(value string) any {
	value = strings.TrimSpace(value)

	// 整数
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return n
	}

	// 小数
	if n, err := strconv.ParseFloat(value, 64); err == nil {
		return n
	}

	// 真偽値
	if value == "true" {
		return true
	}

	if value == "false" {
		return false
	}

	// どれにも該当しなければ文字列
	return value
}

var inputReader = bufio.NewReader(os.Stdin)

func readInput(prompt string) (any, error) {
	fmt.Print(prompt)
	text, _ := inputReader.ReadString('\n')
	return parseInput(text), nil
}
