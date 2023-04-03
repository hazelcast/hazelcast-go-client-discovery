/*
 * Copyright (c) 2023, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
