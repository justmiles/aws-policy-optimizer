package optimizer

import (
	"reflect"
	"testing"
)

func Test_generateGlobPattern(t *testing.T) {
	tests := []struct {
		name string
		ss   []string
		want string
	}{
		{
			name: "test1",
			ss:   []string{"abc/xyz", "abc/123", "abc/xxx"},
			want: "abc/*",
		},
		{
			name: "test2",
			ss:   []string{"abc/xyz/123", "abc/xyz/abc", "abc/xyz/xxx"},
			want: "abc/xyz/*",
		},
		{
			name: "test3",
			ss:   []string{"testing1233/asdf"},
			want: "testing1233/asdf",
		},
		{
			name: "test4",
			ss:   []string{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateGlobPattern(tt.ss); got != tt.want {
				t.Errorf("generateGlobPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consolidateARNs(t *testing.T) {
	tests := []struct {
		name string
		arns []string
		want []string
	}{
		{
			name: "test1",
			arns: []string{
				"arn:aws:s3:::test-bucket/mp/20240118210000_document1.txt",
				"arn:aws:s3:::test-bucket/mp/20240118210000_document2.txt",
				"arn:aws:s3:::test-bucket/mp/20240118210000_document2.txt",
				"arn:aws:s3:::test-bucket/mp/20240118210000_document2.txt.failures",
				"arn:aws:s3:::test2-bucket/mp/20240118210000_document2.txt.failures",
				"arn:aws:s3:::test2-bucket/mp/20240118210000_asdf",
				"arn:aws:s3:::test2-bucket/mp/202401182asdfes",
				"arn:aws:s3:::test2-bucket/mp/2024011821000asdfures",
			},
			want: []string{
				"arn:aws:s3:::test-bucket/mp/*",
				"arn:aws:s3:::test2-bucket/mp/*",
			},
		},
		{
			name: "test2",
			arns: []string{
				"arn:aws:s3:::test-bucket/mp/20240118210000_document1.txt",
				"arn:aws:s3:::test-bucket/mp/20240118210000_document2.txt",
				"arn:aws:s3:::test-bucket/mp/20240118210000_document2.txt",
				"arn:aws:s3:::test-bucket/mp/20240118210000_document2.txt.failures",
			},
			want: []string{
				"arn:aws:s3:::test-bucket/mp/*",
			},
		},
		{
			name: "test3",
			arns: []string{
				"arn:aws:kms:us-east-1:000000000000:key/xxxxxxxx-yyyy-zzzz-xxxx-yyyyyyyyyyyy",
			},
			want: []string{
				"arn:aws:kms:us-east-1:000000000000:key/xxxxxxxx-yyyy-zzzz-xxxx-yyyyyyyyyyyy",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := consolidateARNs(tt.arns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consolidateARNs() = %v, want %v", got, tt.want)
			}
		})
	}
}
