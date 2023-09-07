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
