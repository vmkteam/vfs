package vfs

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/semrush/zenrpc/smd"
)

const (
	interfacePrefix   = "I"
	definitionsPrefix = "#/definitions/"
	voidResponse      = "void"
)

type tsInterface struct {
	Name       string
	Parameters []tsType
}

type tsType struct {
	Name     string
	Comment  string
	Type     string
	Optional bool
}

type tsServiceNamespace struct {
	Name     string
	Services []tsService
}

type tsService struct {
	Namespace string
	Name      string
	NameLCF   string
	HasParams bool
	Params    string
	Response  string
}

type TypeScriptClient struct {
	smd    smd.Schema
	client struct {
		Interfaces []tsInterface
		Namespaces []tsServiceNamespace
	}
	interfaces map[string]interface{}
}

func NewTypeScriptClient(smd smd.Schema) *TypeScriptClient {
	return &TypeScriptClient{
		smd:        smd,
		interfaces: map[string]interface{}{},
	}
}

// Run converts SMD client to TypeScript model.
func (c *TypeScriptClient) Run() ([]byte, error) {
	c.convert()

	var fns = template.FuncMap{
		"len": func(a interface{}) int {
			return reflect.ValueOf(a).Len() - 1
		},
	}

	tmpl, err := template.New("test").Funcs(fns).Parse(
		`/* eslint-disable */{{range .Interfaces}}
export interface {{.Name}} {
{{$len := len .Parameters}}{{range $i,$e := .Parameters}}  {{.Name}}{{if .Optional}}?{{end}}: {{.Type}}{{if ne $i $len}},{{end}}{{if ne .Comment ""}} // {{.Comment}}{{end}}{{if ne $i $len}}
{{end}}{{end}}
}
{{end}}
export const factory = (send) => ({
{{$lenN := len .Namespaces}}{{range $i,$e := .Namespaces}}  {{.Name}}: {
{{$lenS := len .Services}}{{range $i,$e := .Services}}    {{.NameLCF}}({{if .HasParams}}params: {{.Params}}{{end}}): Promise<{{.Response}}> {
      return send('{{.Namespace}}.{{.Name}}'{{if .HasParams}}, params{{end}})
    }{{if ne $i $lenS}},
{{end}}{{end}}
  }{{if ne $i $lenN}},
{{end}}{{end}}
})
`)
	if err != nil {
		return nil, err
	}

	// compile template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c.client); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// convert converts SMD services to TypeScript.
func (c *TypeScriptClient) convert() {
	// iterate over all services
	for serviceName, service := range c.smd.Services {
		serviceNameParts := strings.Split(serviceName, ".")
		if len(serviceNameParts) != 2 {
			continue
		}
		namespace := serviceNameParts[0]
		method := serviceNameParts[1]
		interfaceName := fmt.Sprintf("%s%s%sParams", interfacePrefix, strings.Title(namespace), strings.Title(method))

		// add service params as TypeScript interfaces
		if len(service.Parameters) > 0 {
			tsTypes := make([]tsType, len(service.Parameters))
			for i := range service.Parameters {
				tsTypes[i] = c.convertType(service.Parameters[i], "")
			}
			c.addInterface(tsInterface{
				Name:       interfaceName,
				Parameters: tsTypes,
			})
		}

		// add service "returns" as TypeScript interface
		respType := c.convertType(service.Returns, "")

		// add namespace to TypeScript services
		nIdx := -1
		for i := range c.client.Namespaces {
			if c.client.Namespaces[i].Name == namespace {
				nIdx = i
			}
		}
		if nIdx == -1 {
			c.client.Namespaces = append(c.client.Namespaces, tsServiceNamespace{
				Name:     namespace,
				Services: nil,
			})
			nIdx = len(c.client.Namespaces) - 1
		}

		// add service to TypeScript services
		respService := tsService{
			Namespace: namespace,
			Name:      method,
			NameLCF:   lcfirst(method),
			HasParams: false,
			Params:    "",
			Response:  respType.Type,
		}
		if len(service.Parameters) > 0 {
			respService.HasParams = true
			respService.Params = interfaceName
		}
		if respService.Response == "" {
			respService.Response = voidResponse
		}

		c.client.Namespaces[nIdx].Services = append(c.client.Namespaces[nIdx].Services, respService)
	}

	// sort interfaces
	sort.Slice(c.client.Interfaces, func(i, j int) bool {
		return c.client.Interfaces[i].Name < c.client.Interfaces[j].Name
	})

	// sort namespaces
	sort.Slice(c.client.Namespaces, func(i, j int) bool {
		return c.client.Namespaces[i].Name < c.client.Namespaces[j].Name
	})

	// sort methods
	for idx := range c.client.Namespaces {
		sort.Slice(c.client.Namespaces[idx].Services, func(i, j int) bool {
			return c.client.Namespaces[idx].Services[i].Name < c.client.Namespaces[idx].Services[j].Name
		})
	}
}

// addInterface adds TypeScript interface to client.
func (c *TypeScriptClient) addInterface(ti tsInterface) {
	if len(ti.Parameters) == 0 {
		return
	}

	if _, ok := c.interfaces[ti.Name]; !ok {
		c.client.Interfaces = append(c.client.Interfaces, ti)
		c.interfaces[ti.Name] = struct{}{}
	}
}

// convertScalar converts TypeScript scalars.
func (c *TypeScriptClient) convertScalar(t string) string {
	switch t {
	case "integer", "int":
		return "number"
	default:
		return t
	}
}

// convertType converts smd.JSONSchema to tsType.
func (c *TypeScriptClient) convertType(in smd.JSONSchema, comment string) tsType {
	result := tsType{
		Name:     in.Name,
		Comment:  comment,
		Type:     c.convertScalar(in.Type),
		Optional: in.Optional,
	}

	// detect array sub type
	if in.Type == "array" {
		var subType string
		if scalar, ok := in.Items["type"]; ok {
			subType = c.convertScalar(scalar)
		}
		if ref, ok := in.Items["$ref"]; ok {
			subType = interfacePrefix + strings.TrimPrefix(ref, definitionsPrefix)
		}

		result.Type = fmt.Sprintf("Array<%s>", subType)
	}

	// add object as complex type
	if in.Type == "object" && in.Description != "" {
		c.addComplexInterface(in)
		result.Type = interfacePrefix + in.Description
	}

	// add definitions as complex types
	for name, d := range in.Definitions {
		c.addComplexInterface(smd.JSONSchema{
			Name:        name,
			Description: name,
			Type:        d.Type,
			Properties:  d.Properties,
		})
	}

	return result
}

// addComplexInterface converts complex type stored in smd.JSONSchema to tsInterface and adds it to client.
func (c *TypeScriptClient) addComplexInterface(in smd.JSONSchema) {
	var tsTypes []tsType

	for name, p := range in.Properties {
		tsTypes = append(tsTypes, c.convertType(smd.JSONSchema{
			Name:        name,
			Description: strings.TrimPrefix(p.Ref, definitionsPrefix),
			Type:        p.Type,
			Items:       p.Items,
		}, p.Description))
	}

	c.addInterface(tsInterface{
		Name:       interfacePrefix + in.Description,
		Parameters: tsTypes,
	})
}

func lcfirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToLower(v))
		return u + str[len(u):]
	}
	return ""
}
