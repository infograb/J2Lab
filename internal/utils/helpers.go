package utils

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func RandomColor() *string {
	rand.Seed(time.Now().UnixNano())
	red, green, blue := rand.Intn(256), rand.Intn(256), rand.Intn(256)
	colorHex := fmt.Sprintf("#%02X%02X%02X", red, green, blue)
	return &colorHex
}
