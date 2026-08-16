package authentication

import (
	"regexp"
	"strings"
)

// Static test credentials and isolated bypass configurations
const (
	TestBypassOTP = "123456"
)

// NormalizePhoneNumber standardizes phone numbers into E.164 Indonesian format (+628...)
func NormalizePhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	if strings.HasPrefix(phone, "08") {
		return "+628" + phone[2:]
	}
	if strings.HasPrefix(phone, "628") {
		return "+628" + phone[3:]
	}
	if strings.HasPrefix(phone, "+628") {
		return phone
	}
	return phone
}

// IsValidIndonesianPhone checks if a phone number adheres to the +628 or 08 format
func IsValidIndonesianPhone(phone string) bool {
	norm := NormalizePhoneNumber(phone)
	re := regexp.MustCompile(`^\+628[0-9]{8,12}$`)
	return re.MatchString(norm)
}

var testUserPhones = map[string]string{
	"+6282123456781": "admin@acis.test",
	"+6282123456782": "member1@acis.test",
	"+6282123456783": "member2@acis.test",
}

// IsTestUser returns true if the phone number belongs to the designated test suite users
func IsTestUser(phoneOrEmail string) bool {
	normPhone := NormalizePhoneNumber(phoneOrEmail)
	if _, ok := testUserPhones[normPhone]; ok {
		return true
	}
	// Fallback for email check if email is provided
	normEmail := strings.TrimSpace(strings.ToLower(phoneOrEmail))
	return normEmail == "admin@acis.test" || normEmail == "member1@acis.test" || normEmail == "member2@acis.test"
}

// GetTestUserEmail returns the configured seed email for a test phone number
func GetTestUserEmail(phone string) string {
	norm := NormalizePhoneNumber(phone)
	if email, ok := testUserPhones[norm]; ok {
		return email
	}
	return ""
}

// ValidateTestBypassOTP validates whether a given OTP satisfies the test bypass condition
func ValidateTestBypassOTP(phoneOrEmail, otp string) bool {
	if !IsTestUser(phoneOrEmail) {
		return false
	}
	return strings.TrimSpace(otp) == TestBypassOTP
}

