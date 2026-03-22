package syntax

import "iter"

// WalkAction controls the traversal behavior when walking a green tree.
type WalkAction int

const (
	// Continue proceeds with the traversal normally.
	Continue WalkAction = iota
	// SkipChildren skips the current node's children but continues with siblings.
	SkipChildren
	// Stop terminates the traversal immediately.
	Stop
)

// WalkFunc is the callback for Walk. It receives the current node and its
// depth in the tree (root is depth 0). Return a WalkAction to control traversal.
type WalkFunc func(node *GreenNode, depth int) WalkAction

// Walk performs a depth-first pre-order traversal of the green tree,
// calling fn for each composite node encountered. Tokens are not visited
// since they are leaves — use WalkAll to visit both nodes and tokens.
func Walk(root *GreenNode, fn WalkFunc) {
	walkRec(root, fn, 0)
}

func walkRec(n *GreenNode, fn WalkFunc, depth int) WalkAction {
	action := fn(n, depth)
	switch action {
	case Stop:
		return Stop
	case SkipChildren:
		return Continue
	}

	for _, child := range n.Children {
		if child.Node != nil {
			if walkRec(child.Node, fn, depth+1) == Stop {
				return Stop
			}
		}
	}
	return Continue
}

// WalkElement is used by WalkAll to represent either a node or token visit.
type WalkElement struct {
	Node  *GreenNode
	Token *GreenToken
}

// IsNode reports whether this element is a node.
func (e WalkElement) IsNode() bool { return e.Node != nil }

// IsToken reports whether this element is a token.
func (e WalkElement) IsToken() bool { return e.Token != nil }

// Kind returns the SyntaxKind of the element.
func (e WalkElement) Kind() SyntaxKind {
	if e.Token != nil {
		return e.Token.Kind
	}
	return e.Node.Kind
}

// WalkAllAction controls the traversal behavior for WalkAll.
type WalkAllAction int

const (
	// ContinueAll proceeds with the traversal normally.
	ContinueAll WalkAllAction = iota
	// SkipChildrenAll skips the current node's children.
	SkipChildrenAll
	// StopAll terminates the traversal.
	StopAll
)

// WalkAllFunc is the callback for WalkAll.
type WalkAllFunc func(elem WalkElement, depth int) WalkAllAction

// WalkAll performs a depth-first traversal visiting both nodes and tokens.
func WalkAll(root *GreenNode, fn WalkAllFunc) {
	walkAllRec(root, fn, 0)
}

func walkAllRec(n *GreenNode, fn WalkAllFunc, depth int) WalkAllAction {
	action := fn(WalkElement{Node: n}, depth)
	switch action {
	case StopAll:
		return StopAll
	case SkipChildrenAll:
		return ContinueAll
	}

	for _, child := range n.Children {
		if child.Token != nil {
			if fn(WalkElement{Token: child.Token}, depth+1) == StopAll {
				return StopAll
			}
		} else {
			if walkAllRec(child.Node, fn, depth+1) == StopAll {
				return StopAll
			}
		}
	}
	return ContinueAll
}

// Nodes returns an iterator over all descendant nodes (depth-first pre-order).
// This supports Go's range-over-func pattern.
func Nodes(root *GreenNode) iter.Seq[*GreenNode] {
	return func(yield func(*GreenNode) bool) {
		nodesRec(root, yield)
	}
}

func nodesRec(n *GreenNode, yield func(*GreenNode) bool) bool {
	if !yield(n) {
		return false
	}
	for _, child := range n.Children {
		if child.Node != nil {
			if !nodesRec(child.Node, yield) {
				return false
			}
		}
	}
	return true
}

// Tokens returns an iterator over all tokens in the tree (left-to-right order).
func Tokens(root *GreenNode) iter.Seq[*GreenToken] {
	return func(yield func(*GreenToken) bool) {
		tokensRec(root, yield)
	}
}

func tokensRec(n *GreenNode, yield func(*GreenToken) bool) bool {
	for _, child := range n.Children {
		if child.Token != nil {
			if !yield(child.Token) {
				return false
			}
		} else {
			if !tokensRec(child.Node, yield) {
				return false
			}
		}
	}
	return true
}

// NodesOfKind returns an iterator over nodes of a specific kind.
func NodesOfKind(root *GreenNode, kind SyntaxKind) iter.Seq[*GreenNode] {
	return func(yield func(*GreenNode) bool) {
		for n := range Nodes(root) {
			if n.Kind == kind {
				if !yield(n) {
					return
				}
			}
		}
	}
}

// TokensOfKind returns an iterator over tokens of a specific kind.
func TokensOfKind(root *GreenNode, kind SyntaxKind) iter.Seq[*GreenToken] {
	return func(yield func(*GreenToken) bool) {
		for tok := range Tokens(root) {
			if tok.Kind == kind {
				if !yield(tok) {
					return
				}
			}
		}
	}
}
