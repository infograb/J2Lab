package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func TestVersion(t *testing.T) {
	streams, _, buf, _ := utils.NewTestIOStreams()
	o := NewOptions(streams)

	err := o.run()
	assert.NoError(t, err)
	assert.Equal(t, Version+"\n", buf.String())
}
