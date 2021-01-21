package parser

import (
	"fmt"
	"strings"

	"../lexer"
)

//Parser class
type Parser struct {
	Tokens []*lexer.Token
	Lexer  *lexer.Lexer
}

//Node smallest of AST
type Node struct {
	Name     string
	Options  map[string]interface{}
	Children []*Node
	Token    *lexer.Token
}

func (n *Node) String() string {
	Children := fmt.Sprintf("%v", n.Children)
	NewChildren := strings.ReplaceAll(Children, "\n", "\n|\t")
	return fmt.Sprintf("\n%v: {\n|\tOptions: %v\n|\tChildren: %v\n|\t}", n.Name, n.Options, NewChildren)
}

func reverse(nodes []*Node) []*Node {
	newNodes := make([]*Node, 0, len(nodes))
	for i := len(nodes) - 1; i >= 0; i-- {
		newNodes = append(newNodes, nodes[i])
	}
	return newNodes
}

//Functions name: number of parameters
var Functions map[string]int = make(map[string]int)

//ProcessTokens ...
func (p *Parser) ProcessTokens(tokens []*lexer.Token) ([]*Node, error) {
	var AST []*Node
	iterator := NewIterator()
	iterator.Tokens = tokens

	var (
		functionNode *Node
		err          error
	)

	for iterator.nextNotIdentAndWhiteSpace().Type == `function` {
		functionNode, err = p.ProcessFunction(iterator)
		if err != nil {
			return nil, err
		}
		AST = append(AST, functionNode)

	}

	return AST, nil
}

//ProcessFunction ...
func (p *Parser) ProcessFunction(iterator *Iterator) (*Node, error) {

	if AllEquals(iterator.manypeek(6), []string{`function`, `Name`, `left parentesis`, `right parentesis`, `double colon`, `whitespace`}) {
		iterator.nextToken() //skip def
		var functionNode *Node = &Node{
			Name: `function`,
			Options: map[string]interface{}{
				`Name`: iterator.nextNotIdentAndWhiteSpace().Value,
			},
			Token: iterator.Tokens[iterator.n],
		}
		//add function to memory for using it in codegen
		Functions[functionNode.Options[`Name`].(string)]++
		iterator.findEndOfLine()
		var countSpaces int = 1
		if iterator.Tokens[iterator.n].Type != `ident` { //if no idents present
			return nil, fmt.Errorf("Missing function body at line %d", iterator.Tokens[iterator.n].Line)
		}
		for iterator.peek().Type == `ident` {
			iterator.nextToken()
			countSpaces++
		} //it will stops on last ident

		child, err := p.ProcessStatement(iterator, countSpaces)
		if err != nil {
			return nil, err
		}
		functionNode.Children = append(functionNode.Children, child)
		iterator.findEndOfLine()
		for iterator.countIdents() == countSpaces {
			iterator.movePointer(countSpaces - 1) //move to the last ident
			if iterator.peek().Type == `ident` {
				return nil, fmt.Errorf("Error in fuction body. Expected %d spaces\nLine %d", countSpaces, iterator.Tokens[iterator.n].Line)
			}
			child, err := p.ProcessStatement(iterator, countSpaces)
			if err != nil {
				return nil, err
			}
			functionNode.Children = append(functionNode.Children, child)
			iterator.findEndOfLine()
		}
		return functionNode, nil
	}
	//func with parameters
	if AllEquals(iterator.manypeek(4), []string{`function`, `Name`, `left parentesis`, `Name`}) {
		iterator.nextToken() //skip def

		var functionNode *Node = &Node{
			Name: `function`,
			Options: map[string]interface{}{
				`Name`: iterator.nextNotIdentAndWhiteSpace().Value,
			},
			Token: iterator.Tokens[iterator.n],
		}
		//add function to memory for using it in codegen
		Functions[functionNode.Options[`Name`].(string)]++

		iterator.nextNotIdentAndWhiteSpace() //go to left parentesis
		iterator.nextNotIdentAndWhiteSpace() //and skip
		if AllEquals(iterator.manypeek(2), []string{`Name`, `right parentesis`}) {
			parametrNode := &Node{
				Name: `parameter`,
				Options: map[string]interface{}{
					`Name`: iterator.Tokens[iterator.n].Value,
				},
				Token: iterator.Tokens[iterator.n],
			}
			functionNode.Children = append(functionNode.Children, parametrNode)
		} else if AllEquals(iterator.manypeek(2), []string{`Name`, `comma`}) {
			iterator.n-- //move pointer to left parentesis
			for {
				if iterator.peekNotIdent().Type == `Name` {
					parametrNode := &Node{
						Name: `parameter`,
						Options: map[string]interface{}{
							`Name`: iterator.nextNotIdentAndWhiteSpace().Value,
						},
						Token: iterator.Tokens[iterator.n],
					}
					//add function to memory for using it in codegen
					Functions[functionNode.Options[`Name`].(string)]++

					functionNode.Children = append(functionNode.Children, parametrNode)
					if iterator.peekNotIdent().Type != `comma` && iterator.peekNotIdent().Type != `right parentesis` {
						return nil, fmt.Errorf("Wrong function parametres at line %d index %d\n%s", iterator.Tokens[iterator.n+1].Line, iterator.Tokens[iterator.n+1].Offset, iterator.Tokens[iterator.n+1])
					}
					continue
				} else if iterator.peekNotIdent().Type == `comma` {
					iterator.nextNotIdentAndWhiteSpace() //move to comma
					if iterator.peekNotIdent().Type != `Name` {
						return nil, fmt.Errorf("Expected name after line %d index %d\n%s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
					}
					continue
				} else if iterator.peekNotIdent().Type == `right parentesis` {
					iterator.nextNotIdentAndWhiteSpace() //move to right parentesis
					if AllEquals(iterator.manypeek(3), []string{`right parentesis`, `double colon`, `whitespace`}) {
						break
					}
					return nil, fmt.Errorf("Wrong function declaration start at line %d", iterator.Tokens[iterator.n].Line)
				} else {
					return nil, fmt.Errorf("Expected argument at line %d index %d\n%s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
				}
			}
		}
		iterator.findEndOfLine()

		var countSpaces int = 1
		if iterator.Tokens[iterator.n].Type != `ident` { //if no idents present
			return nil, fmt.Errorf("Missing function body at line %d", iterator.Tokens[iterator.n].Line)
		}
		for iterator.peek().Type == `ident` {
			iterator.nextToken()
			countSpaces++
		} //it will stops on last ident

		child, err := p.ProcessStatement(iterator, countSpaces)
		if err != nil {
			return nil, err
		}
		functionNode.Children = append(functionNode.Children, child)
		iterator.findEndOfLine()
		for iterator.countIdents() == countSpaces {
			iterator.movePointer(countSpaces - 1) //move to the last ident
			if iterator.peek().Type == `ident` {
				return nil, fmt.Errorf("Error in fuction body. Expected %d spaces\nLine %d", countSpaces, iterator.Tokens[iterator.n].Line)
			}
			child, err := p.ProcessStatement(iterator, countSpaces)
			if err != nil {
				return nil, err
			}
			functionNode.Children = append(functionNode.Children, child)
			iterator.findEndOfLine()
		}
		return functionNode, nil
	}
	
	return nil, fmt.Errorf("Wrong function declaration start at line %d", iterator.Tokens[iterator.n].Line)

}

//ProcessStatement ...
func (p *Parser) ProcessStatement(iterator *Iterator, countSpaces int) (*Node, error) {
	if iterator.peek().Type == `while cycle` {
		whileNode := &Node{
			Name:  iterator.nextToken().Type,
			Token: iterator.Tokens[iterator.n],
		}
		child, err := p.ProcessExpression(iterator)
		if err != nil {
			return nil, err
		}
		whileNode.Children = append(whileNode.Children, child) //append condition
		if iterator.peek().Type == `double colon` {
			iterator.nextToken()//move to double collon
			if iterator.peekNotIdent().Type != `whitespace` && iterator.Tokens[iterator.n].Type != `whitespace`{
				return nil, fmt.Errorf("Unexpected token after line: %d, index: %d '%s'.\nExpected new line", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
			}
			iterator.findEndOfLine()//go to the next line
			iterator.movePointer(countSpaces)//go to the last function space

			var countWhileSpaces int = 1
			if iterator.peek().Type != `ident` { //if no idents present
				return nil, fmt.Errorf("Missing while body at line %d", iterator.Tokens[iterator.n].Line)
			}
			for iterator.peek().Type == `ident` {
				iterator.nextToken()
				countWhileSpaces++
			} //it will stops on last ident
			child, err := p.ProcessStatement(iterator, countWhileSpaces)//get first statement
			if err != nil {
				return nil, err
			}
			whileNode.Children = append(whileNode.Children, child)
			iterator.findEndOfLine()
			for iterator.countIdents() == countWhileSpaces+countSpaces {
				iterator.movePointer(countWhileSpaces+countSpaces-1) //move to the last ident
				if iterator.peek().Type == `ident` {
					return nil, fmt.Errorf("Error in fuction body. Expected %d spaces\nLine %d", countSpaces+countWhileSpaces, iterator.Tokens[iterator.n].Line)
				}
				child, err := p.ProcessStatement(iterator, countSpaces+countWhileSpaces)
				if err != nil {
					return nil, err
				}
				whileNode.Children = append(whileNode.Children, child)
				iterator.findEndOfLine()
			}
			iterator.movePointer(-1) //move from the newline back for function processor ok working
			return whileNode, nil
		}//if double collon is missing
		return nil, 
			fmt.Errorf("Unexpected token after line: %d, index: %d Name: '%s'.\nExpected :", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
		
	}
	if iterator.peek().Type == `return value` {
		returnNode := &Node{
			Name:  iterator.nextToken().Type,
			Token: iterator.Tokens[iterator.n],
		}
		child, err := p.ProcessExpression(iterator)
		if err != nil {
			return nil, err
		}
		returnNode.Children = append(returnNode.Children, child)
		return returnNode, nil
	}

	expressionNode, err := p.ProcessExpression(iterator)
	if err != nil {
		return nil, err
	}
	return expressionNode, nil
}

//ProcessExpression ...
func (p *Parser) ProcessExpression(iterator *Iterator) (*Node, error) {
	if AllEquals(iterator.manypeek(2), []string{`Name`, `assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `add and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `substract and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `mul and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `div and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `mod and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `floor div and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `pow and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `and and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `or and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `xor and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `right shift and assign`}) ||
		AllEquals(iterator.manypeek(2), []string{`Name`, `left shift and assign`}) {
		nameNode := &Node{
			Name: `Variable`,
			Options: map[string]interface{}{
				`Name`: iterator.nextNotIdentAndWhiteSpace().Value,
			},
			Token: iterator.Tokens[iterator.n],
		}
		assignNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		nameNode.Children = append(nameNode.Children, assignNode)

		child, err := p.ProcessExpression(iterator)
		if err != nil {
			return nil, err
		}
		assignNode.Children = append(assignNode.Children, child)
		return nameNode, nil
	}
	//funccall

	condExpressionNode, err := p.ProcessConditionalExpression(iterator)
	if err != nil {
		return nil, err
	}

	return condExpressionNode, nil
}

//ProcessConditionalExpression ...
func (p *Parser) ProcessConditionalExpression(iterator *Iterator) (*Node, error) {
	
	nodeLogOr, err := p.ProcessLogOr(iterator)
	if err != nil {
		return nil, err
	}

	if iterator.peekNotIdent().Value == `if` {
		ifNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		nodeLogOrYoung, err := p.ProcessConditionalExpression(iterator)
		if err != nil {
			return nil, err
		}
		ifNode.Children = append(ifNode.Children, nodeLogOrYoung)

		if iterator.peekNotIdent().Value == `else` {
			elseNode := &Node{
				Name:  iterator.nextNotIdentAndWhiteSpace().Type,
				Token: iterator.Tokens[iterator.n],
			}
			nodeLogOrYoung, err := p.ProcessConditionalExpression(iterator)
			if err != nil {
				return nil, err
			}
			elseNode.Children = append(elseNode.Children, nodeLogOrYoung)
			nodeLogOr.Children = append(nodeLogOr.Children, ifNode, elseNode)
			return nodeLogOr, nil
		}
		return nil, fmt.Errorf("Broken ternary operator at line %d\nForgot else", iterator.Tokens[iterator.n].Line)

	}


	return nodeLogOr, nil
	//return nil, fmt.Errorf("Unexpected token at line: %d, index %d", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset)
}

//ProcessLogOr ...
func (p *Parser) ProcessLogOr(iterator *Iterator) (*Node, error) {
	logAndNode, err := p.ProcessLogAnd(iterator)
	if err != nil {
		return nil, err
	}

	for iterator.peekNotIdent().Type == `logical or` {
		logOrNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		logAndYoung, err := p.ProcessLogAnd(iterator)
		if err != nil {
			return nil, err
		}
		logOrNode.Children = append(logOrNode.Children, logAndYoung)
		logAndNode.Children = append(logAndNode.Children, logOrNode)
	}
	return logAndNode, nil
}

//ProcessLogAnd ...
func (p *Parser) ProcessLogAnd(iterator *Iterator) (*Node, error) {
	bitOrNode, err := p.ProcessBitOr(iterator)
	if err != nil {
		return nil, err
	}

	for iterator.peekNotIdent().Type == `logical and` {
		logAndNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		bitOrYoung, err := p.ProcessBitOr(iterator)
		if err != nil {
			return nil, err
		}
		logAndNode.Children = append(logAndNode.Children, bitOrYoung)
		bitOrNode.Children = append(bitOrNode.Children, logAndNode)
	}
	return bitOrNode, nil
}

//ProcessBitOr ...
func (p *Parser) ProcessBitOr(iterator *Iterator) (*Node, error) {

	xorNode, err := p.ProcessXor(iterator)
	if err != nil {
		return nil, err
	}

	for iterator.peekNotIdent().Type == `or` {
		orNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		xorNodeYoung, err := p.ProcessXor(iterator)
		if err != nil {
			return nil, err
		}
		orNode.Children = append(orNode.Children, xorNodeYoung)
		xorNode.Children = append(xorNode.Children, orNode)
	}
	return xorNode, nil
}

//ProcessXor ...
func (p *Parser) ProcessXor(iterator *Iterator) (*Node, error) {
	bitAndNode, err := p.ProcessBitAnd(iterator)
	if err != nil {
		return nil, err
	}

	for iterator.peekNotIdent().Type == `xor` {
		xorNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		bitAndYoung, err := p.ProcessBitAnd(iterator)
		if err != nil {
			return nil, err
		}
		xorNode.Children = append(xorNode.Children, bitAndYoung)
		bitAndNode.Children = append(bitAndNode.Children, xorNode)
	}
	return bitAndNode, nil
}

//ProcessBitAnd ...
func (p *Parser) ProcessBitAnd(iterator *Iterator) (*Node, error) {
	equalsNode, err := p.ProcessEquals(iterator)
	if err != nil {
		return nil, err
	}

	for iterator.peekNotIdent().Type == `and` {
		andNode := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		equalsYoung, err := p.ProcessEquals(iterator)
		if err != nil {
			return nil, err
		}
		andNode.Children = append(andNode.Children, equalsYoung)
		equalsNode.Children = append(equalsNode.Children, andNode)
	}
	return equalsNode, nil
}

//ProcessEquals ...
func (p *Parser) ProcessEquals(iterator *Iterator) (*Node, error) {
	notEqualsNode, err := p.ProcessNotEqual(iterator)
	if err != nil {
		return nil, err
	}

	for (iterator.peekNotIdent().Type == `equal`) || (iterator.peekNotIdent().Type == `not equal`) {
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		notEqualsYoung, err := p.ProcessNotEqual(iterator)
		if err != nil {
			return nil, err
		}
		child.Children = append(child.Children, notEqualsYoung)
		notEqualsNode.Children = append(notEqualsNode.Children, child)
	}
	return notEqualsNode, nil
}

//ProcessNotEqual ...
func (p *Parser) ProcessNotEqual(iterator *Iterator) (*Node, error) {
	shiftNode, err := p.ProcessShift(iterator)
	if err != nil {
		return nil, err
	}

	for (iterator.peekNotIdent().Type == `greater than`) || (iterator.peekNotIdent().Type == `less than`) || (iterator.peekNotIdent().Type == `greater or equal`) || (iterator.peekNotIdent().Type == `less or equal`) {
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		shiftNodeYoung, err := p.ProcessShift(iterator)
		if err != nil {
			return nil, err
		}
		child.Children = append(child.Children, shiftNodeYoung)
		shiftNode.Children = append(shiftNode.Children, child)
	}
	return shiftNode, nil
}

//ProcessShift ...
func (p *Parser) ProcessShift(iterator *Iterator) (*Node, error) {
	addNode, err := p.ProcessAdd(iterator)
	if err != nil {
		return nil, err
	}

	for (iterator.peekNotIdent().Type == `right shift`) || (iterator.peekNotIdent().Type == `left shift`) {
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		addNodeYoung, err := p.ProcessAdd(iterator)
		if err != nil {
			return nil, err
		}
		child.Children = append(child.Children, addNodeYoung)
		addNode.Children = append(addNode.Children, child)
	}
	return addNode, nil
}

//ProcessAdd ...
func (p *Parser) ProcessAdd(iterator *Iterator) (*Node, error) {
	termNode, err := p.ProcessTerm(iterator)
	if err != nil {
		// if (iterator.peekNotIdent().Type == `plus`) || (iterator.peekNotIdent().Type == `minus`) {
		// 	child := &Node{
		// 		Name:  iterator.nextNotIdentAndWhiteSpace().Type,
		// 		Token: iterator.Tokens[iterator.n],
		// 	}
		// 	iterator.movePointer(-1)//move back to plus/minus
		// 	for (iterator.peekNotIdent().Type == `plus`) || (iterator.peekNotIdent().Type == `minus`) {
		// 		iterator.movePointer(1)//move forward from
		// 		termNodeYoung, err := p.ProcessTerm(iterator)
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		child.Children = append(child.Children, termNodeYoung)
		// 	}
		// 	return child, nil
		// } 
		return nil, err
	}

	for (iterator.peekNotIdent().Type == `plus`) || (iterator.peekNotIdent().Type == `minus`) {
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		termNodeYoung, err := p.ProcessTerm(iterator)
		if err != nil {
			return nil, err
		}
		child.Children = append(child.Children, termNodeYoung)
		termNode.Children = append(termNode.Children, child)
	}
	return termNode, nil
}

//ProcessTerm ...
func (p *Parser) ProcessTerm(iterator *Iterator) (*Node, error) {
	factorNode, err := p.ProcessFactor(iterator)
	if err != nil {
		return nil, err
	}
	for (iterator.peekNotIdent().Type == `multiplicate`) || (iterator.peekNotIdent().Type == `divide`) {
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		factorNodeYoung, err := p.ProcessFactor(iterator)
		if err != nil {
			return nil, err
		}

		child.Children = append(child.Children, factorNodeYoung)
		factorNode.Children = append(factorNode.Children, child)
	}
	return factorNode, nil
}

/*
Expression -> Term [("-" | "+") Term]*
Term -> Factor [("*" | "/") Factor]*
Factor -> "(" Expression ")" | unaryOp Factor | Int | Char | Float
*/
/*
<program> ::= <function>
<function> ::= "int" id "(" ")" "{" { <block-item> } "}"
<block-item> ::= <statement> | <declaration>
<statement> ::= "return" <exp> ";" | <exp> ";"
<declaration> ::= <type> <id> [ = <exp> ] ";"
<exp> ::= <id> "=" <exp> | <log_or>
<log_or> = <log_and> { "||" <log_and> }
<log_and> ::= <bit_or> { "&&" <bit_or> }
<bit_or> ::= <xor> { "|" <xor> }
<xor> ::= <bit_and> { "^" <bit_and> }
<bit_and> ::= <equals> { "&" <equals> }
<equals> ::= <not_equals> { ("==" | "!=") <not_equals> }
<not_equas> ::= <shift> { ("<" | ">" | "<=" | ">=") <shift> }
<shift> ::= <add> { ("<<" | ">>") <add> }
<add> ::= <term> { ("+" | "-") <term> }
<term> ::= <factor> { ("*" | "/" | "%") <factor> }
<factor> ::=  "(" <exp> ")" | <unary_op> <factor> | id | int | float | char
<unary_op> ::= "!" | "~" | "-"
<type> ::= int | float | char
*/

//ProcessFactor ...
func (p *Parser) ProcessFactor(iterator *Iterator) (*Node, error) {

	if AllEquals(iterator.manypeek(3), []string{`assign`, `Name`, `left parentesis`}) { //function call in expression

		funcNode := &Node{
			Name: `function call`,
			Options: map[string]interface{}{
				`Name`: iterator.nextNotIdentAndWhiteSpace().Value,
			},
			Token: iterator.Tokens[iterator.n],
		}
		iterator.nextNotIdentAndWhiteSpace() //go to left parentesis
		for {
			if iterator.peekNotIdent().Type == `Name` || iterator.peekNotIdent().Type == `number` {
				parametrNode := &Node{
					Name: `argument`,
					Options: map[string]interface{}{
						iterator.nextNotIdentAndWhiteSpace().Type: iterator.Tokens[iterator.n].Value,
					},
					Token: iterator.Tokens[iterator.n],
				}
				funcNode.Children = append(funcNode.Children, parametrNode)
				if iterator.peekNotIdent().Type != `comma` && iterator.peekNotIdent().Type != `right parentesis` {
					return nil, fmt.Errorf("Wrong function parametres at line %d index %d\n%s", iterator.Tokens[iterator.n+1].Line, iterator.Tokens[iterator.n+1].Offset, iterator.Tokens[iterator.n+1])
				}
				continue
			} else if iterator.peekNotIdent().Type == `comma` {
				iterator.nextNotIdentAndWhiteSpace() //move to comma
				if iterator.peekNotIdent().Type != `Name` && iterator.peekNotIdent().Type != `number` {
					return nil, fmt.Errorf("Expected name after line %d index %d\n%s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
				}
				continue
			} else if iterator.peekNotIdent().Type == `right parentesis` {
				iterator.nextNotIdentAndWhiteSpace() //move to right parentesis
				if AllEquals(iterator.manypeek(2), []string{`right parentesis`, `whitespace`}) || AllEquals(iterator.manypeek(2), []string{`right parentesis`, `commentary`}) {
					//swap arguments places
					funcNode.Children = reverse(funcNode.Children)
					break
				} 
				return nil, fmt.Errorf("Wrong function declaration start at line %d", iterator.Tokens[iterator.n].Line)
			} else {
				return nil, fmt.Errorf("Expected argument at line %d index %d\n%s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
			}
		}
		return funcNode, nil
	}
	//just function call without assigning
	if AllEquals(iterator.manypeek(2), []string{`Name`, `left parentesis`}) { //function call in expression
		funcNode := &Node{
			Name: `function call`,
			Options: map[string]interface{}{
				iterator.nextNotIdentAndWhiteSpace().Type: iterator.Tokens[iterator.n].Value,
			},
			Token: iterator.Tokens[iterator.n],
		}

		iterator.nextNotIdentAndWhiteSpace() //go to left parentesis
		for {
			if iterator.peekNotIdent().Type == `Name` || iterator.peekNotIdent().Type == `number` {
				parametrNode := &Node{
					Name: `argument`,
					Options: map[string]interface{}{
						iterator.nextNotIdentAndWhiteSpace().Type: iterator.Tokens[iterator.n].Value,
					},
					Token: iterator.Tokens[iterator.n],
				}

				funcNode.Children = append(funcNode.Children, parametrNode)
				if iterator.peekNotIdent().Type != `comma` && iterator.peekNotIdent().Type != `right parentesis` {
					return nil, fmt.Errorf("Wrong function parametres at line %d index %d\n%s", iterator.Tokens[iterator.n+1].Line, iterator.Tokens[iterator.n+1].Offset, iterator.Tokens[iterator.n+1])
				}
				continue
			} else if iterator.peekNotIdent().Type == `comma` {
				iterator.nextNotIdentAndWhiteSpace() //move to comma
				if iterator.peekNotIdent().Type != `Name` && iterator.peekNotIdent().Type != `number` {
					return nil, fmt.Errorf("Expected name after line %d index %d\n%s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
				}
				continue
			} else if iterator.peekNotIdent().Type == `right parentesis` {
				iterator.nextNotIdentAndWhiteSpace() //move to right parentesis
				if AllEquals(iterator.manypeek(2), []string{`right parentesis`, `whitespace`}) {
					break
				}
				return nil, fmt.Errorf("Wrong function declaration start at line %d", iterator.Tokens[iterator.n].Line)
			} else {
				return nil, fmt.Errorf("Expected argument at line %d index %d\n%s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n])
			}
		}

		return funcNode, nil
	}

	if iterator.peekNotIdent().Type == `left parentesis` {
		parentesisNode := &Node{
			Name: `parentesis`,
		}
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		expressionYoung, err := p.ProcessExpression(iterator)
		if err != nil {
			return nil, err
		}
		child.Children = append(child.Children, expressionYoung)

		parentesisNode.Children = append(parentesisNode.Children, child)

		if iterator.peekNotIdent().Type == `right parentesis` {
			child := &Node{
				Name:  iterator.nextNotIdentAndWhiteSpace().Type,
				Token: iterator.Tokens[iterator.n],
			}
			parentesisNode.Children = append(parentesisNode.Children, child)
			return parentesisNode, nil
		}
		return nil, fmt.Errorf("Cant find right parentesis\nLine: %d", iterator.Tokens[iterator.n].Line)
	}

	if iterator.peekNotIdent().Type == `logical not` || iterator.peekNotIdent().Type == `logical and` {
		child := &Node{
			Name:  iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}

		childYoung, err := p.ProcessFactor(iterator)
		if err != nil {
			return nil, err
		}
		child.Children = append(child.Children, childYoung)
		return child, nil
	}
	if iterator.peekNotIdent().Type == `Name` {
		child := &Node{
			Name: iterator.nextNotIdentAndWhiteSpace().Type,
			Options: map[string]interface{}{
				`Name`: iterator.Tokens[iterator.n].Value,
			},
			Token: iterator.Tokens[iterator.n],
		}
		return child, nil
	}

	if iterator.peekNotIdent().Type == `number` {
		child := &Node{
			Name: iterator.nextNotIdentAndWhiteSpace().Type,
			Options: map[string]interface{}{
				`Value`: iterator.Tokens[iterator.n].Value,
			},
		}
		return child, nil
	}
	if iterator.peekNotIdent().Type == `Boolean` {
		child := &Node{
			Name: iterator.nextNotIdentAndWhiteSpace().Type,
			Options: map[string]interface{}{
				`Value`: iterator.Tokens[iterator.n].Value,
			},
			Token: iterator.Tokens[iterator.n],
		}
		return child, nil
	}
	if iterator.peekNotIdent().Type == `continue` {
		child := &Node{
			Name: iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		return child, nil
	}
	if iterator.peekNotIdent().Type == `break` {
		child := &Node{
			Name: iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
		}
		return child, nil
	}
	if iterator.peekNotIdent().Type == `commentary` {
		child := &Node{
			Name: iterator.nextNotIdentAndWhiteSpace().Type,
			Token: iterator.Tokens[iterator.n],
			Options: map[string] interface{}{
				iterator.Tokens[iterator.n].Type: iterator.Tokens[iterator.n].Value,
			},
		}
		return child, nil
	}
	if iterator.peekNotIdent().Type == `whitespace` {
		return nil, fmt.Errorf("Unexpected end of line: %d, index: %d %s.\nUnexpected %s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n], iterator.peekNotIdent())
	}
	//iterator.findEndOfLine()
	return nil, fmt.Errorf("Unexpected token after line: %d, index: %d %s.\nUnexpected %s", iterator.Tokens[iterator.n].Line, iterator.Tokens[iterator.n].Offset, iterator.Tokens[iterator.n], iterator.peekNotIdent())
}

//Iterator struct
type Iterator struct {
	n      int
	Tokens []*lexer.Token
}

//NewIterator ...
func NewIterator() *Iterator {
	i := &Iterator{
		n: -1,
	}
	return i
}
func (i *Iterator) nextToken() *lexer.Token {
	if i.hasNext() {
		i.n++
		return i.Tokens[i.n]
	}
	return i.Tokens[i.n]
}

func (i *Iterator) peek() *lexer.Token {
	if i.hasNext() {
		k := i.n + 1
		return i.Tokens[k]
	}
	return i.Tokens[i.n]
}
func (i *Iterator) manypeek(a int) []*lexer.Token {
	var (
		array []*lexer.Token
		currN = i.n
	)

	if i.Tokens[i.n].Type == `ident` {
		i.clearNextSpace()
	}
	for j := 0; j < a; j++ {
		array = append(array, i.Tokens[i.n])
		i.nextToken()
		i.clearNextSpace()
	}
	i.n = currN
	return array

}
func (i *Iterator) clearNextSpace() bool {
	for i.Tokens[i.n].Type == `ident` && i.hasNext() {
		i.n++
	}
	return true
}

func (i *Iterator) hasNext() bool {
	if len(i.Tokens)-1 != i.n {
		return true
	}
	return false
}

func (i *Iterator) findEndOfLine() bool {
	for i.Tokens[i.n].Type != `whitespace` && i.hasNext() {
		i.n++
	}
	if i.hasNext() {
		i.n++
	}
	return true
}

func (i *Iterator) nextNotIdentAndWhiteSpace() *lexer.Token {
	for i.hasNext() {
		if i.nextToken().Type != `ident` && i.Tokens[i.n].Type != `whitespace` {
			return i.Tokens[i.n]
		}
	}
	return i.Tokens[i.n]
}
func (i *Iterator) peekNotIdent() *lexer.Token {
	var (
		currN int = i.n
		token *lexer.Token
	)
	for i.hasNext() {
		if i.nextToken().Type != `ident` {
			token = i.Tokens[i.n]
			i.n = currN
			return token
		}
	}
	return i.Tokens[i.n]
}
func (i *Iterator) countIdents() int {
	spaces := 0
	for i.hasNext() {
		if i.Tokens[i.n+spaces].Type == `ident` {
			spaces++
			continue
		}
		return spaces
	}
	return spaces
}

func (i *Iterator) movePointer(a int) *lexer.Token {
	if a+i.n < len(i.Tokens) {
		i.n += a
		return i.Tokens[i.n]
	}
	return i.Tokens[i.n]
}

//AllEquals ...
func AllEquals(arr []*lexer.Token, stringArray []string) bool {
	if len(stringArray) != len(arr) {
		return false
	}
	for i, v := range arr {
		if v.Type != stringArray[i] {
			return false
		}
	}
	return true
}
