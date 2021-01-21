package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"../../pkg/codegen"
	"../../pkg/lexer"
	"../../pkg/parser"
)

func main() {
	/*
		Функція main() зчитує файл у тій самій папці з назвою vhid.py
		Далі створюються токени CreateTokens у пакеті pkg/lexer (lexer.go)
		Далі створюються ноди CreateNodes у пакеті pkg/parser (parser.go)
		Далі генерується MASM код у пакеты pkg/codegen (codegenerator.go)
	*/

	LabNum := "РГР" //change lab

	var path string

	if len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		path = fmt.Sprintf(`./%s-28-GO-IO-81-Shvachko.py`, LabNum)
	}
	f, err := os.Open(path) //OPEN FILE vhid.py in the same directive
	if err != nil {
		fmt.Println(`Програмі не вдалося знайти вхідний код.
		Щоб цього уникнути перейдіть в робочу директорію програми за допомогою команди cd або надайте другим аргументом шлях до файлу .py
		Приклад: 
		"C:\Users\User\Desktop\kpi\5 Семестр\Системне програмування 2\lab1\main.exe" "C:\Users\User\Desktop\kpi\5 Семестр\Системне програмування 2\lab1\vhid.py"`)
		fmt.Println(err.Error())
		return
	}
	/*Ця частина коду переписує у іншому кодуванні вхідний файл*/
	reader := bufio.NewReader(f)
	char, size, err := reader.ReadRune()
	var runes []rune
	for {
		//catch unintended errors
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == io.EOF {
			break
		}

		for i := 0; i < size; i++ {
			runes = append(runes, char)
		}
		char, size, err = reader.ReadRune()
	}
	/*Ця частина коду переписує у іншому кодуванні вхідний файл*/

	l := lexer.NewLexer()                   //init lexer
	tokens, errors := l.CreateTokens(runes) //create tokens
	if errors != nil {
		for _, v := range errors {
			fmt.Println(v)
		}
		return
	}
	for i, v := range tokens{
		fmt.Println(i+1, v)
	}

	p := new(parser.Parser) //create parser
	p.Tokens = tokens
	p.Lexer = l
	for i, v := range tokens {
		if v.Type ==`commentary` {
			tokens = append(tokens[:i], tokens[i+1:]...)
		}
	}
	Nodes, err := p.ProcessTokens(tokens) //create nodes
	if Nodes == nil || err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Nodes)
	
	code, err := codegen.GenerateMASM(Nodes) //generate code
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	f, err = os.Create(fmt.Sprintf(`./%s-28-GO-IO-81-Shvachko.asm`, LabNum))
	if err != nil {
		fmt.Println(err.Error())
	}
	f.Write([]byte(code))
}

/*
ЛР1: int + oct
ЛР2: not, multiplicate
ЛР3: div, left shift
ЛР4: ternary
ЛР5: func, <<=
ЛР6: while, minus
РГР: 
*/