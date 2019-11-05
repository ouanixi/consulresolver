package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/resolver"
)

func TestBuildTarget(t *testing.T) {
	var tests = []struct {
		input resolver.Target
		want  *Target
	}{
		{input: resolver.Target{Scheme: "consul", Authority: "", Endpoint: "email?tags=grpc"}, want: &Target{serviceName: "email", tags: []string{"grpc"}}},
		{input: resolver.Target{Scheme: "consul", Authority: "", Endpoint: "email?tags=grpc,"}, want: &Target{serviceName: "email", tags: []string{"grpc"}}},
		{input: resolver.Target{Scheme: "consul", Authority: "", Endpoint: "email?wrongparam=grpc"}, want: &Target{serviceName: "email", tags: []string{}}},
	}

	for _, test := range tests {
		actual, _ := buildTarget(test.input)
		assert.Equal(t, test.want, actual)
	}
}
