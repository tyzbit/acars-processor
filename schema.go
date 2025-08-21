package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"

	//"github.com/creasty/defaults"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// These are called when jsonschema Reflects, so we don't need to call these.
func (ac ACARSConnectionConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(Attention("couldn't get selectedfields for acars annotator config type"))
		return
	}
	f := ac.GetDefaultFields()
	s.Examples = append(s.Examples, f)
	j.Properties.Set("SelectedFields", s)
}

// These are called when jsonschema Reflects, so we don't need to call these.
func (vc VDLM2ConnectionConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(Attention("couldn't get selectedfields for acars annotator config type"))
		return
	}
	f := vc.GetDefaultFields()
	s.Examples = append(s.Examples, f)
	j.Properties.Set("SelectedFields", s)
}

func (t Tar1090Annotator) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(Attention("couldn't get selectedfields for ollama annotator config type"))
		return
	}
	f := t.GetDefaultFields()
	s.Examples = append(s.Examples, f)
	j.Properties.Set("SelectedFields", s)
}

func (a ADSBExchangeAnnotator) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(Attention("couldn't get selectedfields for ollama annotator config type"))
		return
	}
	f := a.GetDefaultFields()
	s.Examples = append(s.Examples, f)
	j.Properties.Set("SelectedFields", s)
}

func (o OllamaAnnotator) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(Attention("couldn't get selectedfields for ollama annotator config type"))
		return
	}
	f := o.GetDefaultFields()
	s.Examples = append(s.Examples, f)
	j.Properties.Set("SelectedFields", s)
}

func GenerateSchema(schemaPath string) (schemaUpdated bool) {
	// Add comments to the schema from the code
	r := new(jsonschema.Reflector)
	err := r.AddGoComments("main", "./", jsonschema.WithFullComment())
	if err != nil {
		log.Fatal(Attention("unable to add comments to schema, %s", err))
	}

	// Generate the schema and save it as a file
	r.RequiredFromJSONSchemaTags = true
	schema := r.Reflect(&Config{})
	// Suppress further for clean output
	log.SetLevel(log.InfoLevel)
	json, _ := schema.MarshalJSON()
	if UpdateFile(fmt.Sprintf("./%s", schemaFilePath), json) {
		schemaUpdated = true
		log.Info(Success("Updated schema at %s", schemaFilePath))
	}
	return schemaUpdated
}

func justKeys(m APMessage) (s []string) {
	for f := range m {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

// SetBoolDefaults walks through a struct and sets *bool fields according to the `default` tag.
func SetBoolDefaults(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("expected a non-nil pointer to struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to struct")
	}

	return setBoolDefaultsRecursive(v)
}

func setBoolDefaultsRecursive(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// If field is embedded struct, recurse
		if field.Kind() == reflect.Struct {
			if err := setBoolDefaultsRecursive(field); err != nil {
				return err
			}
			continue
		}

		// Only process *bool fields
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Bool {
			tagVal := fieldType.Tag.Get("default")
			if tagVal == "" {
				continue
			}

			// Parse the string value of the tag into a bool
			boolVal, err := strconv.ParseBool(tagVal)
			if err != nil {
				return fmt.Errorf("invalid default tag for field %s: %w", fieldType.Name, err)
			}

			// Only set if the pointer is nil
			if field.IsNil() {
				field.Set(reflect.ValueOf(&boolVal))
			}
		}
	}

	return nil
}

type commentMap map[string]map[string]string

func MarshalYAMLWithComments(v interface{}) ([]byte, error) {
	comments, err := extractAllComments(v)
	if err != nil {
		return nil, err
	}

	node, err := encodeValue(reflect.ValueOf(v), comments)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(node)
}

func encodeValue(val reflect.Value, comments commentMap) (*yaml.Node, error) {
	if !val.IsValid() {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"}, nil
	}

	switch val.Kind() {
	case reflect.Struct:
		return encodeStruct(val, comments)
	case reflect.Slice, reflect.Array:
		return encodeSequence(val, comments)
	case reflect.Map:
		return encodeMapping(val, comments)
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprint(val.Interface())}, nil
	}
}

func encodeStruct(val reflect.Value, comments commentMap) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	rt := val.Type()
	structComments := comments[rt.Name()]

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported or ignored YAML fields
		if field.PkgPath != "" || field.Tag.Get("yaml") == "-" {
			continue
		}

		// Flatten anonymous structs, pointers to structs, or interfaces to structs
		if field.Anonymous {
			underlying := fieldVal

			// Skip nil interface/pointer
			if (underlying.Kind() == reflect.Interface || underlying.Kind() == reflect.Ptr) && underlying.IsNil() {
				continue
			}

			// Dereference pointers or interfaces
			for underlying.Kind() == reflect.Ptr || underlying.Kind() == reflect.Interface {
				underlying = underlying.Elem()
			}

			if underlying.Kind() == reflect.Struct {
				embeddedNode, err := encodeStruct(underlying, comments)
				if err != nil {
					return nil, err
				}
				node.Content = append(node.Content, embeddedNode.Content...)
				continue
			}

			// For anonymous interfaces that are not structs, skip them
			if underlying.Kind() == reflect.Interface {
				continue
			}
		}

		// Normal field key
		key := field.Tag.Get("yaml")
		if key == "" {
			key = field.Name
		}

		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}

		if structComments != nil {
			if comment, ok := structComments[field.Name]; ok {
				keyNode.HeadComment = normalizeComment(comment)
			}
		}

		valNode, err := encodeValue(fieldVal, comments)
		if err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valNode)
	}

	return node, nil
}

func encodeSequence(val reflect.Value, comments commentMap) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.SequenceNode}
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		elemNode, err := encodeValue(elem, comments)
		if err != nil {
			return nil, err
		}

		if comment := getElementComment(elem); comment != "" {
			elemNode.HeadComment = comment
		}

		node.Content = append(node.Content, elemNode)
	}
	return node, nil
}

func encodeMapping(val reflect.Value, comments commentMap) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, key := range val.MapKeys() {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprint(key.Interface())}

		valNode, err := encodeValue(val.MapIndex(key), comments)
		if err != nil {
			return nil, err
		}

		if comment := getElementComment(val.MapIndex(key)); comment != "" {
			valNode.HeadComment = comment
		}

		node.Content = append(node.Content, keyNode, valNode)
	}
	return node, nil
}

// normalizeComment converts Go doc comments into YAML comments.
func normalizeComment(c string) string {
	lines := strings.Split(c, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			continue
		case line == "//":
			out = append(out, "#")
		case strings.HasPrefix(line, "//"):
			out = append(out, strings.TrimSpace(line[2:]))
		default:
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// getElementComment is a placeholder for element-level comments.
func getElementComment(val reflect.Value) string {
	return ""
}

// extractAllComments parses Go source files to collect struct field doc comments.
func extractAllComments(v interface{}) (commentMap, error) {
	comments := make(commentMap)

	_, callerFile, _, ok := runtime.Caller(2)
	if !ok {
		return nil, fmt.Errorf("unable to locate caller source file")
	}
	dir := filepath.Dir(callerFile)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					fieldComments := make(map[string]string)
					for _, field := range structType.Fields.List {
						if len(field.Names) == 0 || field.Doc == nil {
							continue
						}
						fieldComments[field.Names[0].Name] = field.Doc.Text()
					}

					if len(fieldComments) > 0 {
						comments[typeSpec.Name.Name] = fieldComments
					}
				}
			}
		}
	}

	return comments, nil
}
