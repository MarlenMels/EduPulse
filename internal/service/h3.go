package service

import (
	"fmt"

	zr "github.com/ziprecruiter/h3-go/pkg/h3"
)

func H3FromLatLng(lat, lng float64, resolution int) (string, error) {
	ll := zr.NewLatLng(lat, lng)
	cell, err := zr.NewCellFromLatLng(ll, resolution)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(cell), nil
}