package syntax

// Cursor provides stateful tree navigation over a green tree.
// Unlike the red tree's ephemeral wrappers, a Cursor maintains its
// position and supports directional movement (parent, first child,
// next sibling, previous sibling).
type Cursor struct {
	// stack tracks the path from root to current position.
	// Each entry is a (node, childIndex) pair.
	stack []cursorFrame
	// current is the element at the cursor's current position.
	current GreenElement
}

type cursorFrame struct {
	node  *GreenNode
	index int // which child of node the cursor is at
}

// NewCursor creates a cursor positioned at the root of the given green tree.
func NewCursor(root *GreenNode) *Cursor {
	return &Cursor{
		current: NodeElement(root),
	}
}

// Current returns the element at the cursor's current position.
func (c *Cursor) Current() GreenElement {
	return c.current
}

// Kind returns the SyntaxKind of the current element.
func (c *Cursor) Kind() SyntaxKind {
	return c.current.Kind()
}

// IsNode reports whether the cursor is at a node.
func (c *Cursor) IsNode() bool {
	return c.current.IsNode()
}

// IsToken reports whether the cursor is at a token.
func (c *Cursor) IsToken() bool {
	return c.current.IsToken()
}

// Node returns the current node, or nil if at a token.
func (c *Cursor) Node() *GreenNode {
	return c.current.Node
}

// Token returns the current token, or nil if at a node.
func (c *Cursor) Token() *GreenToken {
	return c.current.Token
}

// Depth returns the depth of the cursor in the tree (root is 0).
func (c *Cursor) Depth() int {
	return len(c.stack)
}

// Parent moves the cursor to its parent node. Returns false if already at root.
func (c *Cursor) Parent() bool {
	if len(c.stack) == 0 {
		return false
	}
	frame := c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
	c.current = NodeElement(frame.node)
	return true
}

// FirstChild moves the cursor to the first child of the current node.
// Returns false if the current element is a token or has no children.
func (c *Cursor) FirstChild() bool {
	if c.current.Node == nil || len(c.current.Node.Children) == 0 {
		return false
	}
	node := c.current.Node
	c.stack = append(c.stack, cursorFrame{node: node, index: 0})
	c.current = node.Children[0]
	return true
}

// LastChild moves the cursor to the last child of the current node.
// Returns false if the current element is a token or has no children.
func (c *Cursor) LastChild() bool {
	if c.current.Node == nil || len(c.current.Node.Children) == 0 {
		return false
	}
	node := c.current.Node
	lastIdx := len(node.Children) - 1
	c.stack = append(c.stack, cursorFrame{node: node, index: lastIdx})
	c.current = node.Children[lastIdx]
	return true
}

// NextSibling moves the cursor to the next sibling. Returns false if at
// the last child or at root level.
func (c *Cursor) NextSibling() bool {
	if len(c.stack) == 0 {
		return false
	}
	frame := &c.stack[len(c.stack)-1]
	nextIdx := frame.index + 1
	if nextIdx >= len(frame.node.Children) {
		return false
	}
	frame.index = nextIdx
	c.current = frame.node.Children[nextIdx]
	return true
}

// PrevSibling moves the cursor to the previous sibling. Returns false if at
// the first child or at root level.
func (c *Cursor) PrevSibling() bool {
	if len(c.stack) == 0 {
		return false
	}
	frame := &c.stack[len(c.stack)-1]
	if frame.index <= 0 {
		return false
	}
	frame.index--
	c.current = frame.node.Children[frame.index]
	return true
}

// GotoChild moves the cursor to the n-th child of the current node.
// Returns false if the index is out of range or current element is a token.
func (c *Cursor) GotoChild(n int) bool {
	if c.current.Node == nil {
		return false
	}
	node := c.current.Node
	if n < 0 || n >= len(node.Children) {
		return false
	}
	c.stack = append(c.stack, cursorFrame{node: node, index: n})
	c.current = node.Children[n]
	return true
}

// Reset moves the cursor back to the root.
func (c *Cursor) Reset(root *GreenNode) {
	c.stack = c.stack[:0]
	c.current = NodeElement(root)
}
