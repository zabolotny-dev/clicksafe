package tmpl

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"text/template/parse"
)

func extractHTMLVars(content []byte, allowedRoots map[string]struct{}) ([]string, error) {
	tmpl, err := template.New("content").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	for _, namedTemplate := range tmpl.Templates() {
		if namedTemplate.Name() != tmpl.Name() {
			return nil, fmt.Errorf("validate template: %w: named templates are not allowed", ErrUnsupportedTemplateSyntax)
		}
	}

	requiredVars := make(map[string]struct{})
	if err := validateNode(tmpl.Tree.Root, allowedRoots, requiredVars); err != nil {
		return nil, fmt.Errorf("validate template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	vars := make([]string, 0, len(requiredVars))
	for v := range requiredVars {
		vars = append(vars, v)
	}

	sort.Strings(vars)

	return vars, nil
}

func renderHTML(content []byte, data map[string]any) ([]byte, error) {
	tmpl, err := template.New("message").Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("execute template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	return out.Bytes(), nil
}

func validateNode(node parse.Node, allowedRoots map[string]struct{}, requiredVars map[string]struct{}) error {
	switch typedNode := node.(type) {
	case nil:
		return nil

	case *parse.ListNode:
		for _, child := range typedNode.Nodes {
			if err := validateNode(child, allowedRoots, requiredVars); err != nil {
				return err
			}
		}
		return nil

	case *parse.TextNode, *parse.CommentNode:
		return nil

	case *parse.ActionNode:
		requiredVar, err := actionRequiredVar(typedNode, allowedRoots)
		if err != nil {
			return err
		}

		requiredVars[requiredVar] = struct{}{}
		return nil

	default:
		return fmt.Errorf("node %T is not allowed", node)
	}
}

func actionRequiredVar(node *parse.ActionNode, allowedRoots map[string]struct{}) (string, error) {
	if node.Pipe == nil {
		return "", fmt.Errorf("action must contain a field path")
	}

	if len(node.Pipe.Decl) > 0 {
		return "", fmt.Errorf("variables are not allowed")
	}

	if len(node.Pipe.Cmds) != 1 {
		return "", fmt.Errorf("pipelines are not allowed")
	}

	command := node.Pipe.Cmds[0]
	if len(command.Args) != 1 {
		return "", fmt.Errorf("functions and complex commands are not allowed")
	}

	fieldNode, ok := command.Args[0].(*parse.FieldNode)
	if !ok {
		return "", fmt.Errorf("action must be a simple field substitution")
	}

	if len(fieldNode.Ident) < 2 {
		return "", fmt.Errorf("field path must contain at least two segments")
	}

	if _, ok := allowedRoots[fieldNode.Ident[0]]; !ok {
		return "", fmt.Errorf("root namespace %q is not allowed", fieldNode.Ident[0])
	}

	for _, ident := range fieldNode.Ident {
		if ident == "" {
			return "", fmt.Errorf("field path contains an empty segment")
		}
	}

	return strings.Join(fieldNode.Ident, "."), nil
}
