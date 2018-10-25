// Copyright (c) 2012-2017 The ANTLR Project. All rights reserved.
// Use of this file is governed by the BSD 3-clause license that
// can be found in the LICENSE.txt file in the project root.

package antlr

// The basic notion of a tree has a parent, a payload, and a list of children.
//  It is the most abstract interface for all the trees used by ANTLR.
///

var TreeInvalidInterval = NewInterval(-1, -2)

type Tree interface {
	GetParent() Tree
	SetParent(Tree)
	GetPayload() interface{}
	GetChild(i int) Tree
	GetChildCount() int
	GetChildren() []Tree
}

type SyntaxTree interface {
	Tree

	GetSourceInterval() *Interval
}

type ParseTree interface {
	SyntaxTree

	Accept(Visitor ParseTreeVisitor) interface{}
	GetText() string

	ToStringTree([]string, Recognizer) string
}

type RuleNode interface {
	ParseTree

	GetRuleContext() RuleContext
	GetBaseRuleContext() *BaseRuleContext
}

type TerminalNode interface {
	ParseTree

	GetSymbol() Token
}

type ErrorNode interface {
	TerminalNode

	errorNode()
}

type ParseTreeVisitor interface {
	Init() interface{}
	VisitNext(node Tree, resultSoFar interface{}) bool
	Aggregate(resultSoFar, childResult interface{}) interface{}
	VisitTerminal(node TerminalNode) interface{}
	VisitErrorNode(node ErrorNode) interface{}
	VisitChildren(node RuleNode, delegate ParseTreeVisitor) interface{}
}

type BaseParseTreeVisitor struct{}

func (*BaseParseTreeVisitor) Init() interface{}                                 { return nil }
func (*BaseParseTreeVisitor) VisitNext(node Tree, resultSoFar interface{}) bool { return true }
func (*BaseParseTreeVisitor) Aggregate(resultSoFar, childResult interface{}) interface{} {
	return childResult
}
func (*BaseParseTreeVisitor) VisitTerminal(node TerminalNode) interface{} { return nil }
func (*BaseParseTreeVisitor) VisitErrorNode(node ErrorNode) interface{}   { return nil }
func (*BaseParseTreeVisitor) VisitChildren(node RuleNode, delegate ParseTreeVisitor) interface{} {
	result := delegate.Init()
	for _, child := range node.GetChildren() {
		if !delegate.VisitNext(child, result) {
			continue
		}
		switch child := child.(type) {
		case TerminalNode:
			childResult := delegate.VisitTerminal(child)
			result = delegate.Aggregate(result, childResult)
		case ErrorNode:
			delegate.VisitErrorNode(child)
		case RuleNode:
			childResult := child.Accept(delegate)
			result = delegate.Aggregate(result, childResult)
		default:
			// can this happen??
		}
	}
	return result
}

// type ParseTreeVisitor interface {
// 	Visit(tree ParseTree) interface{}
// 	VisitChildren(node RuleNode) interface{}
// 	VisitTerminal(node TerminalNode) interface{}
// 	VisitErrorNode(node ErrorNode) interface{}
// }

// type BaseParseTreeVisitor struct{}

// var _ ParseTreeVisitor = &BaseParseTreeVisitor{}

// func (v *BaseParseTreeVisitor) Visit(tree ParseTree) interface{}            { return nil }
// func (v *BaseParseTreeVisitor) VisitChildren(node RuleNode) interface{}     { return nil }
// func (v *BaseParseTreeVisitor) VisitTerminal(node TerminalNode) interface{} { return nil }
// func (v *BaseParseTreeVisitor) VisitErrorNode(node ErrorNode) interface{}   { return nil }

// // // TODO
// // func (this ParseTreeVisitor) Visit(ctx) {
// // 	if (Utils.isArray(ctx)) {
// // 		self := this
// // 		return ctx.map(func(child) { return VisitAtom(self, child)})
// // 	} else {
// // 		return VisitAtom(this, ctx)
// // 	}
// // }

// // func VisitAtom(Visitor, ctx) {
// // 	if (ctx.parser == nil) { //is terminal
// // 		return
// // 	}

// // 	name := ctx.parser.ruleNames[ctx.ruleIndex]
// // 	funcName := "Visit" + Utils.titleCase(name)

// // 	return Visitor[funcName](ctx)
// // }

type ParseTreeListener interface {
	VisitTerminal(node TerminalNode)
	VisitErrorNode(node ErrorNode)
	EnterEveryRule(ctx ParserRuleContext)
	ExitEveryRule(ctx ParserRuleContext)
}

type BaseParseTreeListener struct{}

var _ ParseTreeListener = &BaseParseTreeListener{}

func (l *BaseParseTreeListener) VisitTerminal(node TerminalNode)      {}
func (l *BaseParseTreeListener) VisitErrorNode(node ErrorNode)        {}
func (l *BaseParseTreeListener) EnterEveryRule(ctx ParserRuleContext) {}
func (l *BaseParseTreeListener) ExitEveryRule(ctx ParserRuleContext)  {}

type TerminalNodeImpl struct {
	parentCtx RuleContext

	symbol Token
}

var _ TerminalNode = &TerminalNodeImpl{}

func NewTerminalNodeImpl(symbol Token) *TerminalNodeImpl {
	tn := new(TerminalNodeImpl)

	tn.parentCtx = nil
	tn.symbol = symbol

	return tn
}

func (t *TerminalNodeImpl) GetChild(i int) Tree {
	return nil
}

func (t *TerminalNodeImpl) GetChildren() []Tree {
	return nil
}

func (t *TerminalNodeImpl) SetChildren(tree []Tree) {
	panic("Cannot set children on terminal node")
}

func (t *TerminalNodeImpl) GetSymbol() Token {
	return t.symbol
}

func (t *TerminalNodeImpl) GetParent() Tree {
	return t.parentCtx
}

func (t *TerminalNodeImpl) SetParent(tree Tree) {
	t.parentCtx = tree.(RuleContext)
}

func (t *TerminalNodeImpl) GetPayload() interface{} {
	return t.symbol
}

func (t *TerminalNodeImpl) GetSourceInterval() *Interval {
	if t.symbol == nil {
		return TreeInvalidInterval
	}
	tokenIndex := t.symbol.GetTokenIndex()
	return NewInterval(tokenIndex, tokenIndex)
}

func (t *TerminalNodeImpl) GetChildCount() int {
	return 0
}

func (t *TerminalNodeImpl) Accept(v ParseTreeVisitor) interface{} {
	return v.VisitTerminal(t)
}

func (t *TerminalNodeImpl) GetText() string {
	return t.symbol.GetText()
}

func (t *TerminalNodeImpl) String() string {
	if t.symbol.GetTokenType() == TokenEOF {
		return "<EOF>"
	}

	return t.symbol.GetText()
}

func (t *TerminalNodeImpl) ToStringTree(s []string, r Recognizer) string {
	return t.String()
}

// Represents a token that was consumed during reSynchronization
// rather than during a valid Match operation. For example,
// we will create this kind of a node during single token insertion
// and deletion as well as during "consume until error recovery set"
// upon no viable alternative exceptions.

type ErrorNodeImpl struct {
	*TerminalNodeImpl
}

var _ ErrorNode = &ErrorNodeImpl{}

func NewErrorNodeImpl(token Token) *ErrorNodeImpl {
	en := new(ErrorNodeImpl)
	en.TerminalNodeImpl = NewTerminalNodeImpl(token)
	return en
}

func (e *ErrorNodeImpl) errorNode() {}

func (e *ErrorNodeImpl) Accept(v ParseTreeVisitor) interface{} {
	return v.VisitErrorNode(e)
}

type ParseTreeWalker struct {
}

func NewParseTreeWalker() *ParseTreeWalker {
	return new(ParseTreeWalker)
}

func (p *ParseTreeWalker) Walk(listener ParseTreeListener, t Tree) {
	switch tt := t.(type) {
	case ErrorNode:
		listener.VisitErrorNode(tt)
	case TerminalNode:
		listener.VisitTerminal(tt)
	default:
		p.EnterRule(listener, t.(RuleNode))
		for i := 0; i < t.GetChildCount(); i++ {
			child := t.GetChild(i)
			p.Walk(listener, child)
		}
		p.ExitRule(listener, t.(RuleNode))
	}
}

//
// The discovery of a rule node, involves sending two events: the generic
// {@link ParseTreeListener//EnterEveryRule} and a
// {@link RuleContext}-specific event. First we trigger the generic and then
// the rule specific. We to them in reverse order upon finishing the node.
//
func (p *ParseTreeWalker) EnterRule(listener ParseTreeListener, r RuleNode) {
	ctx := r.GetRuleContext().(ParserRuleContext)
	listener.EnterEveryRule(ctx)
	ctx.EnterRule(listener)
}

func (p *ParseTreeWalker) ExitRule(listener ParseTreeListener, r RuleNode) {
	ctx := r.GetRuleContext().(ParserRuleContext)
	ctx.ExitRule(listener)
	listener.ExitEveryRule(ctx)
}

var ParseTreeWalkerDefault = NewParseTreeWalker()
