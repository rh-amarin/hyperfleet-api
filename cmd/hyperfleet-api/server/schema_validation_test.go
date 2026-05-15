package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/onsi/gomega"
)

const testSchema = `
openapi: 3.0.0
info:
  title: Test Partner Schema
  version: 1.0.0
paths: {}
components:
  schemas:
    ClusterSpec:
      type: object
      required: [region]
      properties:
        region:
          type: string
    NodePoolSpec:
      type: object
      required: [machine_type]
      properties:
        machine_type:
          type: string
`

func TestApplySchemaValidation_EmptyPath_Disabled(t *testing.T) {
	RegisterTestingT(t)

	router := mux.NewRouter()
	// Empty path means schema validation is disabled — no error.
	err := applySchemaValidation(router, "")
	Expect(err).To(BeNil())
}

func TestApplySchemaValidation_ValidSchema(t *testing.T) {
	RegisterTestingT(t)

	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "partner-schema.yaml")
	Expect(os.WriteFile(schemaPath, []byte(testSchema), 0600)).To(Succeed())

	router := mux.NewRouter()
	err := applySchemaValidation(router, schemaPath)
	Expect(err).To(BeNil())
}

func TestApplySchemaValidation_MissingFile_ReturnsError(t *testing.T) {
	RegisterTestingT(t)

	router := mux.NewRouter()
	// Non-empty path that does not exist must return an error so the caller can fail fast.
	err := applySchemaValidation(router, "/nonexistent/partner-schema.yaml")
	Expect(err).ToNot(BeNil())
}

func TestApplySchemaValidation_InvalidSchema_ReturnsError(t *testing.T) {
	RegisterTestingT(t)

	badSchema := `
openapi: 3.0.0
info:
  title: Bad Schema
  version: 1.0.0
paths: {}
components:
  schemas:
    SomeOtherType:
      type: object
`
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "bad-schema.yaml")
	Expect(os.WriteFile(schemaPath, []byte(badSchema), 0600)).To(Succeed())

	router := mux.NewRouter()
	// Schema that is missing ClusterSpec/NodePoolSpec must return an error.
	err := applySchemaValidation(router, schemaPath)
	Expect(err).ToNot(BeNil())
}
