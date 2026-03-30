package main

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/protobuf/compiler/protogen"
)

type Config struct {
	Title       string
	Version     string
	Description string
	Output      string
	Naming      string
}

func Generate(gen *protogen.Plugin, cfg *Config) error {
	doc := &v3.Document{
		Version: "3.0.3",
		Info: &base.Info{
			Title:       cfg.Title,
			Version:     cfg.Version,
			Description: cfg.Description,
		},
		Paths: &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		},
	}
	registry := NewSchemaRegistry(cfg.Naming)
	ann := DiscoverAnnotations(gen.Files)

	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}
		for _, svc := range file.Services {
			svcName := string(svc.Desc.Name())
			svcMeta := ExtractServiceMetadata(svc, ann)
			svcDesc := extractComment(svc.Comments.Leading)
			servicePath := svcMeta["service_path"]

			for _, method := range svc.Methods {
				route, ok := ExtractHTTPRoute(method, ann)
				if !ok {
					continue
				}

				fullPath := route.Path
				if servicePath != "" {
					fullPath = strings.TrimRight(servicePath, "/") + "/" + strings.TrimLeft(fullPath, "/")
				}
				openAPIPath := ConvertPath(fullPath)
				classified := ClassifyFields(method.Input, ann)
				metadata := ExtractMethodMetadata(method, ann)

				tags := resolveTags(svcName, metadata)
				for _, t := range tags {
					desc := ""
					if t == svcName {
						desc = svcDesc
					}
					addTag(doc, t, desc)
				}

				codes := orderedmap.New[string, *v3.Response]()
				op := &v3.Operation{
					Tags:        tags,
					OperationId: fmt.Sprintf("%s_%s", svcName, method.Desc.Name()),
					Responses:   &v3.Responses{Codes: codes},
				}

				comment := extractComment(method.Comments.Leading)
				if name := metadata["name"]; name != "" {
					op.Summary = name
					if comment != "" {
						op.Description = comment
					}
				} else if comment != "" {
					summary, desc := splitComment(comment)
					op.Summary = summary
					if desc != "" {
						op.Description = desc
					}
				}

				if baseURL := resolveBaseURL(metadata, svcMeta); baseURL != "" {
					op.Servers = []*v3.Server{{URL: baseURL}}
				}

				appendParameters(op, classified)
				setRequestBody(op, classified, registry)

				respRef := registry.Register(method.Output)
				respContent := orderedmap.New[string, *v3.MediaType]()
				respContent.Set("application/json", &v3.MediaType{
					Schema: base.CreateSchemaProxyRef(respRef),
				})
				codes.Set("200", &v3.Response{
					Description: "Successful response",
					Content:     respContent,
				})

				addOperation(doc.Paths, openAPIPath, route.Method, op)
			}
		}
	}

	registry.EmitTo(doc)

	g := gen.NewGeneratedFile(cfg.Output, "")
	return EmitYAML(doc, g)
}

func resolveTags(svcName string, metadata map[string]string) []string {
	if tagStr := metadata["tag"]; tagStr != "" {
		var tags []string
		for _, t := range strings.Split(tagStr, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
		if len(tags) > 0 {
			return tags
		}
	}
	return []string{svcName}
}

func resolveBaseURL(methodMeta, svcMeta map[string]string) string {
	if u := methodMeta["baseurl"]; u != "" {
		return u
	}
	return svcMeta["base_domain"]
}

func appendParameters(op *v3.Operation, classified *ClassifiedFields) {
	for _, cf := range classified.Path {
		op.Parameters = append(op.Parameters, &v3.Parameter{
			Name:     cf.ParamName,
			In:       "path",
			Required: ptr(true),
			Schema:   base.CreateSchemaProxy(kindToSchema(cf.Field.Desc.Kind())),
		})
	}
	for _, cf := range classified.Query {
		op.Parameters = append(op.Parameters, &v3.Parameter{
			Name:   cf.ParamName,
			In:     "query",
			Schema: base.CreateSchemaProxy(kindToSchema(cf.Field.Desc.Kind())),
		})
	}
	for _, cf := range classified.Header {
		op.Parameters = append(op.Parameters, &v3.Parameter{
			Name:   cf.ParamName,
			In:     "header",
			Schema: base.CreateSchemaProxy(kindToSchema(cf.Field.Desc.Kind())),
		})
	}
	for _, cf := range classified.Cookie {
		op.Parameters = append(op.Parameters, &v3.Parameter{
			Name:   cf.ParamName,
			In:     "cookie",
			Schema: base.CreateSchemaProxy(kindToSchema(cf.Field.Desc.Kind())),
		})
	}
}

func setRequestBody(op *v3.Operation, classified *ClassifiedFields, registry *SchemaRegistry) {
	if len(classified.RawBody) > 0 {
		content := orderedmap.New[string, *v3.MediaType]()
		content.Set("application/octet-stream", &v3.MediaType{
			Schema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"string"}, Format: "binary",
			}),
		})
		op.RequestBody = &v3.RequestBody{Required: ptr(true), Content: content}
		return
	}

	if len(classified.Body) > 0 {
		content := orderedmap.New[string, *v3.MediaType]()
		content.Set("application/json", &v3.MediaType{
			Schema: buildBodySchema(classified.Body, registry),
		})
		op.RequestBody = &v3.RequestBody{Required: ptr(true), Content: content}
		return
	}

	if len(classified.Form) > 0 || len(classified.FileName) > 0 {
		content := orderedmap.New[string, *v3.MediaType]()
		content.Set("multipart/form-data", &v3.MediaType{
			Schema: buildFormSchema(classified.Form, classified.FileName, registry),
		})
		op.RequestBody = &v3.RequestBody{Content: content}
	}
}

func addTag(doc *v3.Document, name, description string) {
	for _, t := range doc.Tags {
		if t.Name == name {
			if description != "" && t.Description == "" {
				t.Description = description
			}
			return
		}
	}
	doc.Tags = append(doc.Tags, &base.Tag{Name: name, Description: description})
}

func addOperation(paths *v3.Paths, path, method string, op *v3.Operation) {
	item, _ := paths.PathItems.Get(path)
	if item == nil {
		item = &v3.PathItem{}
		paths.PathItems.Set(path, item)
	}
	switch strings.ToUpper(method) {
	case "GET":
		item.Get = op
	case "POST":
		item.Post = op
	case "PUT":
		item.Put = op
	case "DELETE":
		item.Delete = op
	case "PATCH":
		item.Patch = op
	case "HEAD":
		item.Head = op
	case "OPTIONS":
		item.Options = op
	case "TRACE":
		item.Trace = op
	}
}

func buildBodySchema(fields []*ClassifiedField, registry *SchemaRegistry) *base.SchemaProxy {
	props := orderedmap.New[string, *base.SchemaProxy]()
	for _, cf := range fields {
		props.Set(cf.ParamName, registry.fieldToSchemaProxy(cf.Field))
	}
	return base.CreateSchemaProxy(&base.Schema{
		Type:       []string{"object"},
		Properties: props,
	})
}

func buildFormSchema(formFields, fileFields []*ClassifiedField, registry *SchemaRegistry) *base.SchemaProxy {
	props := orderedmap.New[string, *base.SchemaProxy]()
	for _, cf := range formFields {
		props.Set(cf.ParamName, registry.fieldToSchemaProxy(cf.Field))
	}
	for _, cf := range fileFields {
		props.Set(cf.ParamName, base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"}, Format: "binary",
		}))
	}
	return base.CreateSchemaProxy(&base.Schema{
		Type:       []string{"object"},
		Properties: props,
	})
}

func extractComment(c protogen.Comments) string {
	return strings.TrimSpace(string(c))
}

func splitComment(text string) (summary, description string) {
	i := strings.IndexByte(text, '\n')
	if i < 0 {
		return text, ""
	}
	return strings.TrimSpace(text[:i]), strings.TrimSpace(text[i+1:])
}

func ptr[T any](v T) *T { return &v }
