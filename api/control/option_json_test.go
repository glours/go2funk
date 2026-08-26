package control_test

import (
	"bytes"
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"testing"

	"github.com/glours/go2funk/api/control"
)

func TestOptionMarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option any
		want   string
	}{
		{"Some string", control.Some("countess"), `"countess"`},
		{"Some int", control.Some(10), `10`},
		{"Some zero int", control.Some(0), `0`},
		{"Some false", control.Some(false), `false`},
		{"Some slice", control.Some([]int{1, 2}), `[1,2]`},
		{"None string", control.None[string](), `null`},
		{"None int", control.None[int](), `null`},
		{"zero value", control.Option[int]{}, `null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.option)
			if err != nil {
				t.Fatalf("Marshal returned %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("Marshal = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestOptionUnmarshalJSON(t *testing.T) {
	var value control.Option[string]

	if err := json.Unmarshal([]byte(`"countess"`), &value); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if got := value.OrElse("none"); got != "countess" {
		t.Errorf("value = %q, want %q", got, "countess")
	}

	// null gives None, and overwrites whatever was there.
	if err := json.Unmarshal([]byte(`null`), &value); err != nil {
		t.Fatalf("Unmarshal(null) returned %v", err)
	}
	if !value.IsEmpty() {
		t.Error("null must unmarshal to None")
	}
}

func TestOptionUnmarshalJSONError(t *testing.T) {
	value := control.Some("kept")

	err := json.Unmarshal([]byte(`42`), &value)
	if err == nil {
		t.Fatal("unmarshalling a number into an Option[string] must fail")
	}
	if got := value.OrElse("lost"); got != "kept" {
		t.Errorf("the Option was modified on error: %q", got)
	}
}

func TestOptionRoundTrip(t *testing.T) {
	type profile struct {
		Name     string                 `json:"name"`
		Nickname control.Option[string] `json:"nickname"`
		Age      control.Option[int]    `json:"age"`
	}

	for _, original := range []profile{
		{Name: "ada", Nickname: control.Some("countess"), Age: control.Some(36)},
		{Name: "ada", Nickname: control.None[string](), Age: control.None[int]()},
		{Name: "ada", Nickname: control.Some(""), Age: control.Some(0)},
	} {
		encoded, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal returned %v", err)
		}
		var back profile
		if err := json.Unmarshal(encoded, &back); err != nil {
			t.Fatalf("Unmarshal returned %v", err)
		}
		if back != original {
			t.Errorf("round trip: got %+v, want %+v (json %s)", back, original, encoded)
		}
	}
}

func TestOptionMissingKeyIsNone(t *testing.T) {
	type profile struct {
		Name     string                 `json:"name"`
		Nickname control.Option[string] `json:"nickname"`
	}

	var back profile
	if err := json.Unmarshal([]byte(`{"name":"ada"}`), &back); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if !back.Nickname.IsEmpty() {
		t.Error("a missing key must leave the Option empty")
	}
}

// IsZero is what makes `omitzero` able to drop the field entirely.
// `omitempty` cannot: it does not understand struct types.
func TestOptionIsZeroAndOmitzero(t *testing.T) {
	if !control.None[int]().IsZero() {
		t.Error("None must report IsZero")
	}
	if control.Some(0).IsZero() {
		t.Error("Some(0) is not zero: a value is present")
	}

	type tags struct {
		Zero  control.Option[string] `json:"zero,omitzero"`
		Empty control.Option[string] `json:"empty,omitempty"`
		Plain control.Option[string] `json:"plain"`
	}

	got, err := json.Marshal(tags{})
	if err != nil {
		t.Fatalf("Marshal returned %v", err)
	}
	want := `{"empty":null,"plain":null}`
	if string(got) != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

func TestOptionValuesThatCollapseToNull(t *testing.T) {
	for _, tc := range []struct {
		name    string
		encoded func() ([]byte, error)
	}{
		{"nil pointer", func() ([]byte, error) { return json.Marshal(control.Some[*int](nil)) }},
		{"nil slice", func() ([]byte, error) { return json.Marshal(control.Some[[]int](nil)) }},
		{"nil map", func() ([]byte, error) { return json.Marshal(control.Some[map[string]int](nil)) }},
		{"nil interface", func() ([]byte, error) { return json.Marshal(control.Some[any](nil)) }},
		{"nested None", func() ([]byte, error) { return json.Marshal(control.Some(control.None[int]())) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := tc.encoded()
			if err != nil {
				t.Fatalf("Marshal returned %v", err)
			}
			if string(encoded) != "null" {
				t.Errorf("Marshal = %s, want null", encoded)
			}
		})
	}

	var back control.Option[*int]
	if err := json.Unmarshal([]byte("null"), &back); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if !back.IsEmpty() {
		t.Error("a null payload comes back as None, whatever produced it")
	}
}

// pointerMarshaler declares MarshalJSON on its pointer, which only works if the
// value inside the Option is addressed before being handed to the encoder.
type pointerMarshaler struct{ N int }

func (p *pointerMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"ptr"`), nil }

// An Option[T] must encode T exactly as a plain field of type T would.
func TestOptionUsesPointerReceiverMarshaler(t *testing.T) {
	payload := struct {
		Plain pointerMarshaler                 `json:"plain"`
		Boxed control.Option[pointerMarshaler] `json:"boxed"`
	}{pointerMarshaler{7}, control.Some(pointerMarshaler{7})}

	encoded, err := json.Marshal(&payload)
	if err != nil {
		t.Fatalf("Marshal returned %v", err)
	}
	if want := `{"plain":"ptr","boxed":"ptr"}`; string(encoded) != want {
		t.Errorf("Marshal = %s, want %s", encoded, want)
	}
}

func TestOptionWorksWithJSONV2(t *testing.T) {
	encoded, err := jsonv2.Marshal(control.Some("countess"))
	if err != nil {
		t.Fatalf("v2 Marshal returned %v", err)
	}
	if string(encoded) != `"countess"` {
		t.Errorf("v2 Marshal = %s, want %q", encoded, "countess")
	}

	encoded, err = jsonv2.Marshal(control.None[string]())
	if err != nil {
		t.Fatalf("v2 Marshal(None) returned %v", err)
	}
	if string(encoded) != `null` {
		t.Errorf("v2 Marshal(None) = %s, want null", encoded)
	}

	var back control.Option[string]
	if err := jsonv2.Unmarshal([]byte(`"ada"`), &back); err != nil {
		t.Fatalf("v2 Unmarshal returned %v", err)
	}
	if got := back.OrElse("none"); got != "ada" {
		t.Errorf("v2 Unmarshal = %q, want %q", got, "ada")
	}
}

// Under encoding/json/v2 the value inside an Option must be decoded with the
// caller's options, not with v1 semantics.
func TestOptionHonoursV2Semantics(t *testing.T) {
	type inner struct {
		Value int `json:"value"`
	}

	// v2 matches member names case-sensitively; v1 does not.
	var boxed control.Option[inner]
	var plain inner
	boxedErr := jsonv2.Unmarshal([]byte(`{"VALUE":7}`), &boxed)
	plainErr := jsonv2.Unmarshal([]byte(`{"VALUE":7}`), &plain)

	if (boxedErr == nil) != (plainErr == nil) {
		t.Errorf("case sensitivity differs: boxed err=%v, plain err=%v", boxedErr, plainErr)
	}
	if got := boxed.OrElse(inner{}); got != plain {
		t.Errorf("boxed decoded to %+v, plain to %+v: v2 semantics must match", got, plain)
	}

	// Caller options must reach inside the Option.
	var strictBoxed control.Option[inner]
	var strictPlain inner
	boxedErr = jsonv2.Unmarshal([]byte(`{"value":7,"extra":1}`), &strictBoxed, jsonv2.RejectUnknownMembers(true))
	plainErr = jsonv2.Unmarshal([]byte(`{"value":7,"extra":1}`), &strictPlain, jsonv2.RejectUnknownMembers(true))

	if boxedErr == nil {
		t.Error("RejectUnknownMembers must apply inside an Option")
	}
	if (boxedErr == nil) != (plainErr == nil) {
		t.Errorf("strictness differs: boxed err=%v, plain err=%v", boxedErr, plainErr)
	}

	var empty control.Option[inner]
	if err := jsonv2.Unmarshal([]byte(`null`), &empty); err != nil {
		t.Fatalf("v2 Unmarshal(null) returned %v", err)
	}
	if !empty.IsEmpty() {
		t.Error("null must decode to None through the v2 path too")
	}
}

// omitzero behaves the same with both encoders; omitempty does not. Pinned so
// the README advice cannot drift.
func TestOmitTagsDifferBetweenEncoders(t *testing.T) {
	type tags struct {
		Zero  control.Option[string] `json:"zero,omitzero"`
		Empty control.Option[string] `json:"empty,omitempty"`
	}

	v1, err := json.Marshal(tags{})
	if err != nil {
		t.Fatalf("v1 Marshal returned %v", err)
	}
	v2, err := jsonv2.Marshal(tags{})
	if err != nil {
		t.Fatalf("v2 Marshal returned %v", err)
	}

	if want := `{"empty":null}`; string(v1) != want {
		t.Errorf("v1 = %s, want %s (omitzero drops, omitempty does not)", v1, want)
	}
	if want := `{}`; string(v2) != want {
		t.Errorf("v2 = %s, want %s (v2 omitempty drops a null too)", v2, want)
	}
}

// The outer encoder's settings must reach the value inside the Option.
func TestOptionHonoursEncoderSettings(t *testing.T) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(struct {
		Plain string                 `json:"plain"`
		Boxed control.Option[string] `json:"boxed"`
	}{"<x>", control.Some("<x>")})
	if err != nil {
		t.Fatalf("Encode returned %v", err)
	}

	want := `{"plain":"<x>","boxed":"<x>"}` + "\n"
	if buffer.String() != want {
		t.Errorf("Encode = %s, want %s", buffer.String(), want)
	}
}

// json.Marshaler is an alias for jsonv2.Marshaler: the byte-slice form of the
// contract. Both encoders prefer MarshalerTo, so these methods are only reached
// by code that dispatches on the interface itself — which is how they are
// exercised here.
func TestOptionSatisfiesTheByteSliceInterfaces(t *testing.T) {
	var marshaler json.Marshaler = control.Some("countess")

	encoded, err := marshaler.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned %v", err)
	}
	if string(encoded) != `"countess"` {
		t.Errorf("MarshalJSON = %s, want %q", encoded, "countess")
	}

	marshaler = control.None[string]()
	if encoded, err = marshaler.MarshalJSON(); err != nil || string(encoded) != `null` {
		t.Errorf("None.MarshalJSON = %s, %v; want null", encoded, err)
	}

	// The pointer-receiver marshaler of T must be used here too.
	marshaler = control.Some(pointerMarshaler{7})
	if encoded, err = marshaler.MarshalJSON(); err != nil || string(encoded) != `"ptr"` {
		t.Errorf("MarshalJSON = %s, %v; want %q", encoded, err, "ptr")
	}

	var target control.Option[string]
	var unmarshaler json.Unmarshaler = &target

	if err := unmarshaler.UnmarshalJSON([]byte(`"ada"`)); err != nil {
		t.Fatalf("UnmarshalJSON returned %v", err)
	}
	if got := target.OrElse("none"); got != "ada" {
		t.Errorf("UnmarshalJSON gave %q, want %q", got, "ada")
	}

	if err := unmarshaler.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("UnmarshalJSON(null) returned %v", err)
	}
	if !target.IsEmpty() {
		t.Error("null must decode to None through the byte-slice path too")
	}

	target = control.Some("kept")
	if err := unmarshaler.UnmarshalJSON([]byte(`42`)); err == nil {
		t.Error("a payload that does not fit T must fail")
	} else if got := target.OrElse("lost"); got != "kept" {
		t.Errorf("the Option was modified on error: %q", got)
	}
}

// A truncated or malformed null gets past PeekKind and fails while being read.
// The Option must be left untouched, as it is for any other decoding error.
func TestOptionMalformedNull(t *testing.T) {
	for _, payload := range []string{`nul`, `nulX`, `n`} {
		t.Run(payload, func(t *testing.T) {
			value := control.Some(7)

			err := jsonv2.Unmarshal([]byte(payload), &value)
			if err == nil {
				t.Fatalf("Unmarshal(%q) must fail", payload)
			}
			if got := value.OrElse(-1); got != 7 {
				t.Errorf("the Option was modified on error: %d", got)
			}
		})
	}
}

// Decoding into an Option that already holds a value must merge into it, the
// way it does for a plain field of the same type.
func TestOptionDecodeMergesIntoTheHeldValue(t *testing.T) {
	type inner struct{ A, B int }
	type document struct {
		Plain inner                 `json:"plain"`
		Boxed control.Option[inner] `json:"boxed"`
	}

	doc := document{Plain: inner{1, 2}, Boxed: control.Some(inner{1, 2})}
	if err := json.Unmarshal([]byte(`{"plain":{"A":9},"boxed":{"A":9}}`), &doc); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}

	want := inner{A: 9, B: 2}
	if doc.Plain != want {
		t.Errorf("plain = %+v, want %+v", doc.Plain, want)
	}
	if got := doc.Boxed.OrElse(inner{}); got != want {
		t.Errorf("boxed = %+v, want %+v: an Option must merge like a plain field", got, want)
	}

	// An empty Option starts from the zero value, not from stale contents.
	doc.Boxed = control.None[inner]()
	if err := json.Unmarshal([]byte(`{"boxed":{"A":9}}`), &doc); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if got := doc.Boxed.OrElse(inner{}); got != (inner{A: 9}) {
		t.Errorf("boxed = %+v, want {A:9}", got)
	}
}

// A missing key leaves the Option alone; only an explicit null resets it.
func TestOptionMissingKeyLeavesTheValueAlone(t *testing.T) {
	type document struct {
		Nickname control.Option[string] `json:"nickname"`
	}

	doc := document{Nickname: control.Some("kept")}
	if err := json.Unmarshal([]byte(`{}`), &doc); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if got := doc.Nickname.OrElse("lost"); got != "kept" {
		t.Errorf("a missing key changed the Option: %q", got)
	}

	if err := json.Unmarshal([]byte(`{"nickname":null}`), &doc); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if !doc.Nickname.IsEmpty() {
		t.Error("an explicit null must reset the Option")
	}
}

// A caller invoking UnmarshalJSON directly may hand over a padded literal.
func TestOptionPaddedNull(t *testing.T) {
	for _, payload := range []string{"null", " null ", "\n\tnull\n"} {
		value := control.Some(7)
		unmarshaler := json.Unmarshaler(&value)

		if err := unmarshaler.UnmarshalJSON([]byte(payload)); err != nil {
			t.Fatalf("UnmarshalJSON(%q) returned %v", payload, err)
		}
		if !value.IsEmpty() {
			t.Errorf("UnmarshalJSON(%q) left the Option defined", payload)
		}
	}
}
