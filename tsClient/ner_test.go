package tsClient

import (
	"reflect"
	"testing"
)

func TestSpan(t *testing.T) {
	v11 := response{
		Token: "A",
		Pred:  "B-PER",
	}
	v12 := response{
		Token: "A",
		Pred:  "I-PER",
	}
	v13 := response{
		Token: "B",
		Pred:  "B-PER",
	}
	valid1 := []response{v11, v12, v13}

	v21 := response{
		Token: "A",
		Pred:  "B-ORG",
	}
	v22 := response{
		Token: "A",
		Pred:  "I-ORG",
	}
	v23 := response{
		Token: "A",
		Pred:  "I-ORG",
	}
	v24 := response{
		Token: "B",
		Pred:  "O",
	}
	valid2 := []response{v21, v22, v23, v24}

	type args struct {
	}

	tests := []struct {
		name      string
		args      args
		responses []response
		target    string
		want      string
		wantErr   bool
	}{
		{
			name:      "empty responses",
			responses: []response{},
			target:    "PER",
			wantErr:   true,
		},
		{
			name: "no leading B responses",
			responses: []response{
				{
					Token: "",
					Pred:  "I-PER",
				},
			},
			target:  "PER",
			wantErr: true,
		},
		{
			name: "leading O responses",
			responses: []response{
				{
					Token: "",
					Pred:  "O",
				},
			},
			target:  "PER",
			wantErr: true,
		},
		{
			name: "leading B does not match target",
			responses: []response{
				{
					Token: "",
					Pred:  "B-ORG",
				},
			},
			target:  "PER",
			wantErr: true,
		},
		{
			name:      "valid 1",
			responses: valid1,
			target:    "PER",
			want:      "A A",
			wantErr:   false,
		},
		{
			name:      "valid 2",
			responses: valid2,
			target:    "ORG",
			want:      "A A A",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := span(tt.responses, tt.target)
				if (err != nil) != tt.wantErr {
					t.Errorf("span() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("span() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestEntities(t *testing.T) {
	r11 := response{
		Token: "This",
		Pred:  "O",
	}
	r12 := response{
		Token: "is",
		Pred:  "O",
	}
	r13 := response{
		Token: "a",
		Pred:  "O",
	}
	r14 := response{
		Token: "Test",
		Pred:  "O",
	}
	r15 := response{
		Token: ".",
		Pred:  "O",
	}
	r1 := []response{r11, r12, r13, r14, r15}

	r21 := response{
		Token: "O",
		Pred:  "",
	}
	r22 := response{
		Token: "p1",
		Pred:  "B-PER",
	}
	r23 := response{
		Token: "p2",
		Pred:  "I-PER",
	}
	r24 := response{
		Token: "o",
		Pred:  "B-ORG",
	}
	r25 := response{
		Token: "O",
		Pred:  "",
	}
	r26 := response{
		Token: "l",
		Pred:  "B-LOC",
	}
	r27 := response{
		Token: "p3",
		Pred:  "B-PER",
	}
	r2 := []response{r21, r22, r23, r24, r25, r26, r27}

	r31 := response{
		Token: "a",
		Pred:  "B-PER",
	}
	r32 := response{
		Token: "a",
		Pred:  "I-PER",
	}
	r33 := response{
		Token: "a",
		Pred:  "B-PER",
	}
	r34 := response{
		Token: "a",
		Pred:  "I-PER",
	}
	r35 := response{
		Token: "a",
		Pred:  "B-PER",
	}
	r36 := response{
		Token: "a",
		Pred:  "B-PER",
	}
	r37 := response{
		Token: "b",
		Pred:  "B-PER",
	}

	r3 := []response{r31, r32, r33, r34, r35, r36, r37}

	tests := []struct {
		name      string
		responses []response
		wantPers  []string
		wantLocs  []string
		wantOrgs  []string
	}{
		{
			name:      "empty responses",
			responses: []response{},
			wantOrgs:  []string{},
			wantPers:  []string{},
			wantLocs:  []string{},
		},
		{
			name:      "empty results",
			responses: r1,
			wantOrgs:  []string{},
			wantPers:  []string{},
			wantLocs:  []string{},
		},
		{
			name:      "pass",
			responses: r2,
			wantOrgs:  []string{"o"},
			wantPers:  []string{"p1 p2", "p3"},
			wantLocs:  []string{"l"},
		},
		{
			name:      "duplicates",
			responses: r3,
			wantOrgs:  []string{},
			wantPers:  []string{"a a", "a", "b"},
			wantLocs:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				gotPers, gotLocs, gotOrgs := entities(tt.responses)
				if !reflect.DeepEqual(gotPers, tt.wantPers) {
					t.Errorf("entities() gotPers = %v, want %v", gotPers, tt.wantPers)
				}
				if !reflect.DeepEqual(gotLocs, tt.wantLocs) {
					t.Errorf("entities() gotLocs = %v, want %v", gotLocs, tt.wantLocs)
				}
				if !reflect.DeepEqual(gotOrgs, tt.wantOrgs) {
					t.Errorf("entities() gotOrgs = %v, want %v", gotOrgs, tt.wantOrgs)
				}
			},
		)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name  string
		given []string
		want  []string
	}{
		{
			name:  "empty",
			given: []string{},
			want:  []string{},
		},
		{
			name:  "one elem",
			given: []string{"a"},
			want:  []string{"a"},
		},
		{
			name:  "no dups",
			given: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "pass",
			given: []string{"a", "b", "c", "a", "a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := removeDuplicates(tt.given); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("removeDuplicates() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
