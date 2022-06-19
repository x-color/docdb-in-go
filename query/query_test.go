package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseQuery(t *testing.T) {
	type args struct {
		q string
	}
	tests := []struct {
		name    string
		args    args
		want    Queries
		wantErr bool
	}{
		{
			name: "Query: 'a.b:1'",
			args: args{
				q: "a.b:1",
			},
			want: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeEq,
				},
			},
			wantErr: false,
		},
		{
			name: "Query: 'a:<10'",
			args: args{
				q: "a:<10",
			},
			want: Queries{
				{
					Keys:  []string{"a"},
					Value: "10",
					Op:    OpeLt,
				},
			},
			wantErr: false,
		},
		{
			name: "Query: 'a:>10'",
			args: args{
				q: "a:>10",
			},
			want: Queries{
				{
					Keys:  []string{"a"},
					Value: "10",
					Op:    OpeGt,
				},
			},
			wantErr: false,
		},
		{
			name: "Query: 'a.b:>10 a.c:hello'",
			args: args{
				q: "a.b:>10 a.c:hello",
			},
			want: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "10",
					Op:    OpeGt,
				},
				{
					Keys:  []string{"a", "c"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			wantErr: false,
		},
		{
			name: "Query: '\" a \":\" hello \"'",
			args: args{
				q: `" a ":" hello "`,
			},
			want: Queries{
				{
					Keys:  []string{" a "},
					Value: " hello ",
					Op:    OpeEq,
				},
			},
			wantErr: false,
		},
		{
			name: "Query: ' a:hello '",
			args: args{
				q: ` a:hello `,
			},
			want: Queries{
				{
					Keys:  []string{"a"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			wantErr: false,
		},
		{
			name: "No Query: ''",
			args: args{
				q: "",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Invalid Query: 'a'",
			args: args{
				q: "a",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Query: 'a:'",
			args: args{
				q: "a:",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Query: ':1'",
			args: args{
				q: ":1",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseQuery() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestQueries_Match(t *testing.T) {
	type args struct {
		doc map[string]any
	}
	tests := []struct {
		name string
		qs   Queries
		args args
		want bool
	}{
		{
			name: "Simple Query 'a.b:hello'",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": "hello",
					},
				},
			},
			want: true,
		},
		{
			name: "Simple Query 'a.b:hello' (Not Matching)",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": "1",
					},
				},
			},
			want: false,
		},
		{
			name: "Simple Query 'a.b:hello' (Key Does Not Exists)",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			args: args{
				doc: map[string]any{
					"a": 1,
				},
			},
			want: false,
		},
		{
			name: "Simple Query 'a.b:>1'",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeGt,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": 2,
					},
				},
			},
			want: true,
		},
		{
			name: "Simple Query 'a.b:>1' (Not Matching)",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeGt,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": 1,
					},
				},
			},
			want: false,
		},
		{
			name: "Simple Query 'a.b:<1'",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeLt,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": 0.0,
					},
				},
			},
			want: true,
		},
		{
			name: "Simple Query 'a.b:<1' (Not Matching)",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeLt,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": 3,
					},
				},
			},
			want: false,
		},
		{
			name: "Multiple Query 'a.b:1 b.c:hello'",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeEq,
				},
				{
					Keys:  []string{"b", "c"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": 1,
					},
					"b": map[string]any{
						"c": "hello",
					},
				},
			},
			want: true,
		},
		{
			name: "Multiple Query 'a.b:1 b.c:hello' (Not Matching)",
			qs: Queries{
				{
					Keys:  []string{"a", "b"},
					Value: "1",
					Op:    OpeEq,
				},
				{
					Keys:  []string{"b", "c"},
					Value: "hello",
					Op:    OpeEq,
				},
			},
			args: args{
				doc: map[string]any{
					"a": map[string]any{
						"b": 1,
					},
					"b": map[string]any{
						"c": "World",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qs.Match(tt.args.doc); got != tt.want {
				t.Errorf("Queries.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
