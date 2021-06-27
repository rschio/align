package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

type Node struct {
	ID    string
	Label string
}

type Edge struct {
	From, To string
}

type GraphBuilder struct {
	Nodes []*Node
	Edges []*Edge
}

type Graph struct {
	// Nodes store node labels, the index is the node ID
	// and the value is the node's letter.
	Nodes []rune
	Edges [][2]int
}

func Parse(r io.Reader) (*Graph, error) {
	gb, err := parse(r)
	if err != nil {
		return nil, err
	}
	return gb.Build()
}

func (g *GraphBuilder) Build() (*Graph, error) {
	uniqNodes := make(map[string]int)
	uniqLabels := make(map[rune]struct{})
	for i, n := range g.Nodes {
		if _, exist := uniqNodes[n.ID]; exist {
			return nil, fmt.Errorf("duplicated node: %s", n.ID)
		}
		uniqNodes[n.ID] = i
		if l := utf8.RuneCountInString(n.Label); l != 1 {
			return nil, fmt.Errorf("label must have 1 rune, got: %d", l)
		}
		r, _ := utf8.DecodeRuneInString(n.Label)
		uniqLabels[r] = struct{}{}
	}
	nodes := make([]rune, len(g.Nodes))
	for i, n := range g.Nodes {
		r, _ := utf8.DecodeRuneInString(n.Label)
		nodes[i] = r
	}
	edges := make([][2]int, len(g.Edges))
	for i, e := range g.Edges {
		edges[i][0] = uniqNodes[e.From]
		edges[i][1] = uniqNodes[e.To]
	}
	return &Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func parseNode(bs []byte) (*Node, error) {
	var err = fmt.Errorf("invaild node: %s", bs)
	// node: (a,b) or (a,).
	if len(bs) < 4 {
		return nil, err
	}
	if '(' != bs[0] || bs[len(bs)-1] != ')' {
		return nil, err
	}
	bs = bs[1 : len(bs)-1]
	parts := bytes.SplitN(bs, []byte(","), 2)
	if len(parts) != 2 {
		return nil, err
	}
	id := bytes.TrimSpace(parts[0])
	label := bytes.TrimSpace(parts[1])
	return &Node{ID: string(id), Label: string(label)}, nil
}

func parseEdge(bs []byte) (*Edge, error) {
	var err = fmt.Errorf("invaild edge: %s", bs)
	// edge: {a,b}.
	if len(bs) < 4 {
		return nil, err
	}
	if '{' != bs[0] || bs[len(bs)-1] != '}' {
		return nil, err
	}
	bs = bs[1 : len(bs)-1]
	parts := bytes.Split(bs, []byte(","))
	if len(parts) != 2 {
		return nil, err
	}
	from := bytes.TrimSpace(parts[0])
	to := bytes.TrimSpace(parts[1])
	return &Edge{From: string(from), To: string(to)}, nil
}

func parseComment(bs []byte) (string, error) {
	var err = fmt.Errorf("invalid comment: %s", bs)
	if len(bs) < 2 {
		return "", err
	}
	if bs[1] != '/' {
		return "", err
	}
	return string(bs), nil
}

func parse(r io.Reader) (*GraphBuilder, error) {
	sc := bufio.NewScanner(r)
	gb := new(GraphBuilder)
	gb.Nodes = make([]*Node, 0, 20)
	gb.Edges = make([]*Edge, 0, 20)
	for sc.Scan() {
		line := sc.Bytes()
		bs := bytes.TrimSpace(line)
		if len(bs) == 0 {
			continue
		}
		switch bs[0] {
		case '(':
			n, err := parseNode(bs)
			if err != nil {
				return nil, err
			}
			gb.Nodes = append(gb.Nodes, n)
		case '{':
			e, err := parseEdge(bs)
			if err != nil {
				return nil, err
			}
			gb.Edges = append(gb.Edges, e)
		case '/':
			_, err := parseComment(bs)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("invalid line: %s", line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return gb, nil
}
