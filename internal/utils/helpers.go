package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func CheckErr(err error) {
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
}

func RandomColor() *string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	red, green, blue := random.Intn(256), random.Intn(256), random.Intn(256)
	colorHex := fmt.Sprintf("#%02X%02X%02X", red, green, blue)
	return &colorHex
}
