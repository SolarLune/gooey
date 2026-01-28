package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type TemplateField struct {
	Name string
	Type string
}

func (t TemplateField) IfCondition() string {

	typeIs := func(set []string) bool {
		for _, v := range set {
			if v == t.Type {
				return true
			} else if strings.HasPrefix(t.Type, v) {
				return true
			}
		}
		return false
	}

	primitiveNum := []string{
		"int",
		"int32",
		"int16",
		"int64",
		"uint32",
		"uint64",
		"float32",
		"float64",
		"Alignment",
		"StretchMode",
	}

	nilable := []string{
		"[]", // Slice notation followed by any element
		"*",  // A pointer followed by any element
		"map[",
		"func",
		"UIElement",
		"ArrangeFunc",
	}

	base := []string{
		"	if other.{{.}}{",
		"		s.{{.}} = other.{{.}}",
		"	}",
	}

	if t.Type == "bool" {
		base[0] = `	if other.{{.}} {`
	} else if t.Type == "string" {
		base[0] = `	if other.{{.}} != "" {`
	} else if typeIs(nilable) {
		base[0] = `	if other.{{.}} != nil {`
	} else if typeIs(primitiveNum) {
		base[0] = `	if other.{{.}} != 0 {`
	} else { // Assume it's zeroable?
		base[0] = `	if !other.{{.}}.IsZero() {`
	}

	out := strings.ReplaceAll(strings.Join(base, "\n"), "{{.}}", t.Name)

	return out
}

func main() {

	out := "package gooey\n"

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		if strings.Contains(path, "/") {
			return nil
		}

		if !strings.HasPrefix(path, "ui") {
			return nil
		}

		byteData, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		type structFind struct {
			Start int
			End   int
			Name  string
		}
		structBounds := []structFind{}

		structStart := -1
		structEnd := -1
		structName := ""

		fileContents := strings.Split(string(byteData), "\n")

		for i, line := range fileContents {

			if strings.Contains(line, "struct {") && strings.Contains(line, "UI") {

				end := strings.Index(line, "struct {")
				structName = strings.TrimSpace(line[5:end])
				structStart = i

			}

			if strings.Contains(line, "}") && structStart >= 0 {
				structEnd = i
			}

			if structStart > -1 && structEnd > -1 {
				structBounds = append(structBounds, structFind{
					Start: structStart,
					End:   structEnd,
					Name:  structName,
				})

				structStart = -1
				structEnd = -1
			}

		}

		for _, s := range structBounds {

			fields := []TemplateField{}

			if s.Start >= 0 {

				for _, line := range fileContents[s.Start+1 : s.End] {

					if words := strings.Fields(line); len(words) > 0 {

						commentPoint := strings.Index(line, "//")

						if commentPoint >= 0 && commentPoint < strings.Index(line, words[1]) {
							continue
						}

						fields = append(fields, TemplateField{
							Name: words[0],
							Type: words[1],
						})
					}

				}
			}

			templ := template.New("applyTemplate")
			templ.Parse(`// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s {{.StructName}}) Apply(other {{.StructName}}) {{.StructName}} {
	{{range .Fields}}
{{.IfCondition}}
{{end}}
	return s
}`)

			result := new(bytes.Buffer)

			input := struct {
				StructName string
				Fields     []TemplateField
			}{
				StructName: s.Name,
				Fields:     fields,
			}

			err = templ.Execute(result, input)

			if err != nil {
				panic(err)
			}

			out += result.String() + "\n\n"

		}

		return nil
	})

	os.WriteFile("applyfunctions.go", []byte(out), 0644)

	// fmt.Println(out)

	// fileContents, err := os.ReadFile(*fp)
	// if err != nil {
	// 	panic(err)
	// }

	// results := string(fileContents)

	// fmt.Println(results)

}
