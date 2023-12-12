package jsonschema

import (
	"embed"
	"fmt"
)

//go:embed testdata/openapi
var openapiBundle embed.FS

func ExampleCompile() {
	c := NewCompiler()

	oas31SchemaReader, err := openapiBundle.Open("testdata/openapi/openapi_3.1.schema.json")
	if err != nil {
		panic(err)
	}

	if err := c.AddResource("https://spec.openapis.org/oas/3.1/schema/2022-10-07", oas31SchemaReader); err != nil {
		panic(err)
	}

	oas31Schema, err := c.Compile("https://spec.openapis.org/oas/3.1/schema/2022-10-07")
	if err != nil {
		panic(err)
	}

	c.RegisterExtension("openapi_3.1", oas31Schema, &oas3SchemaCompiler{})

	validSpecReader, err := openapiBundle.Open("testdata/openapi/valid-spec.json")
	if err != nil {
		panic(err)
	}

	if err := c.AddResource("valid-spec.json", validSpecReader); err != nil {
		panic(err)
	}

	compiled, err := c.Compile("valid-spec.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("Compiled: valid-spec.json with extensions:", len(compiled.Extensions))

	validSubSchema, err := c.Compile("valid-spec.json#/components/schemas/Foo")
	if err != nil {
		fmt.Println("ERROR: failed to compile subschema \"valid-spec.json#/components/schemas/Foo\":", err.Error())
	} else {
		fmt.Println("Compiled: valid-spec.json#/components/schemas/Foo with extensions:", len(validSubSchema.Extensions))
	}

	invalidSpecReader, err := openapiBundle.Open("testdata/openapi/invalid-spec.json")
	if err != nil {
		panic(err)
	}

	if err := c.AddResource("invalid-spec.json", invalidSpecReader); err != nil {
		panic(err)
	}

	_, err = c.Compile("invalid-spec.json")
	if err != nil {
		fmt.Println("failed to compile invalid-spec.json")
	}

	// invalid spec but valid subSchema
	_, err = c.Compile("invalid-spec.json#/components/schemas/Foo")
	if err != nil {
		fmt.Println("ERROR: failed to compile subschema \"invalid-spec.json#/components/schemas/Foo\":", err.Error())
	} else {
		fmt.Println("Compiled: invalid-spec.json#/components/schemas/Foo with extensions:", len(validSubSchema.Extensions))
	}

	// invalid spec and invalid subSchema
	_, err = c.Compile("invalid-spec.json#/components/schemas/FooInvalid")
	if err != nil {
		fmt.Println("failed to compile subschema \"invalid-spec.json#/components/schemas/FooInvalid\"")
	} else {
		fmt.Println("ERROR: unexpectedly compiled invalid-spec.json#/components/schemas/FooInvalid with extensions:", len(validSubSchema.Extensions))
	}

	// Output:
	// Compiled: valid-spec.json with extensions: 1
	// Compiled: valid-spec.json#/components/schemas/Foo with extensions: 1
	// failed to compile invalid-spec.json
	// Compiled: invalid-spec.json#/components/schemas/Foo with extensions: 1
	// failed to compile subschema "invalid-spec.json#/components/schemas/FooInvalid"
}

type oas3SchemaCompiler struct{}

func (o *oas3SchemaCompiler) Compile(ctx CompilerContext, m map[string]interface{}) (ExtSchema, error) {
	if _, ok := m["openapi"]; ok {
		return &oas3Schema{}, nil
	}

	return nil, nil
}

type oas3Schema struct {
}

func (o *oas3Schema) Validate(ctx ValidationContext, v interface{}) error {
	// I don't think we need custom validation, the schema added with the extension is enough.
	return nil
}
