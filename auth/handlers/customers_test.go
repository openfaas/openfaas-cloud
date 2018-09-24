package handlers

import (
	"log"
	"testing"
)

func TestGet(t *testing.T) {
	c := NewCustomers()
	log.Println(c.Get("alexellis"))
	log.Println(c.Get("alexellisuk"))
	log.Println(c.Get("AlexEllisUK"))
	log.Println(c.Get("rgee0"))
}
