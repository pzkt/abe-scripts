package utils

type PolicyConfig struct {
	PurposeTrees []*Tree
}

type Tree struct {
	Parent   *Tree
	Children []*Tree
	Value    string
}

func NewTree(rootValue string) *Tree {
	t := new(Tree)
	t.Value = rootValue
	return t
}

func (p PolicyConfig) ResolvePurpose(purpose string) []string {
	out := []string{}
	for _, pt := range p.PurposeTrees {
		node, found := pt.findValue(purpose)
		if !found {
			continue
		}
		out = append(out, node.getRootPath()...)
	}
	return out
}

func (t Tree) findValue(value string) (Tree, bool) {
	if t.Value == value {
		return t, true
	}
	for _, ct := range t.Children {
		node, found := ct.findValue(value)
		if found {
			return node, true
		}
	}
	return t, false
}

func (t Tree) getRootPath() []string {
	if t.Parent == nil {
		return []string{t.Value}
	}
	return append(t.Parent.getRootPath(), t.Value)
}

func (t *Tree) AddChild(value string) *Tree {
	child := new(Tree)
	t.Children = append(t.Children, child)
	child.Value = value
	child.Parent = t
	return child
}

func ExamplePolicyConfig() PolicyConfig {
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

	return PolicyConfig{PurposeTrees: []*Tree{firstTree, secondTree}}
}
