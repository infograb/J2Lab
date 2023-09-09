/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infograb-public/j2lab/internal/utils"
)

func TestVersion(t *testing.T) {
	streams, _, buf, _ := utils.NewTestIOStreams()
	o := NewOptions(streams)

	err := o.run()
	assert.NoError(t, err)
	assert.Equal(t, Version+"\n", buf.String())
}
