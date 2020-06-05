package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Type of token
type Type string

// Position ...
type Position struct {
	Index, Line, Column   int
	FileName, FileContent string
}

// Next position
func (p *Position) Next(currentChar rune) {
	p.Index++
	if currentChar == '\n' {
		p.Column = 0
		p.Line++
	}
	p.Column++
}

// Copy position
func (p Position) Copy() Position { return p }

const (
	TypeInt   Type = "INT"
	TypeFloat Type = "FLOAT"
	TypePlus  Type = "PLUS"
	TypeMinus Type = "MINUS"
	TypeMul   Type = "MUL"
	TypeDiv   Type = "DIV"
	TypeLP    Type = "LP"
	TypeRP    Type = "RP"
)

// ERR_EOF ...
var ERR_EOF error = errors.New("EOF")

// Token ...
type Token struct {
	Type  Type
	Value interface{}
}

// IToken ...
type IToken interface {
	FToken() // should be a unique random name, just to use polymorphisme
}

// FToken ...
func (t Token) FToken() {}

// TokenPlus ...
type TokenPlus struct{ Token }

// NewTokenPlus ...
func NewTokenPlus() TokenPlus { return TokenPlus{Token{TypePlus, nil}} }

// Eval ...
func (t TokenPlus) Eval(left, right IExpression) float64 {
	return left.Eval() + right.Eval()
}

// TokenMinus ...
type TokenMinus struct{ Token }

// NewTokenMinus ...
func NewTokenMinus() TokenMinus { return TokenMinus{Token{TypeMinus, nil}} }

// Eval ...
func (t TokenMinus) Eval(left, right IExpression) float64 {
	return left.Eval() - right.Eval()
}

// TokenMul ...
type TokenMul struct{ Token }

// NewTokenMul ...
func NewTokenMul() TokenMul { return TokenMul{Token{TypeMul, nil}} }

// Eval ...
func (t TokenMul) Eval(left, right IExpression) float64 {
	return left.Eval() * right.Eval()
}

// TokenDiv ...
type TokenDiv struct{ Token }

// NewTokenDiv ...
func NewTokenDiv() TokenDiv { return TokenDiv{Token{TypeDiv, nil}} }

// Eval ...
func (t TokenDiv) Eval(left, right IExpression) float64 {
	// TODO: check div by zero
	return left.Eval() / right.Eval()
}

// TokenLP ...
type TokenLP struct{ Token }

// NewTokenLP ...
func NewTokenLP() TokenLP { return TokenLP{Token{TypeLP, nil}} }

// TokenRP ...
type TokenRP struct{ Token }

// NewTokenRP ...
func NewTokenRP() TokenRP { return TokenRP{Token{TypeRP, nil}} }

// TokenInt ...
type TokenInt struct{ Token }

// NewTokenInt ...
func NewTokenInt(value int) TokenInt { return TokenInt{Token{TypeInt, value}} }

// TokenFloat ...
type TokenFloat struct{ Token }

// NewTokenFloat ...
func NewTokenFloat(value float64) TokenFloat { return TokenFloat{Token{TypeFloat, value}} }

// IExpression ...
type IExpression interface {
	Eval() float64
}

// Eval ...
func (t TokenInt) Eval() float64 { return float64(t.Value.(int)) }

// Eval ...
func (t TokenFloat) Eval() float64 { return t.Value.(float64) }

// BinOpNode ...
type BinOpNode struct {
	Left, Right IExpression
	Op          Operation
}

func (b BinOpNode) String() string {
	return fmt.Sprintf("(%v,%v,%v)", b.Left, b.Op, b.Right)
}

// Eval ...
func (b BinOpNode) Eval() float64 {
	return b.Op.Eval(b.Left, b.Right)
}

// Operation ...
type Operation interface {
	Eval(left, right IExpression) float64
}

// String ...
func (t Token) String() string {
	if t.Value == nil {
		return string(t.Type)

	}
	if t.Type == TypeFloat {
		return fmt.Sprintf("%v:%.3f", t.Type, t.Value)
	}
	return fmt.Sprintf("%v:%v", t.Type, t.Value)
}

// Lexer ...
type Lexer struct {
	Text    string
	Pos     Position
	Current rune
}

// Next ...
func (l *Lexer) Next() bool {
	l.Pos.Next(l.Current)
	if l.Pos.Index < len(l.Text) {
		l.Current = rune(l.Text[l.Pos.Index])
		return true
	}
	l.Current = ' '
	return false
}

// Tokens ...
type Tokens []IToken

// Add ...
func (t Tokens) Add(tokens ...IToken) Tokens {
	return append(t, tokens...)
}

// MakeTokens ...
func (l *Lexer) MakeTokens() (Tokens, error) {
	ret := Tokens{}

	current := l.Current
	switch current {
	case ' ', '\t':
		if !l.Next() {
			return ret, nil
		}
		return l.MakeTokens()
	case '+':
		ret = ret.Add(NewTokenPlus())
	case '-':
		ret = ret.Add(NewTokenMinus())
	case '*':
		ret = ret.Add(NewTokenMul())
	case '/':
		ret = ret.Add(NewTokenDiv())
	case '(':
		ret = ret.Add(NewTokenLP())
	case ')':
		ret = ret.Add(NewTokenRP())
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		token := l.MakeNumber()
		ret = ret.Add(token)
	default:
		return ret, fmt.Errorf("unknown token %q at file: %v, line: %v, col:% v", string(current), l.Pos.FileName, l.Pos.Line, l.Pos.Column)
	}
	if !l.Next() {
		return ret, nil
	}
	tokens, err := l.MakeTokens()
	if err != nil {
		return ret, err
	}
	return ret.Add(tokens...), nil
}

func isDigit(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// MakeNumber ...
func (l *Lexer) MakeNumber() IToken {
	dotCount := 0
	numStr := ""

	for {
		if isDigit(l.Current) {
			numStr += string(l.Current)
		} else if l.Current == '.' && dotCount == 0 {
			numStr += "."
			dotCount++
		} else {
			break
		}
		l.Next()
	}
	if dotCount == 0 {
		n, err := strconv.ParseInt(numStr, 10, 32)
		if err != nil {
			panic(err)
		}
		return NewTokenInt(int(n))
	}
	f, err := strconv.ParseFloat(numStr, 32)
	if err != nil {
		panic(err)
	}
	return NewTokenFloat(f)
}

// NewLexer ...
func NewLexer(fileName, text string) *Lexer {
	return &Lexer{text, Position{-1, 0, -1, fileName, text}, ' '}
}

// NumberNode ...
type NumberNode struct{ Token }

// Parser ...
type Parser struct {
	Tokens       Tokens
	TokenIndex   int
	CurrentToken IToken
}

// NewParser ...
func NewParser(tokens Tokens) *Parser {
	p := &Parser{tokens, -1, Token{}}
	p.Next()
	return p
}

// Next ...
func (p *Parser) Next() bool {
	p.TokenIndex++
	if p.TokenIndex < len(p.Tokens) {
		p.CurrentToken = p.Tokens[p.TokenIndex]
		return true
	}
	return false
}

// Parse ...
func (p *Parser) Parse() IExpression {
	return p.Expression()
}

// Factor ...
func (p *Parser) Factor() IExpression {
	var node IExpression
	switch token := p.CurrentToken.(type) {
	case TokenFloat, TokenInt:
		node = token.(IExpression)
	}
	p.Next()
	return node
}

// Term ...
func (p *Parser) Term() IExpression {
	left := p.Factor()

	var op Operation
	var binOp BinOpNode

	switch token := p.CurrentToken.(type) {
	case TokenDiv, TokenMul:
		op = token.(Operation)
		p.Next()
		right := p.Factor()
		binOp = BinOpNode{left, right, op}
	}
	if binOp.Op == nil {
		return left
	}
	return binOp
}

// Expression ...
func (p *Parser) Expression() IExpression {
	left := p.Term()

	var op Operation
	var binOp BinOpNode

	switch token := p.CurrentToken.(type) {
	case TokenPlus, TokenMinus:
		op = token.(Operation)
		p.Next()
		right := p.Term()
		binOp = BinOpNode{left, right, op}
	}
	if binOp.Op == nil {
		return left
	}
	return binOp
}

func main() {

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Basic > ")
		l, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		text := string(l)

		lexer := NewLexer("stdin", text)
		tokens, err := lexer.MakeTokens()
		if err != nil {
			log.Println("err:", err)
			continue
		}
		parser := NewParser(tokens)
		expr := parser.Parse()
		fmt.Println(expr)
		fmt.Println(tokens)
		fmt.Println(expr.Eval())
	}
}
