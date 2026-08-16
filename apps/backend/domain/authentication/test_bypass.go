package authentication

import (
	"strings"
)

// Static test credentials and isolated bypass configurations
const (
	TestBypassOTP = "123456"
)

var testUserEmails = map[string]bool{
	"admin@acis.test":   true,
	"member1@acis.test": true,
	"member2@acis.test": true,
}

// IsTestUser returns true if the email belongs to the designated test suite users
func IsTestUser(email string) bool {
	norm := strings.TrimSpace(strings.ToLower(email))
	if testUserEmails[norm] || strings.HasSuffix(norm, "@acis.test") {
		return true
	}
	return false
}

// ValidateTestBypassOTP validates whether a given OTP satisfies the test bypass condition
func ValidateTestBypassOTP(email, otp string) bool {
	if !IsTestUser(email) {
		return false
	}
	return strings.TrimSpace(otp) == TestBypassOTP
}
