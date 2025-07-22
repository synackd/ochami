// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package config

import (
	"fmt"
	"testing"
)

func TestErrMissingURI_Error(t *testing.T) {
	type fields struct {
		Service ServiceName
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valid error",
			fields: fields{
				Service: ServiceBSS,
			},
			want: fmt.Sprintf("base URI for %s not found (neither cluster.uri nor %s.uri specified)", ServiceBSS, ServiceBSS),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emu := ErrMissingURI{
				Service: tt.fields.Service,
			}
			if got := emu.Error(); got != tt.want {
				t.Errorf("ErrMissingURI.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrInvalidURI_Error(t *testing.T) {
	type fields struct {
		Err error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valid contained error",
			fields: fields{
				Err: fmt.Errorf("unknown URI format (must be \"proto://host[:port][/path]\")"),
			},
			want: fmt.Sprintf("invalid URI: %v", fmt.Sprintf("unknown URI format (must be \"proto://host[:port][/path]\")")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eiu := ErrInvalidURI{
				Err: tt.fields.Err,
			}
			if got := eiu.Error(); got != tt.want {
				t.Errorf("ErrInvalidURI.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrInvalidServiceURI_Error(t *testing.T) {
	type fields struct {
		Err     error
		Service ServiceName
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valid contained error and service",
			fields: fields{
				Err:     fmt.Errorf("unknown URI format (must be \"proto://host[:port][/path]\")"),
				Service: ServiceBSS,
			},
			want: fmt.Sprintf("invalid service URI for %s: %v", ServiceBSS, fmt.Sprintf("unknown URI format (must be \"proto://host[:port][/path]\")")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eisu := ErrInvalidServiceURI{
				Err:     tt.fields.Err,
				Service: tt.fields.Service,
			}
			if got := eisu.Error(); got != tt.want {
				t.Errorf("ErrInvalidServiceURI.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrUnknownService_Error(t *testing.T) {
	type fields struct {
		Service string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "unknown service",
			fields: fields{
				Service: "unk_svc",
			},
			want: "unknown service: unk_svc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eus := ErrUnknownService{
				Service: tt.fields.Service,
			}
			if got := eus.Error(); got != tt.want {
				t.Errorf("ErrUnknownService.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
