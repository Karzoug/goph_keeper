package user

import (
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		email    string
		authHash []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive",
			args: args{
				email:    "test@example.com",
				authHash: []byte("12345"),
			},
			wantErr: false,
		},
		{
			name: "negative: empty hash",
			args: args{
				email:    "test@example.com",
				authHash: []byte(""),
			},
			wantErr: true,
		},
		{
			name: "negative: empty email",
			args: args{
				email:    "",
				authHash: []byte("123456"),
			},
			wantErr: true,
		},
		{
			name: "negative: wrong email format",
			args: args{
				email:    "testexamplecom",
				authHash: []byte("123456"),
			},
			wantErr: true,
		},
		{
			name: "negative: not valid domain",
			args: args{
				email:    "info@pupkinsupercompany.com",
				authHash: []byte("123456"),
			},
			wantErr: true,
		},
		{
			name: "negative: wrong email format #2",
			args: args{
				email:    "infopupkin@",
				authHash: []byte("123456"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.email, tt.args.authHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
