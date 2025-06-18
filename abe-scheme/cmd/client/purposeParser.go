package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

// Token types
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenAND
	TokenOR
	TokenLParen
	TokenRParen
	TokenIdent
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// AST Node types
type NodeType int

const (
	NodeAND NodeType = iota
	NodeOR
	NodeIdent
)

// AST Node
type Node struct {
	Type     NodeType
	Value    []string // Only for identifiers
	Children []*Node  // For operators
}

func (n *Node) String() string {
	switch n.Type {
	case NodeAND:
		return fmt.Sprintf("AND(%v, %v)", n.Children[0], n.Children[1])
	case NodeOR:
		return fmt.Sprintf("OR(%v, %v)", n.Children[0], n.Children[1])
	case NodeIdent:
		return strings.Join(n.Value, " | ")
	default:
		return "UNKNOWN"
	}
}

// Lexer implementation
type Lexer struct {
	input  string
	pos    int
	width  int
	start  int
	tokens chan Token
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		tokens: make(chan Token),
	}
	go l.run()
	return l
}

func (l *Lexer) NextToken() Token {
	return <-l.tokens
}

const eof = -1

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) emit(t TokenType) {
	l.tokens <- Token{
		Type:  t,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
	l.start = l.pos
}

func (l *Lexer) errorf(format string, args ...interface{}) {
	l.tokens <- Token{
		Type:  TokenError,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.start,
	}
}

func (l *Lexer) run() {
	for state := lexWhitespace; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

type stateFn func(*Lexer) stateFn

func lexWhitespace(l *Lexer) stateFn {
	for {
		r := l.next()
		if r == eof {
			l.emit(TokenEOF)
			return nil
		}
		if !unicode.IsSpace(r) {
			l.backup()
			break
		}
	}
	l.start = l.pos
	return lexToken
}

func lexToken(l *Lexer) stateFn {
	r := l.next()

	switch {
	case r == eof:
		l.emit(TokenEOF)
		return nil
	case r == '(':
		l.emit(TokenLParen)
	case r == ')':
		l.emit(TokenRParen)
	case unicode.IsLetter(r):
		l.backup()
		return lexIdent
	default:
		l.errorf("unexpected character: %#U", r)
		return nil
	}

	return lexWhitespace
}

func lexIdent(l *Lexer) stateFn {
	for {
		r := l.next()
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			l.backup()
			break
		}
	}

	word := l.input[l.start:l.pos]
	switch strings.ToUpper(word) {
	case "AND":
		l.emit(TokenAND)
	case "OR":
		l.emit(TokenOR)
	default:
		l.emit(TokenIdent)
	}

	return lexWhitespace
}

type Parser struct {
	pc     utils.PolicyConfig
	lexer  *Lexer
	token  Token
	peek   Token
	errors []string
}

func NewParser(input string, pc utils.PolicyConfig) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	p.nextToken()
	p.nextToken()
	p.pc = pc
	return p
}

func (p *Parser) nextToken() {
	p.token = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) error(msg string) {
	p.errors = append(p.errors, msg)
}

func (p *Parser) Parse() *Node {
	node := p.parseExpression()
	if p.token.Type != TokenEOF {
		p.error("expected EOF")
	}
	if len(p.errors) > 0 {
		fmt.Printf("Parser errors:\n%s\n", strings.Join(p.errors, "\n"))
		return nil
	}
	return node
}

func (p *Parser) parseExpression() *Node {
	return p.parseOR()
}

func (p *Parser) parseOR() *Node {
	node := p.parseAND()

	for p.token.Type == TokenOR {
		p.nextToken()
		right := p.parseAND()
		node = &Node{
			Type:     NodeOR,
			Children: []*Node{node, right},
		}
	}

	return node
}

func (p *Parser) parseAND() *Node {
	node := p.parsePrimary()

	for p.token.Type == TokenAND {
		p.nextToken()
		right := p.parsePrimary()
		node = &Node{
			Type:     NodeAND,
			Children: []*Node{node, right},
		}
	}

	return node
}

func (p *Parser) parsePrimary() *Node {
	switch p.token.Type {
	case TokenIdent:
		node := &Node{
			Type:  NodeIdent,
			Value: p.pc.ResolvePurpose(p.token.Value),
		}
		p.nextToken()
		return node
	case TokenLParen:
		p.nextToken()
		expr := p.parseExpression()
		if p.token.Type != TokenRParen {
			p.error("expected ')'")
			return nil
		}
		p.nextToken()
		return expr
	default:
		p.error(fmt.Sprintf("unexpected token: %v", p.token))
		return nil
	}
}

func toAttr(purposes string, policyConfig utils.PolicyConfig) {
	parser := NewParser(purposes, policyConfig)
	ast := parser.Parse()

	fmt.Printf("Original expression: %s\n", purposes)
	fmt.Printf("AST: %v\n", ast)
}
