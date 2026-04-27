package rules

import "configaudit/internal/audit"

type TLSDisabledRule struct{}

func (TLSDisabledRule) ID() string {
	return "tls-disabled"
}

func (TLSDisabledRule) Check(_ audit.Context, node audit.Node) []audit.Problem {
	value, ok := node.Value.(bool)
	if !ok {
		return nil
	}

	switch {
	case isTLSEnabledField(node) && !value:
		return []audit.Problem{{
			Severity:       audit.SeverityHigh,
			Path:           node.Path,
			Message:        "TLS is disabled.",
			Recommendation: "Do not disable TLS or certificate verification in production. Enable TLS verification and use valid certificates.",
		}}
	case isDisableTLSField(node) && value:
		return []audit.Problem{{
			Severity:       audit.SeverityHigh,
			Path:           node.Path,
			Message:        "TLS is disabled.",
			Recommendation: "Do not disable TLS or certificate verification in production. Enable TLS verification and use valid certificates.",
		}}
	case isTLSVerificationField(node) && !value:
		return []audit.Problem{{
			Severity:       audit.SeverityHigh,
			Path:           node.Path,
			Message:        "TLS certificate verification is disabled.",
			Recommendation: "Do not disable TLS or certificate verification in production. Enable TLS verification and use valid certificates.",
		}}
	case isSkipTLSVerificationField(node) && value:
		return []audit.Problem{{
			Severity:       audit.SeverityHigh,
			Path:           node.Path,
			Message:        "TLS certificate verification is disabled.",
			Recommendation: "Do not disable TLS or certificate verification in production. Enable TLS verification and use valid certificates.",
		}}
	default:
		return nil
	}
}

func isTLSEnabledField(node audit.Node) bool {
	return pathHasSuffix(node.Path, "tls.enabled") || pathHasSuffix(node.Path, "ssl.enabled")
}

func isDisableTLSField(node audit.Node) bool {
	return normalizeIdentifier(node.Key) == "disabletls"
}

func isTLSVerificationField(node audit.Node) bool {
	switch normalizeIdentifier(node.Key) {
	case "verifytls", "tlsverify", "sslverify", "verifycertificate", "certificateverify", "verifycert", "certverify":
		return true
	case "verify":
		return pathHasSuffix(node.ParentPath, "tls")
	default:
		return false
	}
}

func isSkipTLSVerificationField(node audit.Node) bool {
	switch normalizeIdentifier(node.Key) {
	case "insecureskipverify", "skiptlsverify":
		return true
	default:
		return false
	}
}
