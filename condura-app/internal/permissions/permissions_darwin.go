//go:build darwin

package permissions

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics -framework AVFoundation -framework Foundation -framework UserNotifications

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#import <AVFoundation/AVFoundation.h>
#import <Foundation/Foundation.h>
#import <UserNotifications/UserNotifications.h>

// probeTimeoutNs is the wall-clock budget (in nanoseconds) for any
// cgo call that may need to wait on a system completion handler.
// 2s is generous on real hardware and short enough that the UI
// stays responsive if the call hangs (e.g. UNUserNotificationCenter
// never calling back because the bundle has no notification
// entitlement).
static const int64_t probeTimeoutNs = 2 * 1000 * 1000 * 1000LL;

// conduraMicAuthStatus returns AVAuthorizationStatus as int:
//   0 = AVAuthorizationStatusNotDetermined
//   1 = AVAuthorizationStatusRestricted
//   2 = AVAuthorizationStatusDenied
//   3 = AVAuthorizationStatusAuthorized
//
// This is the canonical Apple-blessed way to probe microphone
// permission. Replaces the prior system_profiler substring hack
// that returned StatusGranted on any Mac with audio hardware
// regardless of TCC state.
static int conduraMicAuthStatus(void) {
  return (int)[AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio];
}

// conduraNotifAuthStatus returns UNAuthorizationStatus as int
// synchronously, blocking up to probeTimeoutNs for the system
// completion handler to fire. Return values:
//  -3 = API unusable (binary not in an app bundle / no bundle identifier)
//  -2 = timeout (handler never called back)
//  -1 = API unavailable (pre-10.14)
//   0 = UNAuthorizationStatusNotDetermined
//   1 = UNAuthorizationStatusDenied
//   2 = UNAuthorizationStatusAuthorized
//   3 = UNAuthorizationStatusProvisional
//   4 = UNAuthorizationStatusEphemeral
//
// UNUserNotificationCenter.currentNotificationCenter is only safe
// to call from a binary running inside a .app bundle with the
// user-notifications entitlement. A bare daemon binary (or a Go
// test binary launched via `go test`) lacks the bundleIdentifier
// and the runtime trips an internal precondition when the system
// tries to attribute the request to a non-existent app. We gate
// the call on bundleIdentifier being non-nil to fail safe.
static int conduraNotifAuthStatus(void) {
  if (@available(macOS 10.14, *)) {
    NSString *bundleID = [[NSBundle mainBundle] bundleIdentifier];
    if (bundleID == nil || [bundleID length] == 0) {
      return -3;
    }
    __block int status = -1;
    dispatch_semaphore_t sem = dispatch_semaphore_create(0);
    [[UNUserNotificationCenter currentNotificationCenter]
        getNotificationSettingsWithCompletionHandler:^(UNNotificationSettings * _Nonnull settings) {
      status = (int)settings.authorizationStatus;
      dispatch_semaphore_signal(sem);
    }];
    if (dispatch_semaphore_wait(sem, dispatch_time(DISPATCH_TIME_NOW, probeTimeoutNs)) != 0) {
      return -2;
    }
    return status;
  }
  return -1;
}
*/
import "C"

import (
	"context"
	"os/exec"
	"time"
)

// execProbe is the function used to spawn the osascript subprocess
// probe for AppleEvent permission. Tests override this with a
// stub that returns canned output; the default spawns the real
// osascript with a 3-second timeout.
var execProbe = func(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // fixed probe name, not user input
	return cmd.Output()
}

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
	case KindScreenRecording:
		// CGPreflightScreenCaptureAccess returns true if the
		// process is allowed to capture the screen. Available
		// on macOS 10.15+. Does NOT trigger a TCC prompt.
		if C.CGPreflightScreenCaptureAccess() {
			return Permission{Kind: k, Status: StatusGranted, Note: "CGPreflightScreenCaptureAccess"}
		}
		return Permission{Kind: k, Status: StatusDenied, Note: "grant in System Settings → Privacy & Security → Screen & System Audio Recording"}
	case KindMicrophone:
		return probeMicrophone()
	case KindAutomation:
		return probeAutomation()
	case KindNotifications:
		return probeNotifications()
	default:
		return defaultProbeOne(k)
	}
}

// probeMicrophone asks AVFoundation for the canonical
// authorization status for the AVMediaTypeAudio media type.
// This is the same API every Mac app uses; the result reflects
// actual TCC state (notDetermined | restricted | denied | authorized),
// not hardware presence.
func probeMicrophone() Permission {
	switch C.conduraMicAuthStatus() {
	case 3: // AVAuthorizationStatusAuthorized
		return Permission{Kind: KindMicrophone, Status: StatusGranted, Note: "AVCaptureDevice authorizationStatus = authorized"}
	case 2: // AVAuthorizationStatusDenied
		return Permission{Kind: KindMicrophone, Status: StatusDenied, Note: "AVCaptureDevice authorizationStatus = denied; grant via System Settings → Privacy & Security → Microphone"}
	case 1: // AVAuthorizationStatusRestricted
		return Permission{Kind: KindMicrophone, Status: StatusDenied, Note: "AVCaptureDevice authorizationStatus = restricted (parental controls / MDM); microphone blocked"}
	default: // 0 = NotDetermined
		return Permission{Kind: KindMicrophone, Status: StatusUnknown, Note: "AVCaptureDevice authorizationStatus = notDetermined; trigger requestAccess to surface the TCC prompt"}
	}
}

// probeAutomation checks whether the process can send AppleEvents
// to other applications via a harmless osascript. Wrapped in a
// 3-second timeout so a missing/stalled osascript does not hang
// the permission probe indefinitely.
func probeAutomation() Permission {
	if _, err := execProbe("osascript", "-e",
		`tell application "System Events" to get version`); err == nil {
		return Permission{Kind: KindAutomation, Status: StatusGranted, Note: "AppleEvent to System Events succeeded"}
	}
	return Permission{
		Kind:   KindAutomation,
		Status: StatusUnknown,
		Note:   "grant via System Settings → Privacy & Security → Automation; use request_guide for steps",
	}
}

// probeNotifications asks UNUserNotificationCenter for the
// authorization status, synchronously via a semaphore in cgo.
// This is the canonical API for notification permission on
// macOS 10.14+; it does NOT trigger a TCC prompt.
func probeNotifications() Permission {
	switch C.conduraNotifAuthStatus() {
	case 2: // UNAuthorizationStatusAuthorized
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "UNUserNotificationCenter authorizationStatus = authorized"}
	case 3: // UNAuthorizationStatusProvisional
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "UNUserNotificationCenter authorizationStatus = provisional (quiet notifications allowed)"}
	case 4: // UNAuthorizationStatusEphemeral
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "UNUserNotificationCenter authorizationStatus = ephemeral (App Clip-style)"}
	case 1: // UNAuthorizationStatusDenied
		return Permission{Kind: KindNotifications, Status: StatusDenied, Note: "UNUserNotificationCenter authorizationStatus = denied; grant via System Settings → Notifications"}
	case -3: // class unavailable (binary not entitled)
		return Permission{Kind: KindNotifications, Status: StatusUnknown, Note: "UNUserNotificationCenter class not loaded; binary needs user-notifications entitlement (packaged Mac app)"}
	case -2: // cgo timeout
		return Permission{Kind: KindNotifications, Status: StatusUnknown, Note: "UNUserNotificationCenter completion handler did not fire within 2s"}
	case -1: // API unavailable (pre-10.14)
		return Permission{Kind: KindNotifications, Status: StatusUnknown, Note: "UNUserNotificationCenter requires macOS 10.14+"}
	default: // 0 = NotDetermined
		return Permission{Kind: KindNotifications, Status: StatusUnknown, Note: "UNUserNotificationCenter authorizationStatus = notDetermined; trigger requestAuthorization to surface the prompt"}
	}
}
