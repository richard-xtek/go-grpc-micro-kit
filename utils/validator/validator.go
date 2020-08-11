package validator

import (
	"reflect"
	"regexp"
	"unicode"

	"strings"

	"github.com/asaskevich/govalidator"
)

const (
	phoneChars = "0123456789"
	vneseChars = "đĐ" +
		"àáạảãâầấậẩẫăằắặẳẵ" +
		"ÀÁẠẢÃÂẦẤẬẨẪĂẰẮẶẲẴ" +
		"èéẹẻẽêềếệểễ" +
		"ÈÉẸẺẼÊỀẾỆỂỄ" +
		"òóọỏõôồốộổỗơờớợởỡ" +
		"ÒÓỌỎÕÔỒỐỘỔỖƠỜỚỢỞỠ" +
		"ùúụủũưừứựửữ" +
		"ÙÚỤỦŨƯỪỨỰỬỮ" +
		"ìíịỉĩ" + "ỳýỵỷỹ" +
		"ÌÍỊỈĨ" + "ỲÝỴỶỸ"

	numChars   = "0123456789"
	upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars = "abcdefghijklmnopqrstuvwxyz"
	signChars  = ` .,/\"'_-+=@#%*()[]{}<>!?$`
	nameChars  = signChars + numChars + upperChars + lowerChars + vneseChars
)

var (
	phoneRegexp    = regexp.MustCompile(`^0[0-9]{7,14}$`)
	phoneWhiteList = regexp.MustCompile(`[^0-9]+`)

	nameRegexp     = regexp.MustCompile(`^[\-\[\]/\\ .,"'_+=@#%*(){}<>!?$` + numChars + upperChars + lowerChars + vneseChars + `]{2,200}$`)
	nameWhiteList  = regexp.MustCompile(`[^\-\[\]/\\ .,"'_+=@#%*(){}<>!?$` + numChars + upperChars + lowerChars + vneseChars + `]+`)
	spaceWhiteList = regexp.MustCompile(`\s\s+`)
)

func init() {
	SetupDefault()
}

// SetupDefault ...
func SetupDefault() {
	govalidator.CustomTypeTagMap.Set("phone",
		func(v interface{}, ctx interface{}) bool {
			if s, ok := assertString(v); ok {
				return phoneRegexp.MatchString(s)
			}
			return false
		})

	govalidator.CustomTypeTagMap.Set("code",
		func(v interface{}, ctx interface{}) bool {
			if s, ok := assertString(v); ok {
				return Code(s)
			}
			return false
		})

	govalidator.CustomTypeTagMap.Set("name",
		func(v interface{}, ctx interface{}) bool {
			if s, ok := assertString(v); ok {
				if len(s) < 2 || len(s) > 200 {
					return false
				}
				return nameRegexp.MatchString(s)
			}
			return false
		})
}

func assertString(s interface{}) (string, bool) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.String {
		return "", false
	}
	return v.String(), true
}

// Check ...
func Check(v interface{}) error {
	_, err := govalidator.ValidateStruct(v)
	return err
}

// Code ...
func Code(s string) bool {
	if len(s) < 3 || len(s) > 64 {
		return false
	}
	for i, l := 0, len(s); i < l; i++ {
		// Only allow printable ASCII chars (from `!` to `~`)
		if s[i] < 33 || s[i] > 126 {
			return false
		}
	}
	return true
}

// NormalizePassword ...
func NormalizePassword(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) <= 5 || len(s) > 35 {
		return "", false
	}
	hasUpper := false
	hasLower := false
	hasNumber := false
	// hasSpecial := false
	hasSpace := true
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		// case unicode.IsPunct(char) || unicode.IsSymbol(char):
		// 	hasSpecial = truew
		case unicode.IsSpace(char):
			hasSpace = false
		}
	}
	if hasUpper && hasLower && hasNumber && hasSpace {
		return s, true
	}
	return "", false
}

// NormalizeEmail ...
func NormalizeEmail(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return "", false
	}
	s, err := govalidator.NormalizeEmail(s)
	if err != nil {
		return "", false
	}

	return s, true
}

// NormalizePhone ...
func NormalizePhone(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+84") {
		s = s[3:] // Remove +84
	}
	s = WhiteList(s, phoneWhiteList)

	if len(s) == 0 {
		return "", false
	}
	if s[0] != '0' {
		s = "0" + s
	}
	return s, phoneRegexp.MatchString(s)
}

// NormalizeName ...
func NormalizeName(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = WhiteList(s, nameWhiteList)
	s = TrimInnerSpace(s)

	if len(s) == 0 {
		return "", false
	}
	return s, nameRegexp.MatchString(s)
}

// WhiteList remove characters that do not appear in the whitelist.
func WhiteList(s string, r *regexp.Regexp) string {
	return r.ReplaceAllString(s, "")
}

// TrimInnerSpace remove inner space characters
func TrimInnerSpace(s string) string {
	return spaceWhiteList.ReplaceAllString(s, " ")
}

// ContainsString ...
func ContainsString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// ContainsInt ...
func ContainsInt(a []int, x int) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// ContainsInt32 ...
func ContainsInt32(a []int32, x int32) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
