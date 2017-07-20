package regenbox

import (
	"fmt"
	"github.com/rkjdid/util"
)

type CycleMessage struct {
	Target  int
	Content string
	Error   error
	Final   bool
}

func cycleMessage(target int, message string, err error, final bool) CycleMessage {
	return CycleMessage{
		Target:  target,
		Content: message,
		Error:   err,
		Final:   final,
	}
}

func chargeStarted(target int) CycleMessage {
	return cycleMessage(target, "Charge cycle started...", nil, false)
}

func dischargeStarted(target int) CycleMessage {
	return cycleMessage(target, "Discharge cycle started...", nil, false)
}

func multiCycleStarted(target int, prefix string, n int, of int) CycleMessage {
	return cycleMessage(target, fmt.Sprintf("Multi-cycle: %s cycle %d/%d...", prefix, n, of), nil, false)
}

func chargeEnded(target int) CycleMessage {
	return cycleMessage(target, "Charge: target voltage reached", nil, true)
}

func dischargeEnded(target int) CycleMessage {
	return cycleMessage(target, "Discharge: target voltage reached", nil, true)
}

func multiCycleEnded(target int, n int) CycleMessage {
	return cycleMessage(target, fmt.Sprintf("Multi-cycle: finished all %d half-cycles", n), nil, true)
}

func chargeTimeout(target int, duration util.Duration) CycleMessage {
	return chargeError(target, fmt.Errorf("didn't reach target after %s", duration))
}

func dischargeTimeout(target int, duration util.Duration) CycleMessage {
	return dischargeError(target, fmt.Errorf("didn't reach target after %s", duration))
}

func multiCycleTimeout(target int, prefix string, n int, of int, duration util.Duration) CycleMessage {
	return multiCycleError(target, fmt.Errorf("%s cycle %d/%d didn't reach target after %s", prefix, n, of, duration))
}

func chargeError(target int, err error) CycleMessage {
	return cycleMessage(target, fmt.Sprintf("Charge: %s", err), err, true)
}

func dischargeError(target int, err error) CycleMessage {
	return cycleMessage(target, fmt.Sprintf("Discharge: %s", err), err, true)
}

func multiCycleError(target int, err error) CycleMessage {
	return cycleMessage(target, fmt.Sprintf("Multi-cycle: %s", err), err, true)
}
