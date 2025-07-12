package log

import (
	"testing"
)

func TestInit(t *testing.T) {
	type args struct {
		ll string
		lf string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "supported level and format",
			args: args{
				ll: "warning",
				lf: "basic",
			},
			wantErr: false,
		},
		{
			name: "unsupported level and supported format",
			args: args{
				ll: "unsupported",
				lf: "basic",
			},
			wantErr: true,
		},
		{
			name: "supported level and unsupported format",
			args: args{
				ll: "warning",
				lf: "unsupported",
			},
			wantErr: true,
		},
		{
			name: "unsupported level and unsupported format",
			args: args{
				ll: "unsupported",
				lf: "unsupported",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Init(tt.args.ll, tt.args.lf); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
