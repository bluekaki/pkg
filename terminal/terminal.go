package terminal

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"syscall"

	"github.com/byepichi/pkg/errors"

	"golang.org/x/crypto/ssh/terminal"
)

// ReadPwd read password from terminal
func ReadPwd(doubleCheck bool, desc ...string) ([]string, error) {
	if desc == nil {
		fmt.Println("\nEnter a four-digit password, separated by space or enter.")

	} else {
		fmt.Println(fmt.Sprintf("\nEnter a four-digit password for %s, separated by space or enter.", desc[0]))
	}

	pwd0, err := read()
	if err != nil {
		return nil, err
	}

	if doubleCheck {
		fmt.Println("\nRepeat the four-digit password, separated by space or enter.")

	} else {
		return pwd0, nil
	}

	pwd1, err := read()
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(pwd0, pwd1) {
		return nil, errors.New("Passwords do not match")
	}

	return pwd0, nil
}

func read() ([]string, error) {
	index := 0
	pwds := make([]string, 4)

	for {
		raw, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, val := range strings.Split(strings.TrimSpace(string(raw)), " ") {
			if val = strings.TrimSpace(val); val != "" {
				pwds[index] = val

				if index++; index == 4 {
					return pwds, nil
				}
			}
		}
	}
}

// ReadSomething read something from terminal
func ReadSomething(desc string) (string, error) {
	fmt.Println(desc)

	raw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(bytes.TrimSpace(raw)), nil
}
