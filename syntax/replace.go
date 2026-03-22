package syntax

// ReplaceChild returns a new tree with the child at index childIdx of the
// node at the given path replaced with newChild.
//
// path is a sequence of child indices from the root to the target node.
// For example, path [2, 0] means "root.Children[2].Node.Children[0]".
//
// The function implements structural sharing: only the spine from the
// target node to the root is copied. All other nodes are shared with
// the original tree.
//
// Returns nil if the path is invalid.
func ReplaceChild(root *GreenNode, path []int, childIdx int, newChild GreenElement) *GreenNode {
	if len(path) == 0 {
		// Replace a direct child of root.
		if childIdx < 0 || childIdx >= len(root.Children) {
			return nil
		}
		return root.ReplaceChild(childIdx, newChild)
	}

	// Navigate down the path, building a stack of parent nodes and indices.
	current := root
	for i, idx := range path {
		if idx < 0 || idx >= len(current.Children) {
			return nil
		}
		child := current.Children[idx]
		if child.Node == nil {
			return nil // path points through a token, not a node
		}
		if i == len(path)-1 {
			// Replace the child in this node.
			if childIdx < 0 || childIdx >= len(child.Node.Children) {
				return nil
			}
			newNode := child.Node.ReplaceChild(childIdx, newChild)
			// Walk back up, replacing each parent.
			return replaceSpine(root, path, newNode)
		}
		current = child.Node
	}
	return nil
}

// ReplaceDescendant returns a new tree with the descendant node at the
// given path replaced entirely with newNode.
//
// path is a sequence of child indices from the root to the target.
// For example, path [1, 3] replaces root.Children[1].Node.Children[3].
//
// Structural sharing: only the spine is copied.
// Returns nil if the path is invalid or empty.
func ReplaceDescendant(root *GreenNode, path []int, newElement GreenElement) *GreenNode {
	if len(path) == 0 {
		return nil
	}

	if len(path) == 1 {
		idx := path[0]
		if idx < 0 || idx >= len(root.Children) {
			return nil
		}
		return root.ReplaceChild(idx, newElement)
	}

	// Navigate to parent, then replace the final child.
	parentPath := path[:len(path)-1]
	childIdx := path[len(path)-1]

	// Find the parent node.
	parent := root
	for _, idx := range parentPath {
		if idx < 0 || idx >= len(parent.Children) {
			return nil
		}
		child := parent.Children[idx]
		if child.Node == nil {
			return nil
		}
		parent = child.Node
	}

	if childIdx < 0 || childIdx >= len(parent.Children) {
		return nil
	}

	newParent := parent.ReplaceChild(childIdx, newElement)
	return replaceSpine(root, parentPath, newParent)
}

// replaceSpine walks back up the tree from the given path, replacing each
// node along the spine with a new version that contains the updated child.
// The leaf replacement is newLeaf, and path describes the indices from root
// to the parent of newLeaf.
func replaceSpine(root *GreenNode, path []int, newLeaf *GreenNode) *GreenNode {
	if len(path) == 0 {
		return newLeaf
	}

	// Build the spine bottom-up.
	current := newLeaf
	for i := len(path) - 1; i >= 0; i-- {
		idx := path[i]
		var parent *GreenNode
		if i == 0 {
			parent = root
		} else {
			// Navigate from root to path[:i] to find the parent.
			parent = root
			for j := 0; j < i; j++ {
				parent = parent.Children[path[j]].Node
			}
		}
		current = parent.ReplaceChild(idx, NodeElement(current))
	}
	return current
}

// FindPath returns the path (sequence of child indices) from root to the
// first descendant node whose kind matches the predicate. Returns nil if
// no match is found.
func FindPath(root *GreenNode, pred func(*GreenNode) bool) []int {
	return findPathRec(root, pred, nil)
}

func findPathRec(n *GreenNode, pred func(*GreenNode) bool, prefix []int) []int {
	for i, child := range n.Children {
		if child.Node != nil {
			if pred(child.Node) {
				return append(prefix, i)
			}
			result := findPathRec(child.Node, pred, append(prefix, i))
			if result != nil {
				return result
			}
		}
	}
	return nil
}
