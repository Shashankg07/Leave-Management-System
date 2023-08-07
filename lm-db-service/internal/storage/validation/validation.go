package validation

import (
	"errors"
	"regexp"
)

func ValidateFromDate(fromDate string) error {
	validDateFormat := regexp.MustCompile("((19|20)\\d\\d)\\-(0?[1-9]|1[012])\\-(0?[1-9]|[12][0-9]|3[01])")
	isValid := validDateFormat.MatchString(fromDate)
	if isValid == false {
		return errors.New("write date in YYYY-MM-DD format")
	}
	return nil
}
func ValidateToDate(toDate string) error {
	validDateFormat := regexp.MustCompile("((19|20)\\d\\d)\\-(0?[1-9]|1[012])\\-(0?[1-9]|[12][0-9]|3[01])")
	isValid := validDateFormat.MatchString(toDate)
	if isValid == false {
		return errors.New("write date in YYYY-MM-DD format")
	}
	return nil
}
