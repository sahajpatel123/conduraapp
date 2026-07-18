package daemon

import (
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/failover"
)

// -----------------------------------------------------------------------------
// CheckBudget — DoS protection budget check (spendBudgetChecker adapter)
//
// spendBudgetChecker wraps failover.SpendMonitor to satisfy the
// adaptive.BudgetChecker interface. A return value of nil means
// the agent may continue making API calls; a non-nil error means
// the daily USD spend cap has been reached and the adaptive engine
// should pause. Without coverage here, a regression that swapped
// the two return branches would let the agent burn through the
// budget silently — exactly the kind of bug a DoS budget is meant
// to prevent.
// -----------------------------------------------------------------------------

func TestCheckBudget_UnderCap_ReturnsNil(t *testing.T) {
	m := failover.NewSpendMonitor(failover.SpendCap{USDPerDay: 5.0})
	// No Record() — fresh monitor, nothing spent.
	c := &spendBudgetChecker{m: m}
	if err := c.CheckBudget(); err != nil {
		t.Fatalf("CheckBudget on fresh monitor should return nil; got %v", err)
	}
}

func TestCheckBudget_AtCap_ReturnsNil(t *testing.T) {
	m := failover.NewSpendMonitor(failover.SpendCap{USDPerDay: 5.0})
	m.Record(5.0) // exactly at the cap
	c := &spendBudgetChecker{m: m}
	// Allow(amount) is `spent + amount <= cap`. With amount=0 and
	// spent=cap, this is `cap <= cap` = true. The boundary case
	// must NOT reject — it's only over-cap that triggers the error.
	if err := c.CheckBudget(); err != nil {
		t.Fatalf("CheckBudget at exactly the cap should return nil (boundary inclusive); got %v", err)
	}
}

func TestCheckBudget_OverCap_ReturnsBudgetExceeded(t *testing.T) {
	m := failover.NewSpendMonitor(failover.SpendCap{USDPerDay: 5.0})
	m.Record(6.0) // one dollar over the cap
	c := &spendBudgetChecker{m: m}
	err := c.CheckBudget()
	if err == nil {
		t.Fatal("CheckBudget over the cap must return an error")
	}
	if !strings.Contains(err.Error(), "budget") {
		t.Fatalf("error message should mention \"budget\"; got %q", err.Error())
	}
}
