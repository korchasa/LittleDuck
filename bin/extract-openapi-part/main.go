package main

import (
    "encoding/json"
    "fmt"
    "github.com/getkin/kin-openapi/openapi3"
    "os"
    "strings"
)

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run main.go <path_to_spec_file> <comma_separated_endpoints>")
        os.Exit(1)
    }

    specFile := os.Args[1]
    operationIDs := strings.Split(os.Args[2], ",")

    spec, err := loadOpenAPISpec(specFile)
    if err != nil {
        fmt.Printf("Error loading OpenAPI spec: %v\n", err)
        os.Exit(1)
    }

    extractEndpointsAndEntities(spec, operationIDs)
}

func loadOpenAPISpec(specFile string) (*openapi3.T, error) {
    data, err := os.ReadFile(specFile)
    if err != nil {
        return nil, err
    }

    loader := openapi3.NewLoader()
    spec, err := loader.LoadFromData(data)
    if err != nil {
        return nil, err
    }

    return spec, nil
}

func extractEndpointsAndEntities(spec *openapi3.T, operationIDs []string) {
    result := openapi3.T{
        OpenAPI: spec.OpenAPI,
        Info:    spec.Info,
        Paths:   make(openapi3.Paths),
        Components: &openapi3.Components{
            Schemas: make(map[string]*openapi3.SchemaRef),
        },
    }

    for _, endpoint := range operationIDs {
        for path, pathItem := range spec.Paths {
            for method, operation := range pathItem.Operations() {
                if operation.OperationID == endpoint {
                    printfToStdErr("Extracting %s=%s from %s\n", operation.OperationID, endpoint, path)
                    if _, ok := result.Paths[path]; !ok {
                        result.Paths[path] = &openapi3.PathItem{}
                    }
                    result.Paths[path].SetOperation(method, operation)

                    // Extract used schemas
                    for _, param := range operation.Parameters {
                        extractSchemasFromSchemaRef(param.Value.Schema, spec.Components.Schemas, result.Components.Schemas)
                    }
                    if operation.RequestBody != nil {
                        for _, requestBody := range operation.RequestBody.Value.Content {
                            extractSchemasFromSchemaRef(requestBody.Schema, spec.Components.Schemas, result.Components.Schemas)
                        }
                    }
                    for _, response := range operation.Responses {
                        for _, responseBody := range response.Value.Content {
                            extractSchemasFromSchemaRef(responseBody.Schema, spec.Components.Schemas, result.Components.Schemas)
                        }
                    }
                }
            }
        }
    }

    output, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        fmt.Printf("Error generating output: %v\n", err)
        return
    }

    fmt.Println(string(output))
}

func extractSchemasFromSchemaRef(ref *openapi3.SchemaRef, sourceSchemas, targetSchemas map[string]*openapi3.SchemaRef) {
    if ref == nil {
        return
    }

    if ref.Ref != "" {
        name := strings.TrimPrefix(ref.Ref, "#/components/schemas/")
        if _, ok := targetSchemas[name]; !ok {
            targetSchemas[name] = sourceSchemas[name]
            extractSchemasFromSchema(ref.Value, sourceSchemas, targetSchemas)
        }
    } else {
        extractSchemasFromSchema(ref.Value, sourceSchemas, targetSchemas)
    }
}

func extractSchemasFromSchema(schema *openapi3.Schema, sourceSchemas, targetSchemas map[string]*openapi3.SchemaRef) {
    if schema == nil {
        return
    }
    for _, prop := range schema.Properties {
        extractSchemasFromSchemaRef(prop, sourceSchemas, targetSchemas)
    }

    if schema.Items != nil {
        extractSchemasFromSchemaRef(schema.Items, sourceSchemas, targetSchemas)
    }

    if schema.AdditionalProperties.Has != nil {
        extractSchemasFromSchemaRef(schema.AdditionalProperties.Schema, sourceSchemas, targetSchemas)
    }

    if schema.AllOf != nil {
        for _, subSchema := range schema.AllOf {
            extractSchemasFromSchemaRef(subSchema, sourceSchemas, targetSchemas)
        }
    }

    if schema.OneOf != nil {
        for _, subSchema := range schema.OneOf {
            extractSchemasFromSchemaRef(subSchema, sourceSchemas, targetSchemas)
        }
    }

    if schema.AnyOf != nil {
        for _, subSchema := range schema.AnyOf {
            extractSchemasFromSchemaRef(subSchema, sourceSchemas, targetSchemas)
        }
    }
}

func printfToStdErr(format string, a ...interface{}) {
    if _, err := fmt.Fprintf(os.Stderr, format, a...); err != nil {
        panic(err)
    }
}
