package main

import (
	"testing"
)

func TestFilterMacOSArgs(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "RemovesPsn",
			input: []string{"-psn_0_12345", "version"},
			want:  []string{"version"},
		},
		{
			name:  "PreservesNormalArgs",
			input: []string{"version", "--help"},
			want:  []string{"version", "--help"},
		},
		{
			name:  "EmptyArgs",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "OnlyPsn",
			input: []string{"-psn_0_99999"},
			want:  []string{},
		},
		{
			name:  "MultiplePsn",
			input: []string{"-psn_0_1", "-psn_0_2", "version"},
			want:  []string{"version"},
		},
		{
			name:  "PsnLikeButNot",
			input: []string{"--psn_flag"},
			want:  []string{"--psn_flag"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterMacOSArgs(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("filterMacOSArgs(%v) = %v (len %d), want %v (len %d)",
					tc.input, got, len(got), tc.want, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("filterMacOSArgs(%v)[%d] = %q, want %q",
						tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}
