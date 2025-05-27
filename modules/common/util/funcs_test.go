/*
Copyright 2022 Red Hat

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// +kubebuilder:object:generate:=true

package util

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetOr(t *testing.T) {

	tests := []struct {
		name string
		data map[string]interface{}
		key  string
		want interface{}
	}{
		{
			name: "Key exists with value 111, returns 111",
			data: map[string]interface{}{"one": "111"},
			key:  "one",
			want: "111",
		},
		{
			name: "Key exists and empty string value, returns fallback",
			data: map[string]interface{}{"one": ""},
			key:  "one",
			want: "fallback",
		},
		{
			name: "Key does not exist, returns the fallback",
			data: map[string]interface{}{"one": "111"},
			key:  "four",
			want: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := GetOr(tt.data, tt.key, "fallback")
			g.Expect(newData).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestIsSet(t *testing.T) {

	tests := []struct {
		name string
		data map[string]interface{}
		key  string
		want interface{}
	}{
		{
			name: "Key exists, returns 111",
			data: map[string]interface{}{"one": "111"},
			key:  "one",
			want: "111",
		},
		{
			name: "Key does not exist, returns false",
			data: map[string]interface{}{"one": "111"},
			key:  "four",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := IsSet(tt.data, tt.key)
			g.Expect(newData).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestIsJSON(t *testing.T) {

	tests := []struct {
		name  string
		data  string
		error bool
	}{
		{
			name:  "Valid json string",
			data:  `{"some":"json"}`,
			error: false,
		},
		{
			name:  "Empty string",
			data:  "",
			error: true,
		},
		{
			name:  "Not valid json string",
			data:  "not valid json",
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := IsJSON(tt.data)
			if tt.error {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestRemoveIndex(t *testing.T) {

	tests := []struct {
		name  string
		data  []string
		index int
		want  []string
	}{
		{
			name:  "Remove inx 0",
			data:  []string{"111", "222", "333"},
			index: 0,
			want:  []string{"222", "333"},
		},
		{
			name:  "Remove inx 1",
			data:  []string{"111", "222", "333"},
			index: 1,
			want:  []string{"111", "333"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := RemoveIndex(tt.data, tt.index)
			for idx, d := range newData {
				g.Expect(d).To(BeIdenticalTo(tt.want[idx]))
			}
		})
	}
}

func TestRemoveValue(t *testing.T) {

	tests := []struct {
		name   string
		data   []string
		remove string
		want   []string
	}{
		{
			name:   "Remove empty",
			data:   []string{"111", "222", "333", "444"},
			remove: "",
			want:   []string{"111", "222", "333", "444"},
		},
		{
			name:   "Remove  111",
			data:   []string{"111", "222", "333", "444"},
			remove: "111",
			want:   []string{"222", "333", "444"},
		},
		{
			name:   "Remove 222",
			data:   []string{"111", "222", "333", "444"},
			remove: "222",
			want:   []string{"111", "333", "444"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := RemoveValue(tt.data, tt.remove)
			g.Expect(newData).To(Equal(tt.want))
		})
	}
}

func TestRemoveValues(t *testing.T) {

	tests := []struct {
		name   string
		data   []string
		remove []string
		want   []string
	}{
		{
			name:   "Remove empty",
			data:   []string{"111", "222", "333", "444"},
			remove: []string{""},
			want:   []string{"111", "222", "333", "444"},
		},
		{
			name:   "Remove  111",
			data:   []string{"111", "222", "333", "444"},
			remove: []string{"111"},
			want:   []string{"222", "333", "444"},
		},
		{
			name:   "Remove 222",
			data:   []string{"111", "222", "333", "444"},
			remove: []string{"222"},
			want:   []string{"111", "333", "444"},
		},
		{
			name:   "Remove 222 and 333",
			data:   []string{"111", "222", "333", "444"},
			remove: []string{"222", "333"},
			want:   []string{"111", "444"},
		},
		{
			name:   "Remove 111 and 444",
			data:   []string{"111", "222", "333", "444"},
			remove: []string{"111", "444"},
			want:   []string{"222", "333"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := RemoveValues(tt.data, tt.remove)
			g.Expect(newData).To(Equal(tt.want))
		})
	}
}

func TestKeepOnly(t *testing.T) {

	tests := []struct {
		name string
		data []string
		keep []string
		want []string
	}{
		{
			name: "Remove all",
			data: []string{"111", "222", "333", "444"},
			keep: []string{},
			want: []string{},
		},
		{
			name: "Keep 111",
			data: []string{"111", "222", "333", "444"},
			keep: []string{"111"},
			want: []string{"111"},
		},
		{
			name: "Keep 222 and 333",
			data: []string{"111", "222", "333", "444"},
			keep: []string{"222", "333"},
			want: []string{"222", "333"},
		},
		{
			name: "Keep 111 and non existing 44",
			data: []string{"111", "222", "333", "444"},
			keep: []string{"111", "44"},
			want: []string{"111"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := KeepOnlyValues(tt.data, tt.keep)
			g.Expect(newData).To(Equal(tt.want))
		})
	}
}

func TestKeepIfContainsSubstring(t *testing.T) {

	tests := []struct {
		name string
		data []string
		keep []string
		want []string
	}{
		{
			name: "Remove all",
			data: []string{"111", "222", "333", "444"},
			keep: []string{""},
			want: []string{},
		},
		{
			name: "Keep 111",
			data: []string{"111", "222", "333", "444"},
			keep: []string{"111"},
			want: []string{"111"},
		},
		{
			name: "Keep 222 and 333",
			data: []string{"111", "222", "333", "444"},
			keep: []string{"222", "333"},
			want: []string{"222", "333"},
		},
		{
			name: "Keep 111 and non existing 44",
			data: []string{"111", "222", "333", "444"},
			keep: []string{"111", "44"},
			want: []string{"111", "444"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newData := KeepIfContainsSubstring(tt.data, tt.keep)
			g.Expect(newData).To(Equal(tt.want))
		})
	}
}
