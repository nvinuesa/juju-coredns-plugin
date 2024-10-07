package juju

import (
	"testing"
)

func TestJujuName(t *testing.T) {
	tests := []struct {
		name        string
		jujuName    string
		want        JujuFQDN
		expectedErr bool
	}{
		{
			name:        "Missing labels",
			jujuName:    "a.b.c.myapp.mymodel.mycontroller.juju.local.",
			want:        JujuFQDN{},
			expectedErr: true,
		},
		{
			name:        "Too many labels",
			jujuName:    "controller0.juju.local.",
			want:        JujuFQDN{},
			expectedErr: true,
		},
		{
			name:        "Label with more than 63 chars",
			jujuName:    "0123456789012345678901234567890123456789012345678901234567890123.postgres.mymodel.mycontroller.juju.local.",
			expectedErr: true,
		},
		{
			name:        "Query with more than 255 chars",
			jujuName:    "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123.postgres.mymodel.mycontroller.juju.local.",
			expectedErr: true,
		},
		{
			name:     "Correct FQDN",
			jujuName: "0.myapp.mymodel.mycontroller.juju.local.",
			want: JujuFQDN{
				Controller:  "mycontroller",
				Model:       "mymodel",
				Application: "myapp",
				Unit:        "0",
			},
		},
		{
			name:     "Correct FQDN, leader app",
			jujuName: "leader.myapp.mymodel.mycontroller.juju.local.",
			want: JujuFQDN{
				Controller:  "mycontroller",
				Model:       "mymodel",
				Application: "myapp",
				Unit:        "leader",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJujuFQDN(tt.jujuName)
			if err != nil && !tt.expectedErr {
				t.Errorf("ParseJujuName() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			if tt.expectedErr && err == nil {
				t.Error("ParseJujuName() expected error, got nil")
				return
			} else if got != tt.want {
				t.Errorf("ParseJujuName() = %v, want %v", got, tt.want)
			}
		})
	}
}
