package lexer

import (
	"errors"
	"fmt"
	"strings"
)

//Lexer ...
type Lexer struct {
	SpecialSymbols map[string]string
	Operators      *Operators
	KeyWords       map[string]string
	Values         *Values
	Tokens         []*Token
}

//NewLexer returns lexer filled object for python language
func NewLexer() *Lexer {
	l := new(Lexer)
	l.SpecialSymbols = map[string]string{
		` `: `ident`, `(`: `left parentesis`, `)`: `right parentesis`,
		`;`: `semicolon`, `:`: `double colon`, `,`: `comma`, `{`: `left tuple`,
		`}`: `right tuple`, `[`: `left array`, `]`: `right array`, `.`: `dot`, "\n": `whitespace`,
	}
	l.Operators = NewPythonOperators()
	l.KeyWords = map[string]string{
		`def`: `function`, `return`: `return value`,
		`for`: `for cycle`, `while`: `while cycle`,
	}
	l.Values = NewValues()

	return l
}

//CreateTokens returns token list
func (l *Lexer) CreateTokens(inputText []rune) ([]*Token, []error) {
	var tokens []*Token
	var errGlobal []error
	splittedText := strings.SplitAfter(string(inputText), "")
	//main loop for lexer
	var (
		currLine int = 1
		currOffset int = 0
	)

	for index := 0; index < len(inputText); index++ {
		var tokenExists bool = false
		symbol := string(inputText[index])

		currOffset++
		//check if token is number
		if IsNumber([]rune(symbol)[0]) || (index <= len(splittedText) && symbol == "." && IsNumber([]rune(splittedText[index+1])[0])) {
			i, number, err := CheckNumber(index, splittedText)
			length := i - index
			index = i
			if err != nil { //show error
				errGlobal = append(errGlobal, err)
				token := &Token{
					Type:  err.Error(),
					Value: number,
				}
				tokens = append(tokens, token)
				PrintError(i, splittedText, token)
				errGlobal = append(errGlobal, err)
				continue
			}
			token := &Token{
				Type:  "number",
				Value: number,
				Line: currLine,
				Offset: currOffset,
				Length: length,
			}
			tokens = append(tokens, token)
			continue

		}

		//check if token is string
		if symbol == `"` || symbol == `'` { //stringCheck
			i, name, err := ProcessString(index, splittedText)
			if err != nil {
				errGlobal = append(errGlobal, err)
				index = i
				token := &Token{
					Type:  err.Error(),
					Value: name,
				}
				tokens = append(tokens, token)

				PrintError(i, splittedText, token)
				continue
			}
			//index = i
			token := &Token{
				Type:  "string",
				Value: name,
			}
			PrintError(i, splittedText, token)
			return nil, []error{fmt.Errorf("Знайдений тип, що не можна привести до int")}
			tokens = append(tokens, token)
			continue
		}

		//check if token is one of special symbols
		for key, value := range l.SpecialSymbols {
			if key == symbol {
				token := &Token{
					Value: key,
					Type:  value,
					Line: currLine,
					Offset: currOffset,
					Length: 1,
				}
				if value == `whitespace` {
					currLine++
					currOffset = 0
				}
				tokens = append(tokens, token)
				tokenExists = true

				break
			}
		}
		if tokenExists {
			continue
		}

		//Check if token is one of basic operator symbol
		for _, value := range l.Operators.SimpleSymbol {
			if symbol == value {
				token, i := l.Operators.PredictOperator(index, splittedText)
				index = i
				tokens = append(tokens, token)
				tokenExists = true
				break
			}
		}
		if tokenExists {
			continue
		}
		//check if token is commentary
		if symbol == "#" {
			i, name := ProcessCommentary(index, splittedText)
			length := i - index
			index = i
			token := &Token{
				Type:  "commentary",
				Value: name,
				Line: currLine,
				Offset: currOffset,
				Length: length,
			}
			tokens = append(tokens, token)
			continue
		}

		if IsAlphabet(([]rune(symbol))[0]) {
			i, name := CheckName(index, splittedText)
			length := i - index
			index = i

			for key, value := range l.Operators.LogicalOperators { //check "AND OR "
				if name == key {
					token := &Token{
						Type:  value,
						Value: name,
						Line: currLine,
						Offset: currOffset,
						Length: length,
					}
					tokens = append(tokens, token)
					tokenExists = true
					break
				}
			}
			if tokenExists {
				continue
			}
			for key, value := range l.Operators.IdentityOperators { //check "IS"
				if name == key {
					token := &Token{
						Type:  value,
						Value: name,
						Line: currLine,
						Offset: currOffset,
						Length: length,
					}
					tokens = append(tokens, token)
					tokenExists = true
					break
				}
			}
			if tokenExists {
				continue
			}
			for key, value := range l.Operators.MembershipOperators { //CHECK IN
				if name == key {
					token := &Token{
						Type:  value,
						Value: name,
						Line: currLine,
						Offset: currOffset,
						Length: length,
					}
					tokens = append(tokens, token)
					tokenExists = true
					break
				}
			}
			if tokenExists {
				continue
			}

			for key, value := range l.KeyWords { //DEF FOR
				if name == key {
					token := &Token{
						Type:  value,
						Value: name,
						Line: currLine,
						Offset: currOffset,
						Length: length,
					}
					tokens = append(tokens, token)
					tokenExists = true
					break
				}
			}
			if tokenExists {
				continue
			}
			if name == "True" || name == "False" { //check boolean
				token := &Token{
					Type:  "Boolean",
					Value: name,
					Line: currLine,
						Offset: currOffset,
						Length: length,
				}
				tokens = append(tokens, token)
				continue
			}
			if name == "if" { //check boolean
				token := &Token{
					Type:  "IF",
					Value: name,
					Line: currLine,
					Offset: currOffset,
					Length: length,
				}
				tokens = append(tokens, token)
				continue
			}
			if name == "else" { //check boolean
				token := &Token{
					Type:  "ELSE",
					Value: name,
					Line: currLine,
					Offset: currOffset,
					Length: length,
				}
				tokens = append(tokens, token)
				continue
			}
			if name == "break" { //check boolean
				token := &Token{
					Type:  "break",
					Value: name,
					Line: currLine,
					Offset: currOffset,
					Length: length,
				}
				tokens = append(tokens, token)
				continue
			}
			if name == "continue" { //check boolean
				token := &Token{
					Type:  "continue",
					Value: name,
					Line: currLine,
					Offset: currOffset,
					Length: length,
				}
				tokens = append(tokens, token)
				continue
			}
			token := &Token{
				Type:  "Name",
				Value: name,
				Line: currLine,
				Offset: currOffset,
				Length: length,
			}
			tokens = append(tokens, token)
			continue
		}

	}
	return tokens, errGlobal
}

//Operators ...
type Operators struct {
	ArithmeticOperators map[string]string
	AssignmentOperators map[string]string
	ComprasionOperators map[string]string
	LogicalOperators    map[string]string
	IdentityOperators   map[string]string
	MembershipOperators map[string]string
	BitwiseOperators    map[string]string
	SimpleSymbol        []string
}

//NewPythonOperators ...
func NewPythonOperators() *Operators {
	o := new(Operators)
	o.ArithmeticOperators = map[string]string{
		`+`: `plus`, `-`: `minus`, `*`: `multiplicate`, `**`: `pow`,
		`/`: `divide`, `//`: `floor division`,
	}
	o.AssignmentOperators = map[string]string{
		`=`: `assign`, `+=`: `add and assign`, `-=`: `substract and assign`,
		`*=`: `mul and assign`, `/=`: `div and assign`, `%=`: `mod and assign`,
		`//=`: `floor div and assign`, `**=`: `pow and assign`, `&=`: `and and assign`,
		`|=`: `or and assign`, `^=`: `xor and assign`, `>>=`: `right shift and assign`,
		`<<=`: `left shift and assign`,
	}
	o.ComprasionOperators = map[string]string{
		`==`: `equal`, `!=`: `not equal`, `>`: `greater than`,
		`<`: `less than`, `>=`: `greater or equal`, `<=`: `less or equal`,
	}
	o.LogicalOperators = map[string]string{
		`and`: `logical and`, `or`: `logical or`, `not`: `logical not`,
	}
	o.IdentityOperators = map[string]string{
		`is`: `same object`, //`is not`: `different objects`,
	}
	o.MembershipOperators = map[string]string{
		`in`: `value in object`, //`not in`: `value not in object`,
	}
	o.BitwiseOperators = map[string]string{
		`&`: `and`, `|`: `or`, `^`: `xor`, `~`: `not`, `>>`: `right shift`, `<<`: `left shift`,
	}
	o.SimpleSymbol = []string{
		`+`, `-`, `=`, `>`, `<`, `*`, `/`, `&`, `|`, `^`, `~`, `!`,
	}
	return o
}

//PredictOperator ...
func (o *Operators) PredictOperator(index int, inputText []string) (*Token, int) {
	var token = new(Token)
	ThreeSymbol := inputText[index : index+3]
	ThreeSymbolString := strings.Join(ThreeSymbol, "")

	if ThreeSymbol[len(ThreeSymbol)-1] == "=" {
		for key, value := range o.AssignmentOperators {
			if ThreeSymbolString == key {
				token = &Token{
					Type:  value,
					Value: key,
				}
				return token, index + len(key) - 1
			}
		}
	}

	//TWO SYMBOLS
	TwoSymbol := inputText[index : index+2]
	TwoSymbolString := strings.Join(TwoSymbol, "")

	for key, value := range o.ArithmeticOperators {
		if key == TwoSymbolString {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	for key, value := range o.AssignmentOperators {
		if key == TwoSymbolString {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	for key, value := range o.ComprasionOperators {
		if key == TwoSymbolString {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	for key, value := range o.BitwiseOperators {
		if key == TwoSymbolString {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	//ONE SYMBOL

	for key, value := range o.ArithmeticOperators {
		if key == inputText[index] {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	for key, value := range o.AssignmentOperators {
		if key == inputText[index] {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	for key, value := range o.ComprasionOperators {
		if key == inputText[index] {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	for key, value := range o.BitwiseOperators {
		if key == inputText[index] {
			token = &Token{
				Type:  value,
				Value: key,
			}
			return token, index + len(key) - 1
		}
	}

	token = &Token{
		Type:  "Unexpected token",
		Value: inputText[index],
	}
	return token, index //?
}

//Values is a struct of python values
type Values struct {
	//NoneType           string
	BooleanType string
	IntType     string
	FloatType   string
	//	ComplexType        string
	//SequenceType       string //arrays
	TextSequenceType string
	//	BinarySequenceType string
	//	SetTypes           string
	//	MapTypes           string
}

//NewValues ...
func NewValues() *Values {
	v := new(Values)
	v.BooleanType = `BoolVar`
	v.IntType = `IntVar`
	v.FloatType = `FloatVar`
	v.TextSequenceType = `StringType`

	return v
}

//Token ...
type Token struct {
	Type   string
	Value  interface{}
	Line   int
	Offset int
	Length int
	//tokenstart
	//tokenend
	//
}

func (t *Token) String() string {
	return fmt.Sprintf(`%v: '%v'`, t.Type, t.Value)
}

//IsAlphabet checks if symbol is from latins alphabet
func IsAlphabet(c rune) bool {
	var cIsAChar bool = false
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
		cIsAChar = true
	}
	return cIsAChar
}

//IsNumber checks if number is from numerics
func IsNumber(c rune) bool {
	var cIsANum bool = false
	if c >= '0' && c <= '9' {
		cIsANum = true
	}
	return cIsANum
}

//CheckNumber checks if entity is a number
func CheckNumber(index int, inputText []string) (int, string, error) {
	var c string = inputText[index]
	var Shift int
	var findOctal bool
	for i, v := range inputText[index+1:] {
		if IsNumber([]rune(v)[0]) || v == "." || v == "o" || v == "O" {
			if v == "o" || v == "O" {
				findOctal = true
			}
			c += v
			continue
		}
		Shift = i
		break
	}

	splittedNumber := strings.Split(c, "")

	if findOctal { //Make checks if number is Octal
		if splittedNumber[0] == "0" && (splittedNumber[1] == "o" || splittedNumber[1] == "O") {
			countOs := 0
			for _, v := range splittedNumber {
				if v == "o" || v == "O" {
					countOs++
				} else if !(v >= "0" && v <= "7") {
					return index + Shift, c, errors.New("Bad octal number")
				}
			}
			if countOs > 1 {
				return index + Shift, c, errors.New("Bad octal number")
			}
		} else {
			return index + Shift, c, errors.New("Bad octal number")
		}

	}
	countDots := 0
	for _, v := range splittedNumber {
		if v == "." {
			countDots++
		}
	}
	if countDots > 1 {
		return index + Shift, c, errors.New("Bad float number")
	}

	return index + Shift, c, nil
}

//CheckName ...
func CheckName(index int, inputText []string) (int, string) {
	var name string = inputText[index]
	var Shift int
	for i, v := range inputText[index+1:] {
		if IsAlphabet([]rune(v)[0]) || IsNumber([]rune(v)[0]) || v == "_" {
			name += v
			continue
		}
		Shift = i
		break
	}
	return index + Shift, name
}

//ProcessCommentary ...
func ProcessCommentary(index int, inputText []string) (int, string) {
	var comment string
	var Shift int
	for i, v := range inputText[index:] {
		if v == "\n" {
			break
		}
		comment += v
		Shift = i
	}
	return index + Shift, comment
}

//ProcessString ...
func ProcessString(index int, inputText []string) (int, string, error) {
	var str string = inputText[index]
	var Shift int
	for i, v := range inputText[index+1:] {
		if v == inputText[index] {
			str += v
			break
		} else if v == "\n" {
			return index + Shift, str, errors.New("Invalid string")
		}
		str += v
		Shift = i
	}
	return index + Shift + 2, str, nil
}

//PrintError prints syntax errors
func PrintError(index int, inputText []string, invalidToken *Token) {
	var line int = 1 //count lines while not find error
	var ind int = 0  //count index of error

	for i, v := range inputText {
		if v == "\n" {
			line++
			ind = i
		}
		if i == index {
			break
		}
	}
	fmt.Println(fmt.Sprintf("Syntax error: invalid token \n%s\nLine: %d, Index: %d", invalidToken, line, index-ind))
}
