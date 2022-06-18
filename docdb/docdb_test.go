package docdb

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_getPathValues(t *testing.T) {
	type args struct {
		obj    map[string]any
		prefix string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Generate path and value set",
			args: args{
				obj: map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": 1,
							"d": map[string]any{
								"e": 1,
							},
						},
					},
				},
			},
			want: []string{
				"a.b.c=1",
				"a.b.d.e=1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans := cmp.Transformer("Sort", func(in []int) []int {
				out := append([]int(nil), in...)
				sort.Ints(out)
				return out
			})
			got := getPathValues(tt.args.obj, tt.args.prefix)
			if diff := cmp.Diff(tt.want, got, trans); diff != "" {
				t.Errorf("getPathValues() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
