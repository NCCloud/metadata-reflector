package common

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	errorPossibleFunction := func() error {
		return nil
	}

	Must(errorPossibleFunction())
}

func TestMust_PanicsOnError(t *testing.T) {
	errorFunction := func() error {
		return errors.New("error occured")
	}
	defer func() { _ = recover() }()

	Must(errorFunction())

	t.Errorf("did not panic")
}

func TestMustReturn(t *testing.T) {
	errorPossibleFunction := func(a bool) (bool, error) {
		return a, nil
	}

	result := MustReturn(errorPossibleFunction(true))

	assert.True(t, result)
}

func TestMustReturn_PanicsOnError(t *testing.T) {
	errorPossibleFunction := func(a bool) (bool, error) {
		return a, errors.New("error occured")
	}

	defer func() {
		_ = recover()
	}()

	MustReturn(errorPossibleFunction(true))

	t.Errorf("did not panic")
}

func TestPointerTo(t *testing.T) {
	value := 1

	result := new(value)

	assert.Equal(t, &value, result)
}

func TestMapHasPrefix(t *testing.T) {
	type args struct {
		prefix string
		data   map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Prefix found in map keys",
			args: args{
				prefix: "test",
				data:   map[string]string{"testKey": "value1", "otherKey": "value2"},
			},
			want: true,
		},
		{
			name: "Prefix not found in map keys",
			args: args{
				prefix: "prefix",
				data:   map[string]string{"testKey": "value1", "otherKey": "value2"},
			},
			want: false,
		},
		{
			name: "Empty prefix matches all keys",
			args: args{
				prefix: "",
				data:   map[string]string{"testKey": "value1", "otherKey": "value2"},
			},
			want: true,
		},
		{
			name: "Empty map returns false",
			args: args{
				prefix: "test",
				data:   map[string]string{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapHasPrefix(tt.args.prefix, tt.args.data)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapKeysAsString(t *testing.T) {
	type args struct {
		data map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string // we will compare it to the split string
	}{
		{
			name: "Empty map",
			args: args{
				data: map[string]string{},
			},
			want: []string{""},
		},
		{
			name: "Comma-separated string of map keys",
			args: args{
				data: map[string]string{"testKey": "value1", "otherKey": "value2"},
			},
			want: []string{"testKey", "otherKey"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapKeysAsString(tt.args.data)

			assert.ElementsMatch(t, tt.want, strings.Split(got, ","))
		})
	}
}

func TestExcessiveElements(t *testing.T) {
	type args struct {
		expected []string
		actual   []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "No excessive elements",
			args: args{
				expected: []string{"element1", "element2"},
				actual:   []string{"element1", "element2"},
			},
			want: nil,
		},
		{
			name: "Excessive elements are present",
			args: args{
				expected: []string{"element1", "element2"},
				actual:   []string{"element1", "element2", "element3"},
			},
			want: []string{"element3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExcessiveElements(tt.args.expected, tt.args.actual)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapKeysAsSlice(t *testing.T) {
	type args struct {
		data map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Slice from map",
			args: args{
				data: map[string]string{"element1": "value1"},
			},
			want: []string{"element1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapKeysAsSlice(tt.args.data)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapContainsPartialKey(t *testing.T) {
	type args struct {
		partialKey string
		data       map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Map contains partial key",
			args: args{
				partialKey: "element",
				data:       map[string]string{"element1": "value1", "element2": "value2"},
			},
			want: true,
		},
		{
			name: "Map does not contain partial key",
			args: args{
				partialKey: "item",
				data:       map[string]string{"element1": "value1", "element2": "value2"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapContainsPartialKey(tt.args.partialKey, tt.args.data)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindPartialKeys(t *testing.T) {
	type args struct {
		partialKey string
		data       map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Map contains partial key",
			args: args{
				partialKey: "element",
				data:       map[string]string{"element1": "value1", "item1": "value1"},
			},
			want: map[string]string{"element1": "value1"},
		},
		{
			name: "Map does not contain partial key",
			args: args{
				partialKey: "item",
				data:       map[string]string{"element1": "value1", "element2": "value2"},
			},
			want: make(map[string]string),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindPartialKeys(tt.args.partialKey, tt.args.data)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExactMatchRegex(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Pattern has prefix",
			args: args{
				pattern: "pattern$",
			},
			want: "^pattern$",
		},
		{
			name: "Pattern has suffix",
			args: args{
				pattern: "^pattern",
			},
			want: "^pattern$",
		},
		{
			name: "Pattern has prefix and suffix",
			args: args{
				pattern: "^pattern$",
			},
			want: "^pattern$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExactMatchRegex(tt.args.pattern)

			assert.Equal(t, tt.want, got)
		})
	}
}
