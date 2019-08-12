package main

import (
	"errors"
	"regexp"
	"strconv"
)

//Check is the class providing the methods for the checks
//Each method must return a (bool, error) pair
type Check struct{}

// Equals tracks if both strings are equal
func (t *Check) Equals(ruleValue, treeValue string) (bool, error) {
	return ruleValue == treeValue, nil
}

// Matches tracks if r matches s as a regular expression
func (t *Check) Matches(r, s string) (bool, error) {
	return regexp.MatchString(r, s)
}

// Length compares length to the len of the arry
func (t *Check) Length(length string, array []interface{}) (bool, error) {
	l, err := strconv.Atoi(length)
	return l == len(array), err
}

// Max compares max to val
func (t *Check) Max(max, val int64) (bool, error) {
	return max >= val, nil
}

// Min compares min to the val
func (t *Check) Min(min, val int64) (bool, error) {
	return min <= val, nil
}

// Eval evaluates an expression
func (t *Check) Eval(r, s string) (bool, error) {
	return false, errors.New("Not implemented")
}
