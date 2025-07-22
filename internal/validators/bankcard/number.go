package bankcard

// ValidateBankCardNumber returns true if the input string is a valid number according to the Luhn algorithm.
func ValidateNumber(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	double := false

	// Process digits from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit := number[i] - '0'
		if digit > 9 {
			return false // invalid character detected
		}
		d := int(digit)

		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		double = !double
	}

	return sum%10 == 0
}
