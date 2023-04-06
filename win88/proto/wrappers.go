package proto

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*
 * Helper routines for simplifying the creation of optional fields of basic type.
 */

// Bool is a helper routine that return it self
func Bool(v bool) bool { return v }

// Int32 is a helper routine that return it self
func Int32(v int32) int32 { return v }

// Int is a helper routine that return it self
func Int(v int) int32 { return int32(v) }

// Int64 is a helper routine that return it self
func Int64(v int64) int64 { return v }

// Float32 is a helper routine that return it self
func Float32(v float32) float32 { return v }

// Float64 is a helper routine that return it self
func Float64(v float64) float64 { return v }

// Uint32 is a helper routine that return it self
func Uint32(v uint32) uint32 { return v }

// Uint64 is a helper routine that return it self
func Uint64(v uint64) uint64 { return v }

// String is a helper routine that return it self
func String(v string) string { return v }

// SetDefaults sets unset protocol buffer fields to their default values.
// do nothing
func SetDefaults(pb interface{}) {}

// Marshal returns the wire-format encoding of m.
func Marshal(m proto.Message) ([]byte, error) { return proto.Marshal(m) }

func Unmarshal(b []byte, m proto.Message) error { return proto.Unmarshal(b, m) }

func Has(m proto.Message, fieldName string) bool {
	if m == nil {
		return false
	}
	if fieldName == "" {
		return false
	}

	pr := m.ProtoReflect()
	if pr == nil {
		return false
	}
	return pr.Has(pr.Descriptor().Fields().ByName(protoreflect.Name(fieldName)))
}

type Message = proto.Message

func MessageName(m proto.Message) protoreflect.FullName {
	if m == nil {
		return ""
	}
	return m.ProtoReflect().Descriptor().FullName()
}
