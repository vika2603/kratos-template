package main

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var wellKnownSchemas = map[protoreflect.FullName]*base.Schema{
	"google.protobuf.Timestamp": {Type: []string{"string"}, Format: "date-time"},
	"google.protobuf.Duration":  {Type: []string{"string"}},
	"google.protobuf.Empty":     {Type: []string{"object"}},
	"google.protobuf.Struct":    {Type: []string{"object"}},
	"google.protobuf.Value":     {},
	"google.protobuf.ListValue": {
		Type:  []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{})},
	},
	"google.protobuf.FieldMask": {Type: []string{"string"}},
	"google.protobuf.Any":       {Type: []string{"object"}},
	"google.protobuf.DoubleValue": {
		Type: []string{"number"}, Format: "double", Nullable: ptr(true),
	},
	"google.protobuf.FloatValue": {
		Type: []string{"number"}, Format: "float", Nullable: ptr(true),
	},
	"google.protobuf.Int64Value": {
		Type: []string{"string"}, Format: "int64", Nullable: ptr(true),
	},
	"google.protobuf.UInt64Value": {
		Type: []string{"string"}, Format: "uint64", Nullable: ptr(true),
	},
	"google.protobuf.Int32Value": {
		Type: []string{"integer"}, Format: "int32", Nullable: ptr(true),
	},
	"google.protobuf.UInt32Value": {
		Type: []string{"integer"}, Format: "uint32", Nullable: ptr(true),
	},
	"google.protobuf.BoolValue": {
		Type: []string{"boolean"}, Nullable: ptr(true),
	},
	"google.protobuf.StringValue": {
		Type: []string{"string"}, Nullable: ptr(true),
	},
	"google.protobuf.BytesValue": {
		Type: []string{"string"}, Format: "byte", Nullable: ptr(true),
	},
}
