package utils

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func RandomColor() *string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	red, green, blue := random.Intn(256), random.Intn(256), random.Intn(256)
	colorHex := fmt.Sprintf("#%02X%02X%02X", red, green, blue)
	return &colorHex
}
