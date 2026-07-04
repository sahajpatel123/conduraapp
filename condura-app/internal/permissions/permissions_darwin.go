//go:build darwin

package permissions

/*
#cgo LDFLAGS: -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>
*/
import "C"

func init() {
	probeOneImpl = darwinProbeOne
}

func darwinProbeOne(k Kind) Permission {
	switch k {
	case KindAccessibility:
		if C.AXIsProcessTrusted() != 0 {
			return Permission{Kind: k, Status: StatusGranted, Note: "AXIsProcessTrusted"}
		}
		return Permission{Kind: k, Status: StatusDenied, Note: "grant in System Settings → Privacy & Security → Accessibility"}
	case KindScreenRecording, KindMicrophone, KindAutomation, KindNotifications:
		// macOS does not expose a reliable per-process probe without
		// attempting the protected action. Guides are returned via
		// permissions.request_guide.
		return Permission{Kind: k, Status: StatusUnknown, Note: "grant via System Settings; use request_guide for steps"}
	default:
		return defaultProbeOne(k)
	}
}
