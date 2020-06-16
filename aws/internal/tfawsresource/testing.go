package tfawsresource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	sentinelIndex = "*"
)

// TestCheckTypeSetElemNestedAttrs is a resource.TestCheckFunc that accepts a resource
// name, an attribute path, which should use the sentinel value '*' for indexing
// into a TypeSet. The function verifies that an element matches the whole value
// map.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a nested object with their own attrs/values.
//
// Please note, if the provided value map is not granular enough, there exists
// the possibility you match an element you were not intending to, in the TypeSet.
// Provide a full mapping of attributes to be sure the unique element exists.
func TestCheckTypeSetElemNestedAttrs(res, attr string, values map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[res]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", res, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", res, ms.Path)
		}

		matches := make(map[string]int)
		attrParts := strings.Split(attr, ".")
		if attrParts[len(attrParts)-1] != sentinelIndex {
			return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
		}
		for stateKey, stateValue := range is.Attributes {
			stateKeyParts := strings.Split(stateKey, ".")
			// a Set/List item with nested attrs would have a flatmap address of
			// at least length 3
			// foo.0.name = "bar"
			if len(stateKeyParts) < 3 {
				continue
			}
			var pathMatch bool
			for i := range attrParts {
				if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
					break
				}
				if i == len(attrParts)-1 {
					pathMatch = true
				}
			}
			if !pathMatch {
				continue
			}
			elementId := stateKeyParts[len(attrParts)-1]
			nestedAttr := strings.Join(stateKeyParts[len(attrParts):], ".")
			if v, keyExists := values[nestedAttr]; keyExists && v == stateValue {
				matches[elementId] = matches[elementId] + 1
				if matches[elementId] == len(values) {
					return nil
				}
			}
		}

		return fmt.Errorf("%q no TypeSet element %q, with nested attrs %#v in state: %#v", res, attr, values, is.Attributes)
	}
}

// TestCheckTypeSetElemAttr is a resource.TestCheckFunc that accepts a resource
// name, an attribute path, which should use the sentinel value '*' for indexing
// into a TypeSet. The function verifies that an element matches the provided
// value.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a simple value
func TestCheckTypeSetElemAttr(res, attr, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[res]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", res, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", res, ms.Path)
		}

		attrParts := strings.Split(attr, ".")
		if attrParts[len(attrParts)-1] != sentinelIndex {
			return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
		}
		for stateKey, stateValue := range is.Attributes {
			if stateValue == value {
				stateKeyParts := strings.Split(stateKey, ".")
				if len(stateKeyParts) == len(attrParts) {
					for i := range attrParts {
						if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
							break
						}
						if i == len(attrParts)-1 {
							return nil
						}
					}
				}
			}
		}

		return fmt.Errorf("%q no TypeSet element %q, with value %q in state: %#v", res, attr, value, is.Attributes)
	}
}
