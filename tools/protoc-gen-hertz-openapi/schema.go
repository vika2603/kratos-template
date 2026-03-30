package main

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type SchemaRegistry struct {
	messagesByFullName map[string]*protogen.Message
	displayNameByFull  map[string]string
	usedDisplayNames   map[string]bool
	order              []string
	naming             string
}

func NewSchemaRegistry(naming string) *SchemaRegistry {
	return &SchemaRegistry{
		messagesByFullName: make(map[string]*protogen.Message),
		displayNameByFull:  make(map[string]string),
		usedDisplayNames:   make(map[string]bool),
		naming:             naming,
	}
}

func (r *SchemaRegistry) Register(msg *protogen.Message) string {
	fullName := string(msg.Desc.FullName())
	if name, exists := r.displayNameByFull[fullName]; exists {
		return "#/components/schemas/" + name
	}

	r.messagesByFullName[fullName] = msg
	shortName := string(msg.Desc.Name())
	var displayName string
	if !r.usedDisplayNames[shortName] {
		displayName = shortName
	} else {
		displayName = strings.ReplaceAll(fullName, ".", "_")
	}
	r.usedDisplayNames[displayName] = true
	r.displayNameByFull[fullName] = displayName
	r.order = append(r.order, fullName)
	return "#/components/schemas/" + displayName
}

func (r *SchemaRegistry) EmitTo(doc *v3.Document) {
	if len(r.order) == 0 {
		return
	}
	if doc.Components == nil {
		doc.Components = &v3.Components{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}
	for i := 0; i < len(r.order); i++ {
		fullName := r.order[i]
		msg := r.messagesByFullName[fullName]
		displayName := r.displayNameByFull[fullName]
		doc.Components.Schemas.Set(displayName, base.CreateSchemaProxy(r.messageToSchema(msg)))
	}
}

func (r *SchemaRegistry) messageToSchema(msg *protogen.Message) *base.Schema {
	props := orderedmap.New[string, *base.SchemaProxy]()
	for _, field := range msg.Fields {
		props.Set(r.fieldName(field), r.fieldToSchemaProxy(field))
	}
	schema := &base.Schema{
		Type:       []string{"object"},
		Properties: props,
	}
	if desc := extractComment(msg.Comments.Leading); desc != "" {
		schema.Description = desc
	}
	return schema
}

func (r *SchemaRegistry) fieldToSchemaProxy(field *protogen.Field) *base.SchemaProxy {
	desc := field.Desc

	if desc.IsMap() {
		valueField := field.Message.Fields[1]
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				A: r.fieldToSchemaProxy(valueField),
			},
		})
	}

	var proxy *base.SchemaProxy
	switch desc.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		fullName := desc.Message().FullName()
		if wk, ok := wellKnownSchemas[fullName]; ok {
			cp := *wk
			proxy = base.CreateSchemaProxy(&cp)
		} else {
			proxy = base.CreateSchemaProxyRef(r.Register(field.Message))
		}
	case protoreflect.EnumKind:
		values := desc.Enum().Values()
		enumVals := make([]*yaml.Node, values.Len())
		for i := 0; i < values.Len(); i++ {
			enumVals[i] = &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: string(values.Get(i).Name()),
			}
		}
		proxy = base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
			Enum: enumVals,
		})
	default:
		proxy = base.CreateSchemaProxy(kindToSchema(desc.Kind()))
	}

	if desc.IsList() {
		return base.CreateSchemaProxy(&base.Schema{
			Type:  []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: proxy},
		})
	}

	return proxy
}

func kindToSchema(kind protoreflect.Kind) *base.Schema {
	switch kind {
	case protoreflect.BoolKind:
		return &base.Schema{Type: []string{"boolean"}}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return &base.Schema{Type: []string{"integer"}, Format: "int32"}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return &base.Schema{Type: []string{"integer"}, Format: "uint32"}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return &base.Schema{Type: []string{"string"}, Format: "int64"}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return &base.Schema{Type: []string{"string"}, Format: "uint64"}
	case protoreflect.FloatKind:
		return &base.Schema{Type: []string{"number"}, Format: "float"}
	case protoreflect.DoubleKind:
		return &base.Schema{Type: []string{"number"}, Format: "double"}
	case protoreflect.StringKind:
		return &base.Schema{Type: []string{"string"}}
	case protoreflect.BytesKind:
		return &base.Schema{Type: []string{"string"}, Format: "byte"}
	default:
		return &base.Schema{Type: []string{"string"}}
	}
}

func (r *SchemaRegistry) fieldName(field *protogen.Field) string {
	if r.naming == "proto" {
		return string(field.Desc.Name())
	}
	return string(field.Desc.JSONName())
}
