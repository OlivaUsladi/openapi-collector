package output

import (
	"fmt"
	"strconv"
	"strings"
)

func MarshalJSON(doc map[string]any) ([]byte, error) {
	var sb strings.Builder
	err := writeJSON(&sb, doc, "", 0)
	if err != nil {
		return nil, err
	}
	sb.WriteString("\n")
	return []byte(sb.String()), nil
}

// превращает данные в форматированный JSON
func writeJSON(sb *strings.Builder, value any, docPath string, depth int) error {
	indent := strings.Repeat("  ", depth)
	childIndent := strings.Repeat("  ", depth+1)

	dictionary, ok := value.(map[string]any)
	if ok {
		if len(dictionary) == 0 {
			sb.WriteString("{}")
			return nil
		}

		sb.WriteString("{\n")

		keys := sortedKeys(dictionary, orderFor(docPath))

		for i := 0; i < len(keys); i++ {
			key := keys[i]

			childPath := key
			if docPath != "" {
				childPath = docPath + "." + key
			}

			sb.WriteString(childIndent)
			writeJSONString(sb, key)
			sb.WriteString(": ")

			err := writeJSON(sb, dictionary[key], childPath, depth+1)
			if err != nil {
				return err
			}

			if i < len(keys)-1 {
				sb.WriteString(",")
			}

			sb.WriteString("\n")
		}

		sb.WriteString(indent + "}")
		return nil
	}

	array, ok := value.([]any)
	if ok {
		if len(array) == 0 {
			sb.WriteString("[]")
			return nil
		}

		sb.WriteString("[\n")

		for i := 0; i < len(array); i++ {
			item := array[i]

			itemPath := fmt.Sprintf("%s[%d]", docPath, i)

			sb.WriteString(childIndent)

			err := writeJSON(sb, item, itemPath, depth+1)
			if err != nil {
				return err
			}

			if i < len(array)-1 {
				sb.WriteString(",")
			}

			sb.WriteString("\n")
		}

		sb.WriteString(indent + "]")
		return nil
	}

	text, ok := value.(string)
	if ok {
		writeJSONString(sb, text)
		return nil
	}

	flag, ok := value.(bool)
	if ok {
		if flag {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		return nil
	}

	number, ok := value.(int)
	if ok {
		sb.WriteString(strconv.Itoa(number))
		return nil
	}

	if number, ok := value.(int64); ok {
		sb.WriteString(strconv.FormatInt(number, 10))
		return nil
	}

	if number, ok := value.(float64); ok {
		sb.WriteString(strconv.FormatFloat(number, 'g', -1, 64))
		return nil
	}

	if value == nil {
		sb.WriteString("null")
		return nil
	}

	return fmt.Errorf("неожиданный тип: %T", value)
}

// записывает строку в JSON формате
func writeJSONString(sb *strings.Builder, s string) {
	sb.WriteByte('"')

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if ch == '"' {
			sb.WriteString(`\"`)
			continue
		}

		if ch == '\\' {
			sb.WriteString(`\\`)
			continue
		}

		if ch == '\n' {
			sb.WriteString(`\n`)
			continue
		}

		if ch == '\r' {
			sb.WriteString(`\r`)
			continue
		}

		if ch == '\t' {
			sb.WriteString(`\t`)
			continue
		}

		if ch < 32 {
			sb.WriteString(fmt.Sprintf(`\u%04x`, ch))
			continue
		}

		sb.WriteByte(ch)
	}

	sb.WriteByte('"')
}
