// Package sensitive detects banking, health, and other high-risk
// sites so the Gatekeeper can escalate consent requirements before
// the agent performs any action on them.
//
// Per MISSION §10.7: "Domain allowlist (banking, health). Heuristic
// detection (form labels like 'credit card', 'SSN'). User overrides."
//
// The Detector is deterministic, pure logic, no model calls. It
// returns true when a URL or form context matches known sensitive
// patterns, triggering RequirePresenceAndConsent in the Gatekeeper.
package sensitive

import (
	"regexp"
	"strings"
)

// Detector checks whether a target URL, page title, or form context
// belongs to a known sensitive category.
type Detector struct{}

// NewDetector returns a Detector ready for use.
func NewDetector() *Detector {
	return &Detector{}
}

// IsSensitiveURL reports whether the given URL domain matches any
// known banking, health, government, or financial service pattern.
func (d *Detector) IsSensitiveURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	host := extractHost(rawURL)
	domain := cleanDomain(host)

	for _, p := range sensitiveDomains {
		if matchDomain(domain, p) {
			return true
		}
	}
	return false
}

// IsSensitiveContext checks whether form labels, page titles, or
// surrounding text indicate sensitive operations (credit card entry,
// SSN, tax, payment).
func (d *Detector) IsSensitiveContext(text string) bool {
	if text == "" {
		return false
	}
	textLower := strings.ToLower(text)
	for _, p := range sensitiveContexts {
		if p.MatchString(textLower) {
			return true
		}
	}
	return false
}

// Match reports whether the given URL and/or context text are
// sensitive. Either parameter may be empty.
func (d *Detector) Match(url, context string) bool {
	return d.IsSensitiveURL(url) || d.IsSensitiveContext(context)
}

// --- domain matching ---

// sensitiveDomains is the built-in allowlist of sensitive TLDs,
// second-level domains, and exact hostnames. Sorted by category.
var sensitiveDomains = []string{
	// Banking — TLDs
	".bank",
	".creditunion",

	// Banking — major providers
	"paypal.com",
	"stripe.com",
	"venmo.com",
	"cash.app",
	"wise.com",
	"revolut.com",
	"monzo.com",
	"chime.com",

	// Banking — US
	"chase.com",
	"bankofamerica.com",
	"wellsfargo.com",
	"citibank.com",
	"usbank.com",
	"pnc.com",
	"capitalone.com",
	"truist.com",
	"td.com",
	"schwab.com",
	"fidelity.com",
	"vanguard.com",
	"etrade.com",
	"robinhood.com",
	"coinbase.com",
	"binance.com",
	"kraken.com",

	// Banking — UK/EU
	"barclays.co.uk",
	"hsbc.com",
	"hsbc.co.uk",
	"lloydsbank.com",
	"natwest.com",
	"santander.com",
	"deutsche-bank.de",
	"bnpparibas.com",
	"ing.com",

	// Government
	".gov",
	".gov.uk",
	".mil",
	"irs.gov",
	"ssa.gov",
	"medicare.gov",

	// Health
	"epic.com",
	"mychart.",
	"mychart.",
	"cerner.com",
	"patient.",
	"health.",
	".healthcare",
	".medical",
	".clinic",
	".hospital",
	"khealth.com",
	"teladoc.com",

	// Insurance
	"geico.com",
	"progressive.com",
	"allstate.com",
	"statefarm.com",
	"libertymutual.com",

	// Tax
	"turbotax.com",
	"taxact.com",
	"hrblock.com",

	// Legal
	"docusign.net",
	"docusign.com",
	"hellosign.com",
}

// sensitiveContexts is compiled regexps that match form labels
// or page content indicating sensitive data entry.
var sensitiveContexts = []*regexp.Regexp{
	regexp.MustCompile(`credit\s*card`),
	regexp.MustCompile(`card\s*number`),
	regexp.MustCompile(`cvv`),
	regexp.MustCompile(`expir(?:y|ation)\s*date`),
	regexp.MustCompile(`security\s*code.*card`),
	regexp.MustCompile(`social\s*security`),
	regexp.MustCompile(`ssn`),
	regexp.MustCompile(`tax\s*id`),
	regexp.MustCompile(`routing\s*number`),
	regexp.MustCompile(`account\s*number.*bank`),
	regexp.MustCompile(`bank\s*account`),
	regexp.MustCompile(`passport\s*number`),
	regexp.MustCompile(`driver(?:'s)?\s*license`),
	regexp.MustCompile(`national\s*id`),
	regexp.MustCompile(`date\s*of\s*birth`),
	regexp.MustCompile(`medical\s*record`),
	regexp.MustCompile(`health\s*insurance`),
	regexp.MustCompile(`patient\s*(?:id|number)`),
	regexp.MustCompile(`wire\s*transfer`),
	regexp.MustCompile(`bank\s*transfer`),
	regexp.MustCompile(`ach\s*payment`),
	regexp.MustCompile(`purchase\s*(?:order|confirmation)`),
	regexp.MustCompile(`payment\s*method`),
}

// --- helpers ---

func extractHost(rawURL string) string {
	s := rawURL
	// Strip scheme
	if idx := strings.Index(s, "://"); idx != -1 {
		s = s[idx+3:]
	}
	// Strip path/query
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}
	// Strip port
	if idx := strings.LastIndex(s, ":"); idx != -1 {
		// Check if it's a port (only digits after colon)
		after := s[idx+1:]
		if isAllDigits(after) {
			s = s[:idx]
		}
	}
	return strings.ToLower(s)
}

func cleanDomain(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	// Strip leading "www."
	host = strings.TrimPrefix(host, "www.")
	return host
}

func matchDomain(domain, pattern string) bool {
	// Exact match
	if domain == pattern {
		return true
	}
	// Suffix match (e.g. ".bank" or ".gov")
	if strings.HasPrefix(pattern, ".") && strings.HasSuffix(domain, pattern) {
		return true
	}
	// Pattern is a subdomain prefix (e.g. "mychart.")
	if strings.HasSuffix(pattern, ".") && strings.Contains(domain, pattern) {
		return true
	}
	return false
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
