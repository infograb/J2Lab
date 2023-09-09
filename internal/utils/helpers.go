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

package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

func CheckErr(err error) {
	if logrus.GetLevel() >= logrus.DebugLevel {
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
		}
	} else {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func RandomColor() *string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	red, green, blue := random.Intn(256), random.Intn(256), random.Intn(256)
	colorHex := fmt.Sprintf("#%02X%02X%02X", red, green, blue)
	return &colorHex
}
