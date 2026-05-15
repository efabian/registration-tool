package main

import (
	"regexp"
	"strings"
)

var (
	rxEmail = regexp.MustCompile(".+@.+\\..+")
	rxDigit = regexp.MustCompile(`\d`)
)

type Inputs struct {
	Email        string
	FirstName    string
	LastName     string
	Area         string
	Group        string
	Function     string
	Gender       string
	Local        string
	District     string
	Status       string
	PreferredDay string
	Errors       map[string]string
}

func (msg *Inputs) Validate() bool {
	msg.Errors = make(map[string]string)

	// FIX m-1: trim email before matching to reject leading/trailing spaces.
	msg.Email = strings.TrimSpace(msg.Email)
	match := rxEmail.MatchString(msg.Email)
	matchArea := rxDigit.FindString(msg.Area)
	matchGroup := rxDigit.FindString(msg.Group)

	// FIX m-2: use idiomatic !match instead of match == false.
	if !match {
		msg.Errors["Email"] = "Please enter a valid email address"
	}
	if strings.TrimSpace(msg.FirstName) == "" {
		msg.Errors["FirstName"] = "Please enter your first name"
	}
	if strings.TrimSpace(msg.LastName) == "" {
		msg.Errors["LastName"] = "Please enter your family name"
	}
	if matchArea == "" {
		msg.Errors["Area"] = "Please enter your area (ex: 1)"
	}
	if matchGroup == "" {
		msg.Errors["Group"] = "Please enter your group (ex: 1)"
	}
	// FIX m-8: validate Function field (was missing, would cause a DB NOT NULL error).
	if strings.TrimSpace(msg.Function) == "" {
		msg.Errors["Function"] = "Please enter your function"
	}
	if strings.TrimSpace(msg.Gender) == "" {
		msg.Errors["Gender"] = "Please indicate your gender"
	}
	if strings.TrimSpace(msg.Local) == "" {
		msg.Errors["Local"] = "Please indicate your local"
	}
	if strings.TrimSpace(msg.District) == "" {
		msg.Errors["District"] = "Please indicate your district"
	}
	if strings.TrimSpace(msg.Status) == "" {
		msg.Errors["Status"] = "Please indicate your status"
	}
	if strings.TrimSpace(msg.PreferredDay) == "" {
		msg.Errors["PreferredDay"] = "Please pick your preferred day & time"
	}
	if strings.TrimSpace(msg.PreferredDay) == "overbooked" {
		msg.Errors["PreferredDay"] = "It's overbooked. Please pick another day & time"
	}
	return len(msg.Errors) == 0
}
