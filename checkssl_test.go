package main

import (
	"testing"
	"time"
)

func Test_Display(t *testing.T) {

	date := time.Date(2020,1,2, 3,4,5,6, time.UTC)
	actual := displayDate(date)

	if actual != ""{
		t.Fail()
	}
}
