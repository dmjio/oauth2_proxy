package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"unsafe"
)

type UserMap struct {
	usersFile string
	m         unsafe.Pointer
}

func NewUserMap(usersFile string, done <-chan bool, onUpdate func()) *UserMap {
	um := &UserMap{usersFile: usersFile}
	m := make(map[string]bool)
	atomic.StorePointer(&um.m, unsafe.Pointer(&m))
	if usersFile != "" {
		log.Printf("using authenticated emails file %s", usersFile)
		WatchForUpdates(usersFile, done, func() {
			um.LoadAuthenticatedEmailsFile()
			onUpdate()
		})
		um.LoadAuthenticatedEmailsFile()
	}
	return um
}

func (um *UserMap) IsValid(email string) (result bool) {
	m := *(*map[string]bool)(atomic.LoadPointer(&um.m))
	_, result = m[email]
	return
}

func (um *UserMap) LoadAuthenticatedEmailsFile() {
	r, err := os.Open(um.usersFile)
	if err != nil {
		log.Fatalf("failed opening authenticated-emails-file=%q, %s", um.usersFile, err)
	}
	defer r.Close()
	csv_reader := csv.NewReader(r)
	csv_reader.Comma = ','
	csv_reader.Comment = '#'
	csv_reader.TrimLeadingSpace = true
	records, err := csv_reader.ReadAll()
	if err != nil {
		log.Printf("error reading authenticated-emails-file=%q, %s", um.usersFile, err)
		return
	}
	updated := make(map[string]bool)
	for _, r := range records {
		address := strings.ToLower(strings.TrimSpace(r[0]))
		updated[address] = true
	}
	atomic.StorePointer(&um.m, unsafe.Pointer(&updated))
}

func newValidatorImpl(domains []string, usersFile string,
	done <-chan bool, onUpdate func()) func(string) bool {
//	validUsers := NewUserMap(usersFile, done, onUpdate)
	validator := func(email string) (valid bool) {
		valid = false
		for _, domain := range domains {
//   		        log.PrintF("checking email: %s agains %s", email, domain)
			if domain == email {
				valid = true
			}
		}
		return valid
	}
	return validator
}

func NewValidator(domains []string, usersFile string) func(string) bool {
	return newValidatorImpl(domains, usersFile, nil, func() {})
}
