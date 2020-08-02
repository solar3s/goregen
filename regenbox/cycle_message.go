package regenbox

import (
	"fmt"
	"github.com/rkjdid/util"
)

const (
	CycleCharge    = "Charge"
	CycleChargeX4  = "ChargeX4"
	CycleDischarge = "Discharge"
	CycleMulti     = "Multi-cycle"
)

type CycleMessage struct {
	Type     string
	Target   int
	Status   string
	Erronous bool
	Final    bool
}

func cycleMessage(t string, target int, message string, bErr bool, final bool) CycleMessage {
	return CycleMessage{
		Type:     t,
		Target:   target,
		Status:   message,
		Erronous: bErr,
		Final:    final,
	}
}

func cycleStarted(t string, target int) CycleMessage {
	return cycleMessage(t, target, "Started...", false, false)
}

func cycleReached(t string, target int) CycleMessage {
	return cycleMessage(t, target, "Target voltage reached", false, true)
}

func cycleTimeout(t string, target int, timeout util.Duration) CycleMessage {
	return cycleMessage(t, target, fmt.Sprintf("Didn't reach target after %s", timeout), true, true)
}

func chargeStarted(target int) CycleMessage {
	return cycleStarted(CycleCharge, target)
}

func chargeStartedX4(target int) CycleMessage {
	return cycleStarted(CycleChargeX4, target)
}

func dischargeStarted(target int) CycleMessage {
	return cycleStarted(CycleDischarge, target)
}

func multiCycleStarted(target int, t string, n int, of int) CycleMessage {
	return cycleMessage(CycleMulti, target, fmt.Sprintf("%s %d/%d...", t, n, of), false, false)
}

func chargeReached(target int) CycleMessage {
	return cycleReached(CycleCharge, target)
}

func chargeReachedX4(battery int, target int) CycleMessage {
	return cycleMessage(CycleChargeX4, target, fmt.Sprintf("Target voltage reached for battery #%d...", battery), false, true)
}

func dischargeReached(target int) CycleMessage {
	return cycleReached(CycleDischarge, target)
}

func multiCycleReached(target int, n int) CycleMessage {
	return cycleMessage(CycleMulti, target, fmt.Sprintf("Completed %d half-cycles", n), false, true)
}

func chargeTimeout(target int, duration util.Duration) CycleMessage {
	return cycleTimeout(CycleCharge, target, duration)
}

func dischargeTimeout(target int, duration util.Duration) CycleMessage {
	return cycleTimeout(CycleDischarge, target, duration)
}

func multiCycleTimeout(target int, t string, n int, of int, duration util.Duration) CycleMessage {
	return multiCycleError(target, fmt.Errorf("%s %d/%d didn't reach target after %s", t, n, of, duration))
}

func chargeError(target int, err error) CycleMessage {
	return cycleMessage(CycleCharge, target, err.Error(), true, true)
}

func dischargeError(target int, err error) CycleMessage {
	return cycleMessage(CycleDischarge, target, err.Error(), true, true)
}

func multiCycleError(target int, err error) CycleMessage {
	return cycleMessage(CycleMulti, target, err.Error(), true, true)
}
