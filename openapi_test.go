package jsonschema

import (
	"embed"
	"fmt"
)

//go:embed testdata/openapi
var openapiBundle embed.FS

func ExampleCompile() {
	c := NewCompiler()

	// we want to validate OpenAPI 3.1 specifications against its meta schema. As I understand it
	// this is the correct way to add support for custom vocabularies?
	registerExtension(c)

	// load the spec, this works as expected
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

	// now we receive a request with some data which is associated to one of the operations defined
	// in the spec. We want to validate the request data against that operation, which we'll say
	// here is defined by the Foo schema, so we want to compile this into a proper Schema so that we
	// can call `validSubSchema.Validate()`.
	//
	// What's happening is it's compiling the referenced schema against the root of the openapi meta
	// schema. I'd have expected it to compile against the schema located at:
	//
	//   https://spec.openapis.org/oas/3.1/schema/2022-10-07#/properties/components/properties/schemas/additionalProperties
	validSubSchema, err := c.Compile("valid-spec.json#/components/schemas/Foo")
	if err != nil {
		fmt.Println("ERROR: failed to compile subschema \"valid-spec.json#/components/schemas/Foo\":", err.Error())
	} else {
		fmt.Println("Compiled: valid-spec.json#/components/schemas/Foo with extensions:", len(validSubSchema.Extensions))
	}

	// ---
	// now we load an invalid spec, this kind of works as expected, however it would be nice if we
	// could still compile the subschema and only care about those errors (if there are any).
	invalidSpecReader, err := openapiBundle.Open("testdata/openapi/invalid-spec.json")
	if err != nil {
		panic(err)
	}

	if err := c.AddResource("invalid-spec.json", invalidSpecReader); err != nil {
		panic(err)
	}

	_, err = c.Compile("invalid-spec.json")
	if err != nil {
		// expected
		fmt.Println("failed to compile invalid-spec.json")
	}

	// invalid spec but valid subSchema
	_, err = c.Compile("invalid-spec.json#/components/schemas/Foo")
	if err != nil {
		// this fails for the same reason above, it's compiling against the root of the meta schema
		// and not the schema pointed at /properties/components/properties/schemas/additionalProperties
		fmt.Println("ERROR: failed to compile subschema \"invalid-spec.json#/components/schemas/Foo\":", err.Error())
	} else {
		fmt.Println("Compiled: invalid-spec.json#/components/schemas/Foo with extensions:", len(validSubSchema.Extensions))
	}

	// invalid spec and invalid subSchema
	_, err = c.Compile("invalid-spec.json#/components/schemas/FooInvalid")
	if err != nil {
		// This fails as expected :)
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

func registerExtension(c *Compiler) {
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
