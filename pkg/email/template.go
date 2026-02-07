package email

import (
	"bytes"
	"text/template"
)

// Render 渲染模板
func Render(tpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("email").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
