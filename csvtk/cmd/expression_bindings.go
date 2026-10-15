package cmd

import (
	"math"
	"strconv"
	"strings"

	"github.com/casbin/govaluate"
)

// expressionBindings gives column references stable identifiers, allowing a
// compiled expression to receive differently typed values on successive rows.
type expressionBindings struct {
	source     string                 // Expression with column references replaced by identifiers.
	aliases    map[string]string      // Original column reference to generated identifier.
	values     map[string]interface{} // Reused evaluation environment for the current row.
	template   expressionTemplate     // Original locations for the literal-substitution fallback.
	exprEngine bool                   // Expr and govaluate use different string literal syntax.
	fallback   bool                   // Preserve literal-dependent expressions by rendering each row.
	literalRow bool                   // Whether the previous/current row uses literal substitution.
	recompile  bool                   // Source or input types changed since the preceding row.
	kinds      map[string]byte        // Last value type for Expr's static type checking.
	negative   map[string]bool        // References whose negative literals affect syntax or precedence.
}

// newExpressionBindings avoids identifier collisions with user expression text.
func newExpressionBindings(source string, parameters map[string]interface{}, exprEngine bool) *expressionBindings {
	prefix := "csvtkField"
	for strings.Contains(source, prefix) {
		prefix += "_"
	}
	aliases := make(map[string]string, len(parameters))
	for key := range parameters {
		aliases[key] = prefix + strconv.Itoa(len(aliases))
	}
	template := newExpressionTemplate(source, aliases)
	bindings := &expressionBindings{
		source: template.render(aliases), aliases: aliases, template: template,
		values: make(map[string]interface{}, len(parameters)+1), exprEngine: exprEngine,
		kinds:    make(map[string]byte, len(parameters)),
		negative: make(map[string]bool),
	}
	if reNullCoalescence.MatchString(source) {
		bindings.values["shenweiNULL"] = nil
	}
	// References inside quoted literals require the historical substitution
	// semantics. Dynamic govaluate regexes also retain compile-time validation.
	bindings.fallback = quotedExpressionReference(template, exprEngine) ||
		(!exprEngine && (strings.Contains(source, "=~") || strings.Contains(source, "!~")))
	for i, part := range template {
		if !part.variable {
			continue
		}
		if exprEngine {
			bindings.negative[part.text] = strings.Contains(source, "^") || strings.Contains(source, "**")
		} else if i > 0 {
			previous := template[i-1].text
			if (len(previous) > 0 && strings.ContainsRune("+-*/%><=&|^!~?:", rune(previous[len(previous)-1]))) ||
				strings.HasSuffix(strings.TrimSpace(previous), "-") {
				bindings.negative[part.text] = true
			}
		}
	}
	return bindings
}

// prepare updates values without changing the compiled source in the common path.
func (b *expressionBindings) prepare(parameters map[string]interface{}) (string, map[string]interface{}, error) {
	b.recompile = b.literalRow
	b.literalRow = b.fallback
	for key, value := range parameters {
		if b.literalRow {
			break
		}
		// A minus sign inserted next to an operator can change parsing;
		// in Expr it can also change exponentiation precedence.
		if number, ok := value.(float64); ok && math.Signbit(number) && b.negative[key] {
			b.literalRow = true
			break
		}
		if !b.exprEngine {
			// TIME tokens have special syntax restrictions in govaluate. Keep
			// its original parsing and validation for rows containing dates.
			if text, ok := value.(string); ok && strings.ContainsAny(text, ":-") {
				literal, err := govaluate.NewEvaluableExpression(quoteGovaluateString(text))
				if err != nil {
					return "", nil, err
				}
				if literal.Tokens()[0].Kind == govaluate.TIME {
					b.literalRow = true
					break
				}
			}
		} else {
			// Expr checks each concrete type combination, as literal parsing
			// did. Recompile only when a field changes between number, text, or nil.
			kind := byte(1)
			switch value.(type) {
			case float64:
				kind = 2
			case string:
				kind = 3
			}
			if b.kinds[key] != kind {
				b.recompile = true
				b.kinds[key] = kind
			}
		}
		b.values[b.aliases[key]] = value
	}
	if b.literalRow {
		b.recompile = true
		literals := make(map[string]string, len(parameters))
		for key, value := range parameters {
			literals[key] = expressionLiteral(value, b.exprEngine)
		}
		return b.template.render(literals), b.values, nil
	}
	return b.source, b.values, nil
}

// expressionLiteral preserves the old parser behavior for uncommon fallback expressions.
func expressionLiteral(value interface{}, exprEngine bool) string {
	switch value := value.(type) {
	case float64:
		return formatExpressionFloat(value)
	case string:
		if exprEngine {
			return strconv.Quote(value)
		}
		return quoteGovaluateString(value)
	default:
		return "shenweiNULL"
	}
}

// quotedExpressionReference detects substitutions whose meaning depends on quotes.
func quotedExpressionReference(template expressionTemplate, exprEngine bool) bool {
	var quote byte
	var escaped bool
	for _, part := range template {
		if part.variable {
			if quote != 0 {
				return true
			}
			continue
		}
		for i := 0; i < len(part.text); i++ {
			c := part.text[i]
			if escaped {
				escaped = false
			} else if quote != 0 && c == '\\' {
				escaped = true
			} else if c == quote {
				quote = 0
			} else if quote == 0 && (c == '\'' || c == '"' || c == '`') {
				quote = c
			} else if quote == 0 && c == '[' && !exprEngine {
				// Brackets quote variable names in govaluate, rather than arrays.
				quote = ']'
			}
		}
	}
	return false
}
