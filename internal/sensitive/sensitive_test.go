package sensitive

import "testing"

func TestIsSensitiveURL_BankingTLD(t *testing.T) {
	d := NewDetector()
	tests := []struct {
		url  string
		want bool
	}{
		{"https://chase.bank/accounts", true},       // .bank TLD
		{"https://www.any.creditunion/login", true}, // .creditunion TLD
		{"https://chase.com/accounts", true},        // exact chase.com
		{"https://www.paypal.com/checkout", true},   // paypal
		{"https://stripe.com/dashboard", true},      // stripe
		{"https://coinbase.com/trade", true},        // coinbase
		{"https://irs.gov/payments", true},          // .gov
		{"https://www.ssa.gov/myaccount", true},     // .gov
		{"https://medicare.gov/login", true},        // medicare
		{"https://mychart.epic.com/login", true},    // healthcare
		{"https://mychart.anyhospital.org", true},   // health
		{"https://www.geico.com/quote", true},       // insurance
		{"https://turbotax.com", true},              // tax
		{"https://docusign.net/sign", true},         // legal
	}
	for _, tt := range tests {
		if got := d.IsSensitiveURL(tt.url); got != tt.want {
			t.Fatalf("IsSensitiveURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestIsSensitiveURL_SafeSites(t *testing.T) {
	d := NewDetector()
	tests := []string{
		"https://github.com",
		"https://google.com",
		"https://stackoverflow.com/questions",
		"https://news.ycombinator.com",
		"https://en.wikipedia.org",
		"https://synaptic.app",
		"https://reddit.com",
		"",
	}
	for _, url := range tests {
		if d.IsSensitiveURL(url) {
			t.Fatalf("IsSensitiveURL(%q) should be false", url)
		}
	}
}

func TestIsSensitiveContext_CreditCard(t *testing.T) {
	d := NewDetector()
	tests := []string{
		"Enter your credit card number",
		"Card Number: 4111-1111-1111-1111",
		"CVV: 123",
		"Expiry Date: 12/28",
	}
	for _, ctx := range tests {
		if !d.IsSensitiveContext(ctx) {
			t.Fatalf("IsSensitiveContext(%q) should be true", ctx)
		}
	}
}

func TestIsSensitiveContext_SSN_Identity(t *testing.T) {
	d := NewDetector()
	tests := []string{
		"Social Security Number: 123-45-6789",
		"SSN: ***-**-1234",
		"Tax ID: 12-3456789",
		"Passport Number: A12345678",
		"Driver's License: D1234567",
	}
	for _, ctx := range tests {
		if !d.IsSensitiveContext(ctx) {
			t.Fatalf("IsSensitiveContext(%q) should be true", ctx)
		}
	}
}

func TestIsSensitiveContext_Banking(t *testing.T) {
	d := NewDetector()
	tests := []string{
		"Routing Number: 021000021",
		"Bank Account Number: 123456789",
		"Wire Transfer Details",
		"ACH Payment Confirmation",
	}
	for _, ctx := range tests {
		if !d.IsSensitiveContext(ctx) {
			t.Fatalf("IsSensitiveContext(%q) should be true", ctx)
		}
	}
}

func TestIsSensitiveContext_Health(t *testing.T) {
	d := NewDetector()
	tests := []string{
		"Medical Record #12345",
		"Health Insurance ID: 123456789",
		"Patient ID: 987654",
		"Date of Birth: 01/01/1990",
	}
	for _, ctx := range tests {
		if !d.IsSensitiveContext(ctx) {
			t.Fatalf("IsSensitiveContext(%q) should be true", ctx)
		}
	}
}

func TestIsSensitiveContext_Safe(t *testing.T) {
	d := NewDetector()
	tests := []string{
		"",
		"Hello, how are you?",
		"Search results for golang",
		"Your order has shipped",
		"Welcome to Synaptic!",
		"Click here to continue",
	}
	for _, ctx := range tests {
		if d.IsSensitiveContext(ctx) {
			t.Fatalf("IsSensitiveContext(%q) should be false", ctx)
		}
	}
}

func TestMatch_Combined(t *testing.T) {
	d := NewDetector()
	if !d.Match("https://chase.com", "") {
		t.Fatal("sensitive URL alone should match")
	}
	if !d.Match("", "Enter your credit card number") {
		t.Fatal("sensitive context alone should match")
	}
	if d.Match("", "") {
		t.Fatal("empty both should not match")
	}
	if d.Match("https://github.com", "Hello world") {
		t.Fatal("safe URL + safe context should not match")
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		url, want string
	}{
		{"https://chase.com/accounts", "chase.com"},
		{"http://localhost:8080/test", "localhost"},
		{"https://www.google.com", "www.google.com"},
		{"chase.com", "chase.com"},
		{"", ""},
	}
	for _, tt := range tests {
		got := extractHost(tt.url)
		if got != tt.want {
			t.Fatalf("extractHost(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		domain, pattern string
		want            bool
	}{
		{"chase.com", "chase.com", true},
		{"chase.com", "chase.co", false},
		{"example.bank", ".bank", true},
		{"example.gov", ".gov", true},
		{"example.com", ".gov", false},
		{"mychart.epic.com", "mychart.", true},
		{"sandbox.epic.com", "mychart.", false},
	}
	for _, tt := range tests {
		if got := matchDomain(tt.domain, tt.pattern); got != tt.want {
			t.Fatalf("matchDomain(%q, %q) = %v, want %v", tt.domain, tt.pattern, got, tt.want)
		}
	}
}
