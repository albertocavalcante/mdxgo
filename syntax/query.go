package syntax

// FindAll returns all descendant nodes matching the predicate.
func FindAll(root *GreenNode, pred func(*GreenNode) bool) []*GreenNode {
	var results []*GreenNode
	Walk(root, func(n *GreenNode, depth int) WalkAction {
		if pred(n) {
			results = append(results, n)
		}
		return Continue
	})
	return results
}

// FindFirst returns the first descendant node matching the predicate, or nil.
func FindFirst(root *GreenNode, pred func(*GreenNode) bool) *GreenNode {
	var result *GreenNode
	Walk(root, func(n *GreenNode, depth int) WalkAction {
		if pred(n) {
			result = n
			return Stop
		}
		return Continue
	})
	return result
}

// FindToken returns the first token in the tree matching the predicate, or nil.
func FindToken(root *GreenNode, pred func(*GreenToken) bool) *GreenToken {
	var result *GreenToken
	WalkAll(root, func(elem WalkElement, depth int) WalkAllAction {
		if elem.IsToken() && pred(elem.Token) {
			result = elem.Token
			return StopAll
		}
		return ContinueAll
	})
	return result
}

// FindAllTokens returns all tokens in the tree matching the predicate.
func FindAllTokens(root *GreenNode, pred func(*GreenToken) bool) []*GreenToken {
	var results []*GreenToken
	WalkAll(root, func(elem WalkElement, depth int) WalkAllAction {
		if elem.IsToken() && pred(elem.Token) {
			results = append(results, elem.Token)
		}
		return ContinueAll
	})
	return results
}

// CountNodes returns the number of descendant nodes matching the predicate.
func CountNodes(root *GreenNode, pred func(*GreenNode) bool) int {
	count := 0
	Walk(root, func(n *GreenNode, depth int) WalkAction {
		if pred(n) {
			count++
		}
		return Continue
	})
	return count
}
