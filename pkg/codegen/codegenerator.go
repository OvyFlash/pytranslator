package codegen

import (
	"fmt"
	"strings"

	"../parser"
)

var template string = `
.586
.model flat, stdcall
option casemap:none		;розрізнювати великі та маленькі букви.
; Підключає файли та бібліотеки, необхідні для роботи програми.
include     \masm32\include\windows.inc
include     \masm32\include\kernel32.inc
include     \masm32\include\masm32.inc
include     \masm32\include\masm32rt.inc
includelib  \masm32\lib\kernel32.lib
includelib  \masm32\lib\masm32.lib

main        PROTO

.data
%s
.code
start:	; точка входу у програму
; директива INVOKE означає виклик процедури замість call для виклику процедур
; з параметрами, які передаються через стек командами push.
    invoke  main
	fn MessageBox, 0, str$(eax), "Shvachko", MB_OK
	invoke  ExitProcess,0		; завершення роботи програми
%s
END start`

//GenerateMASM generates masm code
func GenerateMASM(nodes []*parser.Node) (string, error) {
	//create channel with Nodes for iteration
	NodesChannel := NodeChannelIterator(nodes)
	var (
		masmCode []string
		masmData []string
	)
	for node := range NodesChannel {
		if node.Name == `function` {
			childCode, childData, err := GenerateCode(node)
			if err != nil {
				return "", err
			}
			masmCode = append(masmCode, childCode)
			masmData = append(masmData, childData)
		}
	}

	return fmt.Sprintf(template, strings.Join(masmData, "\n"), strings.Join(masmCode, "\n\n")), nil //first is .data section, second is .data section
}

//GenerateCode ...
func GenerateCode(node *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	switch node.Name {
	case `function`:
		//Update currFunction value
		//Create template with saving stack and return
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing function %s body\nLine: %d", node.Options[`Name`].(string), node.Token.Line)
		}
		childCode, childData, err := GenerateFunction(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `parameter`:
		//write to memory ebp [--8]
		childCode, childData, err := GenerateParameter(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil

	case `return value`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Empty return\nLine: %d", node.Token.Line)
		}
		childCode, childData, err := GenerateReturn(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `Variable`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing variable declaration %s.\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateVariable(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	// case `assign`:
	// 	if len(node.Children) == 0 {
	// 		return "", "", fmt.Errorf("Empty assign\nLine: %d, index %d", node.Token.Line, node.Token.Offset)
	// 	}
	// 	childCode, childData, err := GenerateLSAAssign(node)
	// 	if err != nil {
	// 		return "", "", err
	// 	}
	// 	return childCode, childData, nil
	// case `left shift and assign`:
	// 	if len(node.Children) == 0 {
	// 		return "", "", fmt.Errorf("Missing assign\nLine: %d, index %d", node.Token.Line, node.Token.Offset)
	// 	}
	// 	childCode, childData, err := GenerateAssign(node)
	// 	if err != nil {
	// 		return "", "", err
	// 	}
	// 	return childCode, childData, nil
	case `number`:
		//can be zero
		childCode, childData, err := GenerateNumber(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `Name`:
		//can be zero
		childCode, childData, err := GenerateName(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil

	case `function call`:
		//check number of parameters
		childCode, childData, err := GenerateFunctionCall(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `plus`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`NGenerateMinusame`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GeneratePlus(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `minus`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateMinus(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `left shift`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateLeftShift(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `divide`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateDivide(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `multiplicate`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateMultiplicate(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `IF`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateIF(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `ELSE`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateELSE(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `logical not`:
		if len(node.Children) == 0 {
			return "", "", fmt.Errorf("Missing token after %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateLogicalNot(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `while cycle`:
		if len(node.Children) == 1 {
			return "", "", fmt.Errorf("Forever while loop %s\nLine: %d, index %d", node.Options[`Name`].(string), node.Token.Line, node.Token.Offset)
		}
		childCode, childData, err := GenerateWhile(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `Boolean`:
		childCode, childData, err := GenerateBoolean(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `break`:
		childCode, childData, err := GenerateBreak(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `continue`:
		childCode, childData, err := GenerateContinue(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `less or equal`: 
		childCode, childData, err := GenerateLessOrEqual(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	case `logical and`:
		childCode, childData, err := GenerateLogicalAnd(node)
		if err != nil {
			return "", "", err
		}
		return childCode, childData, nil
	default:
		return "", "", fmt.Errorf("UNKNOWN NODE %v", node)
	}
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

type Variables struct {
	v map[string]*isParam
}

type isParam struct {
	ID      int
	IsParam bool
}

var (
	//keeps current function variables list
	CurrVariables map[string]*Variables = make(map[string]*Variables)
	//keeps current function name
	CurrFunction string
)

//GenerateFunction writes to CurrFunction function name and prepares entry and return templates
func GenerateFunction(FunctionNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	CurrFunction = FunctionNode.Options[`Name`].(string)
	CurrVariables[CurrFunction] = &Variables{
		v: make(map[string]*isParam),
	}

	code = append(code,
		fmt.Sprintf("%s PROC", CurrFunction),
		"push ebp",
		"mov ebp, esp")

	for _, v := range FunctionNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	
	code = append(code,
		"mov esp, ebp",
		"pop ebp",
		"ret\n"+
		fmt.Sprintf("%s ENDP", CurrFunction))

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil

}

//GenerateParameter pushes from memory values to stack and create them in local memory
func GenerateParameter(ParameterNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	variableName := ParameterNode.Options[`Name`].(string) //save name

	if CurrVariables[CurrFunction].v[variableName] == nil {
		CurrVariables[CurrFunction].v[variableName] = &isParam{
			ID:      len(CurrVariables[CurrFunction].v)+1,
			IsParam: true,
		}
	} else {
		return "", "", fmt.Errorf("You cannot use same names for parameters\nLine: %d, index: %d, %s", ParameterNode.Token.Line, ParameterNode.Token.Offset, ParameterNode.Token)
	}

	code = append(code,
		fmt.Sprintf("mov eax, [ebp + %d] ;var %s", (CurrVariables[CurrFunction].v[variableName].ID-1)*4+8, variableName),
		"push eax")

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil

}

//GenerateReturn only pops the variable to eax
func GenerateReturn(ReturnNode *parser.Node) (string, string, error){
	var (
		code []string
		data []string
	)

	for _, v := range ReturnNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "pop eax")
	
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
} 
//GenerateVariable only pops the variable to eax
func GenerateVariable(VariableNode *parser.Node) (string, string, error){
	var (
		code []string
		data []string
	)
	variableName := VariableNode.Options[`Name`].(string)
	var justCreated bool
	if CurrVariables[CurrFunction].v[variableName] == nil {
		justCreated = true
		var countNotParams int
		for _, v := range CurrVariables[CurrFunction].v {//count not params
			if !v.IsParam {
				countNotParams++
			}
		}
		CurrVariables[CurrFunction].v[variableName] = &isParam{
			ID:      countNotParams + 1,
			IsParam: false,
		}
	} 

	for _, v := range VariableNode.Children {
		if v.Name == `assign` {
			if len(v.Children) == 0 {
				return "", "", fmt.Errorf("Empty assign\nLine: %d, index %d", v.Token.Line, v.Token.Offset)
			}
			childCode, childData, err := GenerateAssign(v)
			if err != nil {
				return "", "", err
			}
			code = append(code, childCode)
			data = append(data, childData)
			break
		}
		if v.Name == `left shift and assign` {
			if justCreated {
				return "", "", fmt.Errorf("Local variable '%s' referenced before assignnment\nLine: %d, index: %d", variableName, VariableNode.Token.Line, VariableNode.Token.Offset)
			}
			if len(v.Children) == 0 {
				return "", "", fmt.Errorf("Missing assign\nLine: %d, index %d", v.Token.Line, v.Token.Offset)
			}
			code = append(code, 
					fmt.Sprintf("mov eax, [ebp-%d]", CurrVariables[CurrFunction].v[variableName].ID*4),
					"push eax; start left shift and assign")
			childCode, childData, err := GenerateLSAAssign(v)
			if err != nil {
				return "", "", err
			}
			code = append(code, childCode)
			data = append(data, childData)
			code = append(code ,"pop ecx", "pop eax", "shl eax, cl", "push eax ; finish left shift and assign")
			break
		}
	}
	if CurrVariables[CurrFunction].v[variableName].IsParam {
		code = append(code, "pop eax",
			fmt.Sprintf("mov [ebp+%d], eax ;var %s", (CurrVariables[CurrFunction].v[variableName].ID-1)*4+8, variableName,),
			"push eax")
	} else {
		code = append(code, "pop eax",
			fmt.Sprintf("mov [ebp-%d], eax ;var %s", CurrVariables[CurrFunction].v[variableName].ID*4, variableName,),
			"push eax")
	}
	
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
} 
//GenerateAssign only pushes the variable to eax
func GenerateAssign(AssignNode *parser.Node) (string, string, error){
	var (
		code []string
		data []string
	)
	
	for _, v := range AssignNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
} 
//GenerateLSAAssign only pushes the variable to eax
func GenerateLSAAssign(LSAAssignNode *parser.Node) (string, string, error){
	var (
		code []string
		data []string
	)
	
	for _, v := range LSAAssignNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
} 

//GenerateNumber gives number 
func GenerateNumber(NumberNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	numberCode := NumberNode.Options[`Value`].(string)

	findOctal := false
	for _, v := range numberCode {
		if findOctal {
			numberCode += string(v)
		}
		if string(v) == "o" || string(v) == "O" {
			findOctal = true
			numberCode = ""
		}
	}
	if findOctal {
		numberCode += "o"
	}
	code = append(code, "mov eax, "+ numberCode, "push eax")

	for _, v := range NumberNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil

}

//GenerateFunctionCall ...
func GenerateFunctionCall(FunctionCallNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	funcName := FunctionCallNode.Options[`Name`].(string)
	if parser.Functions[funcName] == 0 {
		return "", "", fmt.Errorf("Function with name %s does not exists,\nLine: %d, index: %d", funcName, FunctionCallNode.Token.Line, FunctionCallNode.Token.Offset)
	}
	for _, v := range FunctionCallNode.Children {
		if v.Options[`number`] != nil {
			code = append(code, 
				fmt.Sprintf("mov eax, %s", v.Options[`number`]),
				"push eax")
			continue
		}
		if v.Options[`Name`] != nil {
			if CurrVariables[CurrFunction].v[v.Options[`Name`].(string)] == nil {
				return "", "", fmt.Errorf("Local variable '%s' referenced before assignnment\nLine: %d, index: %d", v.Options[`Name`].(string), FunctionCallNode.Token.Line, FunctionCallNode.Token.Offset)
			}
			code = append(code,
				fmt.Sprintf("mov eax, [ebp - %d]", CurrVariables[CurrFunction].v[v.Options[`Name`].(string)].ID * 4),
				"push eax")
			continue
		}
 	}
	code = append(code, 
		"call "+ funcName,
		fmt.Sprintf("add esp, %d", 4*len(FunctionCallNode.Children)),
		"push eax")
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

//GenerateName ...
func GenerateName(NameNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	name := NameNode.Options[`Name`].(string)
	if CurrVariables[CurrFunction].v[name] == nil {
		return "", "", fmt.Errorf("Local variable '%s' referenced before assignnment\nLine: %d, index: %d", name, NameNode.Token.Line, NameNode.Token.Offset)
	}
	if CurrVariables[CurrFunction].v[name].IsParam {
		code = append(code,
			fmt.Sprintf("mov eax, [ebp+%d]", (CurrVariables[CurrFunction].v[name].ID-1)*4+8),
			"push eax")
	} else {
		code = append(code, 
			fmt.Sprintf("mov eax, [ebp-%d]", CurrVariables[CurrFunction].v[name].ID*4),
			"push eax")
	}
	for _, v := range NameNode.Children {
	//	code = append(code, fmt.Sprintf("push [ebp-%d]", CurrVariables[CurrFunction].v[name].ID*4))
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//GeneratePlus ...
func GeneratePlus(PlusNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range PlusNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code,
	"pop eax", "pop ebx", "add eax, ebx", "push eax")
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

//GenerateMinus ...
func GenerateMinus(MinusNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range MinusNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code,
	"pop ebx", "pop eax", "sub eax, ebx", "push eax")
	
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//GenerateLeftShift ...
func GenerateLeftShift(LeftShiftNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range LeftShiftNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "pop ecx", "pop eax", "shl eax, cl", "push eax")
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//GenerateDivide ...
func GenerateDivide(DivideNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range DivideNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "mov edx, 0", "pop eax", "pop ebx", "div ebx", "push edx")
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//GenerateMultiplicate ...
func GenerateMultiplicate(MultiplicateNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range MultiplicateNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "pop eax", "pop ebx", "imul ebx", "push eax")
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

var countIF int = 1
//GenerateIF ...
func GenerateIF(IFNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range IFNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "pop eax", "cmp eax, 0", fmt.Sprintf("je _e%d", countIF), fmt.Sprintf("jmp _post_cond%d", countIF))
	countIF++
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//GenerateELSE ...
func GenerateELSE(ELSENode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	code = append(code, fmt.Sprintf("_e%d:",countIF-1), "pop eax")
	for _, v := range ELSENode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, fmt.Sprintf("_post_cond%d:", countIF-1))
	countIF++
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

var countLN int
//GenerateLogicalNot ...
func GenerateLogicalNot(LogicalNotNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
	)
	for _, v := range LogicalNotNode.Children {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}

	countLN+=2
	//cmp eax, 0 if true then set 1 if false then set 0 
	code = append(code, "pop eax;start not",
						"cmp eax, 0",
						fmt.Sprintf("je _ln%d; if true", countLN-1),
						"mov eax, 0; set eax 0 if it wasnt",
						"push eax",
						fmt.Sprintf("jmp _ln%d; post conditional", countLN),
						fmt.Sprintf("_ln%d:", countLN-1),
						"mov eax, 1",
						"push eax",
						fmt.Sprintf("_ln%d:", countLN))
	 //"neg eax", "push eax;finish not")
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

var countWhile int
//GenerateWhile ...
func GenerateWhile(WhileNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
		
	)
	countWhile += 2
	code = append(code, fmt.Sprintf("_cw%d: ;start while", countWhile-1))
	

	// if WhileNode.Children[0].Options[`Name`].(string) != "" {
	// 	childCode, _, err := GenerateName(WhileNode.Children[0])
	// 	if err != nil {
	// 		return "", "", err
	// 	}
	// 	code = append(code, childCode)
	// 	code = append(code, "pop eax",
	// 				"cmp eax, 0",//if 0 - false
	// 				fmt.Sprintf("je _cw%d",countWhile))//jump to post exp
	// } else if WhileNode.Children[0].Name == `number` {
	// 	childCode, _, err := GenerateNumber(WhileNode.Children[0])
	// 	if err != nil {
	// 		return "", "", err
	// 	}
	// 	code = append(code, childCode)
	// }
	childCode, _, err := GenerateCode(WhileNode.Children[0])
	if err != nil {
		return "", "", err
	}
	code = append(code, childCode)
	code = append(code, "pop eax",
					"cmp eax, 0",//if 0 - false
					fmt.Sprintf("je _cw%d",countWhile))//jump to post exp

	//then the cycle body starts
	for _, v := range WhileNode.Children[1:] {
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}
		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, fmt.Sprintf("jmp _cw%d", countWhile-1))
	code = append(code, fmt.Sprintf("_cw%d:",countWhile))//post exp
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

//GenerateBoolean ...
func GenerateBoolean(BooleanNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
		
	)
	if BooleanNode.Options[`Value`].(string) == `True` {
		code = append(code, `push 1`)
	} else {
		code = append(code, `push 0`)
	}
	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

//GenerateBreak ...
func GenerateBreak(BreakNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
		
	)
	code = append(code, fmt.Sprintf("jmp _cw%d",countWhile))//post exp

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//GenerateContinue ...
func GenerateContinue(BreakNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
		
	)
	code = append(code, fmt.Sprintf("jmp _cw%d",countWhile-1))//post exp

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

var lessOrEqualCount int = 0
//GenerateLessOrEqual ...
func GenerateLessOrEqual(LessOrEqualNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
		
	)
	for _, v := range LessOrEqualNode.Children {//push smth
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}

		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "pop ebx", "pop eax", "cmp eax, ebx")
	code = append(code, fmt.Sprintf("jle _less_or_eq%d",lessOrEqualCount),
						"push 0", fmt.Sprintf("jmp _post_jle%d", lessOrEqualCount),
						fmt.Sprintf("_less_or_eq%d:",lessOrEqualCount), "push 1",
						fmt.Sprintf("_post_jle%d:", lessOrEqualCount))//post exp
	lessOrEqualCount++
	//if true push 1 
	//if false push 0

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}

var logicalAndCount int = 0
//GenerateLogicalAnd ...
func GenerateLogicalAnd(LogicalAndNode *parser.Node) (string, string, error) {
	var (
		code []string
		data []string
		
	)
	for _, v := range LogicalAndNode.Children {//push smth
		childCode, childData, err := GenerateCode(v)
		if err != nil {
			return "", "", err
		}

		code = append(code, childCode)
		data = append(data, childData)
	}
	code = append(code, "pop ebx", "pop eax", "cmp eax, ebx")
	code = append(code, fmt.Sprintf("je _logand%d",logicalAndCount),
						"push 0", fmt.Sprintf("jmp _post_logand%d", logicalAndCount),
						fmt.Sprintf("_logand%d:",logicalAndCount), "push 1",
						fmt.Sprintf("_post_logand%d:", logicalAndCount))//post exp
	logicalAndCount++
	//if true push 1 
	//if false push 0

	return strings.Join(code, "\n\t"), strings.Join(data, "\n\t"), nil
}
//NodeChannelIterator Iterator for nodes
func NodeChannelIterator(nodes []*parser.Node) <-chan *parser.Node {
	ch := make(chan *parser.Node)
	go func() {
		for _, val := range nodes {
			ch <- val
		}
		close(ch)
	}()
	return ch
}
