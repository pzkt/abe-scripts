/*

functions for the purpose hierarchy tree struct

*/

package utils

import "strings"

type Tree struct {
	Parent   *Tree   `json:"-"`
	Children []*Tree `json:"children"`
	Value    string  `json:"value"`
}

func NewTree(rootValue string) *Tree {
	t := new(Tree)
	t.Value = rootValue
	return t
}

func (t Tree) FindValue(value string) (Tree, bool) {
	if t.Value == value {
		return t, true
	}
	for _, ct := range t.Children {
		node, found := ct.FindValue(value)
		if found {
			return node, true
		}
	}
	return t, false
}

func (t Tree) GetRootPath() []string {
	if t.Parent == nil {
		return []string{t.Value}
	}
	return append(t.Parent.GetRootPath(), t.Value)
}

func (t *Tree) ReconnectParents(p *Tree) {
	t.Parent = p
	for _, c := range t.Children {
		c.ReconnectParents(t)
	}
}

func (t *Tree) DisconnectParents() {
	t.Parent = nil
	for _, c := range t.Children {
		c.DisconnectParents()
	}
}

func (t *Tree) AddChild(value string) *Tree {
	child := new(Tree)
	t.Children = append(t.Children, child)
	child.Value = value
	child.Parent = t
	return child
}

func ExamplePurposeTrees() []*Tree {
	firstTree := NewTree("General-Purpose")
	firstTree.AddChild("Purchase")
	firstTree.AddChild("Shipping")
	admin := firstTree.AddChild("Admin")
	admin.AddChild("Profiling")
	admin.AddChild("Analysis")
	marketing := firstTree.AddChild("Marketing")
	marketing.AddChild("Direct")
	thirdParty := marketing.AddChild("Third-Party")
	thirdParty.AddChild("Email")
	thirdParty.AddChild("Phone")

	secondTree := NewTree("General-Purpose")
	healthRecord := secondTree.AddChild("Health-Record")
	healthRecord.AddChild("Optometry")
	healthRecord.AddChild("Radiology")
	healthRecord.AddChild("Need-To-Know")
	research := secondTree.AddChild("Research")
	research.AddChild("Anonymized-Research")
	research.AddChild("Masked-Research")

	return []*Tree{firstTree, secondTree}
}

func (t *Tree) String() string {
	var builder strings.Builder
	t.stringHelper(&builder, 0)
	return builder.String()
}

func (t *Tree) stringHelper(b *strings.Builder, depth int) {
	if t == nil {
		return
	}

	b.WriteString(strings.Repeat("-", depth))
	b.WriteString(t.Value)
	b.WriteByte('\n')

	// Recursively print children
	for _, child := range t.Children {
		child.stringHelper(b, depth+1)
	}
}
