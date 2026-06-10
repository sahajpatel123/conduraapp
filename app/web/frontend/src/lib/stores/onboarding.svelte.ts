// Onboarding state. Tracks which step of the first-run wizard
// the user is on, and which fields they've filled in.

export type OnboardingStep = 'welcome' | 'provider' | 'apikey' | 'test' | 'voice' | 'hotkey' | 'privacy' | 'done'

interface OnboardingState {
  step: OnboardingStep
  provider: string
  apiKey: string
  hotkey: string
  telemetryEnabled: boolean
  testPassed: boolean
}

class OnboardingStore {
  state = $state<OnboardingState>({
    step: 'welcome',
    provider: 'openai',
    apiKey: '',
    hotkey: 'Cmd+Shift+Space',
    telemetryEnabled: false,
    testPassed: false
  })

  reset(): void {
    this.state = {
      step: 'welcome',
      provider: 'openai',
      apiKey: '',
      hotkey: 'Cmd+Shift+Space',
      telemetryEnabled: false,
      testPassed: false
    }
  }

  nextStep(): void {
    const order: OnboardingStep[] = ['welcome', 'provider', 'apikey', 'test', 'voice', 'hotkey', 'privacy', 'done']
    const idx = order.indexOf(this.state.step)
    if (idx >= 0 && idx < order.length - 1) {
      this.state.step = order[idx + 1]
    }
  }

  prevStep(): void {
    const order: OnboardingStep[] = ['welcome', 'provider', 'apikey', 'test', 'voice', 'hotkey', 'privacy', 'done']
    const idx = order.indexOf(this.state.step)
    if (idx > 0) {
      this.state.step = order[idx - 1]
    }
  }
}

export const onboarding = new OnboardingStore()
