package encode

import "testing"

func TestParseSliceSpecs(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []int
	}{
		{
			name:  "space separated list",
			input: []string{"0 37 40 46"},
			want:  []int{0, 37, 40, 46},
		},
		{
			name:  "mixed punctuation",
			input: []string{"0,37，40;46；67"},
			want:  []int{0, 37, 40, 46, 67},
		},
		{
			name:  "repeated flags",
			input: []string{"0", "37", "40"},
			want:  []int{0, 37, 40},
		},
		{
			name:  "range",
			input: []string{"10-12"},
			want:  []int{10, 11, 12},
		},
		{
			name:  "receiver text",
			input: []string{"切片缺失: 0 37 40 46 ..."},
			want:  []int{0, 37, 40, 46},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSliceSpecs(tt.input)
			if err != nil {
				t.Fatalf("ParseSliceSpecs returned error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d indexes, want %d: %#v", len(got), len(tt.want), got)
			}
			for _, index := range tt.want {
				if !got[index] {
					t.Fatalf("missing index %d in %#v", index, got)
				}
			}
		})
	}
}

func TestParseSliceSpecsInvalidRange(t *testing.T) {
	if _, err := ParseSliceSpecs([]string{"12-10"}); err == nil {
		t.Fatal("expected invalid range error")
	}
}
