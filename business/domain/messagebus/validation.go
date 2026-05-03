package messagebus

import (
	"fmt"
	"html/template"
	"sort"
	"strings"
	"text/template/parse"
)

var allowedTemplateRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Campaign":     {},
}

func validateAndExtractRequiredVars(content []byte) ([]string, error) {
	tmpl, err := template.New("message").Parse(string(content))
	if err != nil {
		return nil, unsupportedTemplateSyntax(err)
	}

	for _, namedTemplate := range tmpl.Templates() {
		if namedTemplate.Name() != tmpl.Name() {
			return nil, unsupportedTemplateSyntax(fmt.Errorf("named templates are not allowed"))
		}
	}

	requiredVars := make(map[string]struct{})
	if err := validateTemplateNode(tmpl.Tree.Root, requiredVars); err != nil {
		return nil, err
	}

	vars := make([]string, 0, len(requiredVars))
	for requiredVar := range requiredVars {
		vars = append(vars, requiredVar)
	}

	sort.Strings(vars)

	return vars, nil
}

func validateTemplateNode(node parse.Node, requiredVars map[string]struct{}) error {
	switch typedNode := node.(type) {
	case nil:
		return nil

	case *parse.ListNode:
		for _, child := range typedNode.Nodes {
			if err := validateTemplateNode(child, requiredVars); err != nil {
				return err
			}
		}
		return nil

	case *parse.TextNode, *parse.CommentNode:
		return nil

	case *parse.ActionNode:
		requiredVar, err := actionRequiredVar(typedNode)
		if err != nil {
			return err
		}

		requiredVars[requiredVar] = struct{}{}
		return nil

	default:
		return unsupportedTemplateSyntax(fmt.Errorf("node %T is not allowed", node))
	}
}

func actionRequiredVar(node *parse.ActionNode) (string, error) {
	if node.Pipe == nil {
		return "", unsupportedTemplateSyntax(fmt.Errorf("action must contain a field path"))
	}

	if len(node.Pipe.Decl) > 0 {
		return "", unsupportedTemplateSyntax(fmt.Errorf("variables are not allowed"))
	}

	if len(node.Pipe.Cmds) != 1 {
		return "", unsupportedTemplateSyntax(fmt.Errorf("pipelines are not allowed"))
	}

	command := node.Pipe.Cmds[0]
	if len(command.Args) != 1 {
		return "", unsupportedTemplateSyntax(fmt.Errorf("functions and complex commands are not allowed"))
	}

	fieldNode, ok := command.Args[0].(*parse.FieldNode)
	if !ok {
		return "", unsupportedTemplateSyntax(fmt.Errorf("action must be a simple field substitution"))
	}

	if len(fieldNode.Ident) < 2 {
		return "", unsupportedTemplateSyntax(fmt.Errorf("field path must contain at least two segments"))
	}

	if _, ok := allowedTemplateRoots[fieldNode.Ident[0]]; !ok {
		return "", unsupportedTemplateSyntax(fmt.Errorf("root namespace %q is not allowed", fieldNode.Ident[0]))
	}

	for _, ident := range fieldNode.Ident {
		if ident == "" {
			return "", unsupportedTemplateSyntax(fmt.Errorf("field path contains an empty segment"))
		}
	}

	return strings.Join(fieldNode.Ident, "."), nil
}

func unsupportedTemplateSyntax(err error) error {
	return fmt.Errorf("%w: %v", ErrUnsupportedTemplateSyntax, err)
}
