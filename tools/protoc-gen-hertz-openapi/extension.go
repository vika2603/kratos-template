package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var httpMethodNames = map[protoreflect.Name]string{
	"get": "GET", "post": "POST", "put": "PUT",
	"delete": "DELETE", "patch": "PATCH",
	"options": "OPTIONS", "head": "HEAD", "any": "GET",
}

var methodMetadataNames = map[protoreflect.Name]bool{
	"tag": true, "name": true, "baseurl": true,
}

var fieldLocationNames = map[protoreflect.Name]bool{
	"path": true, "query": true, "header": true,
	"cookie": true, "body": true, "form": true,
	"raw_body": true, "file_name": true,
}

var serviceMetadataNames = map[protoreflect.Name]bool{
	"base_domain": true, "service_path": true,
}

type Annotations struct {
	methodRoutes   map[protowire.Number]string
	methodMetadata map[protowire.Number]string
	fieldTypes     map[protowire.Number]string
	serviceMetadata map[protowire.Number]string
}

func DiscoverAnnotations(files []*protogen.File) *Annotations {
	ann := &Annotations{
		methodRoutes:    make(map[protowire.Number]string),
		methodMetadata:  make(map[protowire.Number]string),
		fieldTypes:      make(map[protowire.Number]string),
		serviceMetadata: make(map[protowire.Number]string),
	}
	for _, f := range files {
		exts := f.Desc.Extensions()
		for i := 0; i < exts.Len(); i++ {
			ext := exts.Get(i)
			num := protowire.Number(ext.Number())
			name := ext.Name()
			switch ext.ContainingMessage().FullName() {
			case "google.protobuf.MethodOptions":
				if method, ok := httpMethodNames[name]; ok {
					ann.methodRoutes[num] = method
				}
				if methodMetadataNames[name] {
					ann.methodMetadata[num] = string(name)
				}
			case "google.protobuf.FieldOptions":
				if fieldLocationNames[name] {
					ann.fieldTypes[num] = string(name)
				}
			case "google.protobuf.ServiceOptions":
				if serviceMetadataNames[name] {
					ann.serviceMetadata[num] = string(name)
				}
			}
		}
	}
	return ann
}

type HTTPRoute struct {
	Method string
	Path   string
}

func ExtractHTTPRoute(method *protogen.Method, ann *Annotations) (*HTTPRoute, bool) {
	opts := method.Desc.Options()
	if opts == nil {
		return nil, false
	}
	b, err := proto.Marshal(opts)
	if err != nil || len(b) == 0 {
		return nil, false
	}
	for num, httpMethod := range ann.methodRoutes {
		if path, ok := scanForString(b, num); ok {
			return &HTTPRoute{Method: httpMethod, Path: path}, true
		}
	}
	return nil, false
}

func ExtractMethodMetadata(method *protogen.Method, ann *Annotations) map[string]string {
	opts := method.Desc.Options()
	if opts == nil {
		return nil
	}
	b, err := proto.Marshal(opts)
	if err != nil || len(b) == 0 {
		return nil
	}
	var result map[string]string
	for num, name := range ann.methodMetadata {
		if val, ok := scanForString(b, num); ok {
			if result == nil {
				result = make(map[string]string)
			}
			result[name] = val
		}
	}
	return result
}

func ExtractServiceMetadata(svc *protogen.Service, ann *Annotations) map[string]string {
	opts := svc.Desc.Options()
	if opts == nil {
		return nil
	}
	b, err := proto.Marshal(opts)
	if err != nil || len(b) == 0 {
		return nil
	}
	var result map[string]string
	for num, name := range ann.serviceMetadata {
		if val, ok := scanForString(b, num); ok {
			if result == nil {
				result = make(map[string]string)
			}
			result[name] = val
		}
	}
	return result
}

type ClassifiedField struct {
	Field     *protogen.Field
	ParamName string
}

type ClassifiedFields struct {
	Path     []*ClassifiedField
	Query    []*ClassifiedField
	Header   []*ClassifiedField
	Cookie   []*ClassifiedField
	Body     []*ClassifiedField
	Form     []*ClassifiedField
	RawBody  []*ClassifiedField
	FileName []*ClassifiedField
}

func ClassifyFields(msg *protogen.Message, ann *Annotations) *ClassifiedFields {
	cf := &ClassifiedFields{}
	for _, field := range msg.Fields {
		classifyOneField(field, cf, ann)
	}
	return cf
}

func classifyOneField(field *protogen.Field, cf *ClassifiedFields, ann *Annotations) {
	opts := field.Desc.Options()
	if opts == nil {
		return
	}
	b, err := proto.Marshal(opts)
	if err != nil || len(b) == 0 {
		return
	}
	for num, locType := range ann.fieldTypes {
		if name, ok := scanForString(b, num); ok {
			var dest *[]*ClassifiedField
			switch locType {
			case "path":
				dest = &cf.Path
			case "query":
				dest = &cf.Query
			case "header":
				dest = &cf.Header
			case "cookie":
				dest = &cf.Cookie
			case "body":
				dest = &cf.Body
			case "form":
				dest = &cf.Form
			case "raw_body":
				dest = &cf.RawBody
			case "file_name":
				dest = &cf.FileName
			default:
				return
			}
			*dest = append(*dest, &ClassifiedField{
				Field:     field,
				ParamName: name,
			})
			return
		}
	}
}

func scanForString(b []byte, fieldNum protowire.Number) (string, bool) {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return "", false
		}
		b = b[n:]
		switch typ {
		case protowire.VarintType:
			_, n = protowire.ConsumeVarint(b)
		case protowire.Fixed32Type:
			_, n = protowire.ConsumeFixed32(b)
		case protowire.Fixed64Type:
			_, n = protowire.ConsumeFixed64(b)
		case protowire.BytesType:
			v, vn := protowire.ConsumeBytes(b)
			if vn < 0 {
				return "", false
			}
			if num == fieldNum {
				return string(v), true
			}
			n = vn
		case protowire.StartGroupType:
			_, n = protowire.ConsumeGroup(num, b)
		default:
			return "", false
		}
		if n < 0 {
			return "", false
		}
		b = b[n:]
	}
	return "", false
}
