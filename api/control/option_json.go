package control

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

// Option implements both forms of the marshaling contract that encoding/json/v2
// defines: the byte-slice one, which encoding/json.Marshaler is an alias for,
// and the streaming one. Both encoders prefer the streaming form, which is what
// lets the value inside the Option be encoded with the caller's options; the
// byte-slice form is what code dispatching on json.Marshaler reaches.
var (
	_ json.Marshaler         = Option[int]{}
	_ json.Unmarshaler       = (*Option[int])(nil)
	_ jsonv2.MarshalerTo     = Option[int]{}
	_ jsonv2.UnmarshalerFrom = (*Option[int])(nil)
)

// MarshalJSON encodes a defined Option as its value and an empty one as null.
//
// The zero value of Option is None, so a field left untouched encodes as null.
// To drop the field from the object entirely, tag it with omitzero, which
// consults IsZero:
//
//	type Profile struct {
//		Nickname Option[string] `json:"nickname,omitzero"`
//	}
//
// omitzero works with both encoders. omitempty does not, and the two differ:
// encoding/json emits null, while encoding/json/v2 drops the field. Prefer
// omitzero so the shape is the same either way.
//
// Any value that encodes as null comes back as None, so the outer "a value is
// present" bit is lost for a nil pointer, a nil slice or map, a nil interface,
// and a nested None. Some((*T)(nil)) and None cannot be told apart once encoded.
// This form cannot see the caller's encoder settings, so it always escapes HTML
// the way encoding/json does by default. It also always addresses the value,
// which a plain field of type T only does when it is addressable.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.defined {
		return []byte("null"), nil
	}
	// The value is addressed so that a MarshalJSON declared on *T is used, the
	// way it would be for a plain struct field of type T.
	return json.Marshal(&o.value)
}

// UnmarshalJSON decodes null into None and anything else into Some. A missing
// key leaves the Option untouched, which for a fresh value means None.
//
// The Option is left unchanged when the payload does not fit T.
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	// The encoders hand over the bare literal, but a caller invoking this
	// method directly may pass a padded json.RawMessage.
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*o = Option[T]{}
		return nil
	}
	value := o.currentOrZero()
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*o = Some(value)
	return nil
}

// currentOrZero returns the value already held, so that decoding merges into it
// the way it does for a plain field of type T, or the zero value when empty.
func (o Option[T]) currentOrZero() T {
	if o.defined {
		return o.value
	}
	var zero T
	return zero
}

// MarshalJSONTo is the encoding/json/v2 form of MarshalJSON. It hands the value
// to the caller's encoder, so the options in force apply to the contents of the
// Option as they would to a bare T.
func (o Option[T]) MarshalJSONTo(encoder *jsontext.Encoder) error {
	if !o.defined {
		return encoder.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(encoder, &o.value)
}

// UnmarshalJSONFrom is the encoding/json/v2 form of UnmarshalJSON. It reads the
// value through the caller's decoder, so options such as RejectUnknownMembers
// apply inside the Option too.
func (o *Option[T]) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	if decoder.PeekKind() == 'n' {
		if _, err := decoder.ReadToken(); err != nil {
			return err
		}
		*o = Option[T]{}
		return nil
	}
	value := o.currentOrZero()
	if err := jsonv2.UnmarshalDecode(decoder, &value); err != nil {
		return err
	}
	*o = Some(value)
	return nil
}

// IsZero reports whether the Option is empty. It is what lets the omitzero
// struct tag drop the field, and it is honoured by both encoders.
func (o Option[T]) IsZero() bool { return !o.defined }
