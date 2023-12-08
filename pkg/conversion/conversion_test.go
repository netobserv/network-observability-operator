package conversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperToPascal(t *testing.T) {
	assert := assert.New(t)

	v := UpperToPascal("")
	assert.Equal("", v)

	v = UpperToPascal("EBPF")
	assert.Equal("eBPF", v)

	v = UpperToPascal("ENDED_CONVERSATIONS")
	assert.Equal("EndedConversations", v)

	v = UpperToPascal("SCRAM-SHA512")
	assert.Equal("ScramSHA512", v)
}

func TestPascalToUpper(t *testing.T) {
	assert := assert.New(t)

	v := PascalToUpper("", ' ')
	assert.Equal("", v)

	v = PascalToUpper("eBPF", ' ')
	assert.Equal("EBPF", v)

	v = PascalToUpper("EndedConversations", '_')
	assert.Equal("ENDED_CONVERSATIONS", v)

	v = PascalToUpper("ScramSHA512", '-')
	assert.Equal("SCRAM-SHA512", v)
}
