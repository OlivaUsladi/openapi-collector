package output

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func MarshalYAML(doc map[string]any) ([]byte, error) {
	node, err := toNode(doc, "")
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	enc := yaml.NewEncoder(&sb)
	enc.SetIndent(2)
	err = enc.Encode(node)
	if err != nil {
		return nil, fmt.Errorf("сериализация YAML: %w", err)
	}
	err = enc.Close()
	if err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// превращает данные в YAML узлы
func toNode(value any, docPath string) (*yaml.Node, error) {

	dataMap, ok := value.(map[string]any)
	if ok {
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

		orderRule := orderFor(docPath)
		keys := sortedKeys(dataMap, orderRule)

		for _, key := range keys {
			keyNode := scalarNode(key)

			childPath := docPath + "." + key
			if docPath == "" {
				childPath = key
			}

			valueNode, err := toNode(dataMap[key], childPath)
			if err != nil {
				return nil, err
			}

			node.Content = append(node.Content, keyNode, valueNode)
		}

		return node, nil
	}

	dataArray, ok := value.([]any)
	if ok {
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}

		for index, item := range dataArray {
			itemPath := fmt.Sprintf("%s[%d]", docPath, index)

			itemNode, err := toNode(item, itemPath)
			if err != nil {
				return nil, err
			}

			node.Content = append(node.Content, itemNode)
		}

		return node, nil
	}

	str, ok := value.(string)
	if ok {
		return scalarNode(str), nil
	}

	if value == nil {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	}

	node := &yaml.Node{}
	err := node.Encode(value)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func scalarNode(s string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
	if needsQuoting(s) {
		node.Style = yaml.DoubleQuotedStyle
	}
	return node
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	var parsed any
	err := yaml.Unmarshal([]byte(s), &parsed)
	if err != nil {
		return true
	}
	str, ok := parsed.(string)
	return !ok || str != s
}
