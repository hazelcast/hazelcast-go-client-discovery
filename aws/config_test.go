package aws_test

import (
	"fmt"
	"testing"

	"github.com/hazelcast/hazelcast-go-client/cluster"

	"github.com/hazelcast/hazelcast-go-client-discovery/aws"
)

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		cfg     aws.Config
		filters []aws.Filter
		valid   bool
	}{
		{
			name:  "zero value",
			valid: true,
		},
		{
			name: "valid port range",
			cfg: aws.Config{
				PortRange: cluster.PortRange{Min: 5000, Max: 5000},
			},
			valid: true,
		},
		{
			name: "invalid port range",
			cfg: aws.Config{
				PortRange: cluster.PortRange{Min: 6000, Max: 5000},
			},
			valid: false,
		},
		{
			name:    "tag filter",
			filters: []aws.Filter{aws.Tag("Foo", "Bar")},
			valid:   true,
		},
		{
			name:    "failing filter",
			filters: []aws.Filter{failingFilter()},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.filters != nil {
				tc.cfg.SetFilters(tc.filters...)
			}
			err := tc.cfg.Validate()
			if tc.valid && err != nil {
				t.Fatal(err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("should have failed")
			}
		})
	}
}

func failingFilter() aws.Filter {
	return func(cfg *aws.Config) error {
		return fmt.Errorf("some error")
	}
}
