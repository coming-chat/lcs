package lcs

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	v             interface{}
	b             []byte
	skipMarshal   bool
	skipUnmarshal bool
	errMarshal    error
	errUnmarshal  error
	name          string
}

func runTest(t *testing.T, cases []*testCase) {
	var b []byte
	var err error

	for idx, c := range cases {
		if !c.skipMarshal {
			b, err = Marshal(c.v)
			if c.errMarshal != nil {
				assert.EqualError(t, err, c.errMarshal.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.b, b)
			}
			t.Logf("Case #%d(%s) marshal: Done", idx, c.name)
		}
		if !c.skipUnmarshal {
			v := reflect.New(reflect.TypeOf(c.v))
			err = Unmarshal(c.b, v.Interface())
			if c.errUnmarshal != nil {
				assert.EqualError(t, err, c.errUnmarshal.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.v, v.Elem().Interface())
			}
			t.Logf("Case #%d(%s) unmarshal: Done", idx, c.name)
		}
	}
}

func hexMustDecode(s string) []byte {
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		panic(err)
	}
	return b
}

func TestBool(t *testing.T) {
	vTrue := true
	type Bool bool
	vTrue2 := Bool(true)

	runTest(t, []*testCase{
		{
			v:    bool(true),
			b:    []byte{1},
			name: "bool true",
		},
		{
			v:    bool(false),
			b:    []byte{0},
			name: "bool false",
		},
		{
			v:    &vTrue,
			b:    []byte{1},
			name: "ptr to bool",
		},
		{
			v:    &vTrue2,
			b:    []byte{1},
			name: "ptr to alias of bool",
		},
	})
}

func TestInts(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    int8(-1),
			b:    []byte{0xFF},
			name: "int8 neg",
		},
		{
			v:    uint8(1),
			b:    []byte{1},
			name: "uint8 pos",
		},
		{
			v:    int16(-4660),
			b:    hexMustDecode("CCED"),
			name: "int16 neg",
		},
		{
			v:    uint16(4660),
			b:    hexMustDecode("3412"),
			name: "uint16 pos",
		},
		{
			v:    int32(-305419896),
			b:    hexMustDecode("88A9CBED"),
			name: "int32 neg",
		},
		{
			v:    uint32(305419896),
			b:    hexMustDecode("78563412"),
			name: "uint32 pos",
		},
		{
			v:    int64(-1311768467750121216),
			b:    hexMustDecode("0011325487A9CBED"),
			name: "int64 neg",
		},
		{
			v:    uint64(1311768467750121216),
			b:    hexMustDecode("00EFCDAB78563412"),
			name: "uint64 pos",
		},
	})
}

func TestBasicSlice(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    []byte{0x11, 0x22, 0x33, 0x44, 0x55},
			b:    hexMustDecode("05000000 11 22 33 44 55"),
			name: "byte slice",
		},
		{
			v:    [6]byte{0x11, 0x22, 0x33, 0x44, 0x55},
			b:    hexMustDecode("06000000 11 22 33 44 55 00"),
			name: "byte array",
		},
		{
			v:    []uint16{0x11, 0x22},
			b:    hexMustDecode("02000000 1100 2200"),
			name: "uint16 slice",
		},
		{
			v:    [3]uint16{0x11, 0x22},
			b:    hexMustDecode("03000000 1100 2200 0000"),
			name: "uint16 array",
		},
		{
			v:    "ሰማይ አይታረስ ንጉሥ አይከሰስ።",
			b:    hexMustDecode("36000000E188B0E1889BE18BAD20E18AA0E18BADE189B3E188A8E188B520E18A95E18C89E188A520E18AA0E18BADE18AA8E188B0E188B5E18DA2"),
			name: "utf8 string",
		},
	})
}

func TestBasicStruct(t *testing.T) {
	type MyStruct struct {
		Boolean    bool
		Bytes      []byte
		Label      string
		unexported uint32
	}
	type Wrapper struct {
		Inner *MyStruct
		Name  string
	}

	runTest(t, []*testCase{
		{
			v: MyStruct{
				Boolean: true,
				Bytes:   []byte{0x11, 0x22},
				Label:   "hello",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f"),
			name: "struct with unexported fields",
		},
		{
			v: &MyStruct{
				Boolean: true,
				Bytes:   []byte{0x11, 0x22},
				Label:   "hello",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f"),
			name: "pointer to struct",
		},
		{
			v: Wrapper{
				Inner: &MyStruct{
					Boolean: true,
					Bytes:   []byte{0x11, 0x22},
					Label:   "hello",
				},
				Name: "world",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f 05000000 776f726c64"),
			name: "nested struct",
		},
		{
			v: &Wrapper{
				Inner: &MyStruct{
					Boolean: true,
					Bytes:   []byte{0x11, 0x22},
					Label:   "hello",
				},
				Name: "world",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f 05000000 776f726c64"),
			name: "pointer to nested struct",
		},
	})
}

func TestOptional(t *testing.T) {
	type Wrapper struct {
		Ignored int     `lcs:"-"`
		Name    *string `lcs:"optional"`
	}
	hello := "hello"

	runTest(t, []*testCase{
		{
			v: Wrapper{
				Name: &hello,
			},
			b:    hexMustDecode("01 05000000 68656c6c6f"),
			name: "struct with set optional fields",
		},
		{
			v: Wrapper{
				Name: nil,
			},
			b:    hexMustDecode("00"),
			name: "struct with unset optional fields",
		},
	})
}

func TestMap(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    map[uint8]string{1: "hello", 2: "world"},
			b:    hexMustDecode("02000000 01 05000000 68656c6c6f 02 05000000 776f726c64"),
			name: "map[uint8]string",
		},
		{
			v:    map[string]uint8{"hello": 1, "world": 2},
			b:    hexMustDecode("02000000 05000000 68656c6c6f 01 05000000 776f726c64 02"),
			name: "map[string]uint8",
		},
	})
}

type MyStruct1 struct {
	Boolean bool
}
type MyStruct2 struct {
	Label string
}
type Wrapper struct {
	Name  string
	Inner interface{} `lcs:"enum:Wrapper.Inner"`
}

func (*Wrapper) EnumTypes() []EnumVariant {
	return []EnumVariant{
		{
			Name:     "Wrapper.Inner",
			Value:    5,
			Template: (*MyStruct1)(nil),
		},
		{
			Name:     "Wrapper.Inner",
			Value:    6,
			Template: (*MyStruct2)(nil),
		},
	}
}

func TestEnum(t *testing.T) {
	runTest(t, []*testCase{
		{
			v: &Wrapper{
				Name:  "1",
				Inner: &MyStruct1{true},
			},
			b:             hexMustDecode("01000000 31 05000000 01"),
			name:          "ptr to struct with enum 1",
			skipUnmarshal: true,
		},
		{
			v: &Wrapper{
				Name:  "2",
				Inner: &MyStruct2{"hello"},
			},
			b:             hexMustDecode("01000000 32 06000000 05000000 68656c6c6f"),
			name:          "ptr to struct with enum 2",
			skipUnmarshal: true,
		},
		{
			v: Wrapper{
				Name:  "2",
				Inner: &MyStruct2{"hello"},
			},
			b:             hexMustDecode("01000000 32 06000000 05000000 68656c6c6f"),
			name:          "struct with enum",
			skipUnmarshal: true,
		},
	})
}