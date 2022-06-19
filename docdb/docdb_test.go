package docdb

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			sortOpt := cmpopts.SortSlices(func(x, y string) bool {
				return x < y
			})
			got := getPathValues(tt.args.obj, tt.args.prefix)
			if diff := cmp.Diff(tt.want, got, sortOpt); diff != "" {
				t.Errorf("getPathValues() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
