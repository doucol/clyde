package tui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
)

func intos(v int64) string {
	return strconv.FormatInt(v, 10)
}

func tf(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func policyHitsToString(hits []*flowdata.PolicyHit) string {
	s := ""
	for _, ph := range hits {
		s += policyHitToString(ph) + "\n"
	}
	return s
}

func policyHitToString(ph *flowdata.PolicyHit) string {
	s := fmt.Sprintf("Kind: %s\nName: %s\nNamespace: %s\nTier: %s\nAction: %s\nPolicyIndex: %d\nRuleIndex: %d\n",
		ph.Kind, ph.Name, ph.Namespace, ph.Tier, ph.Action, ph.PolicyIndex, ph.RuleIndex)
	if ph.Trigger != nil {
		s += "\nTriggers:\n" + policyHitTriggerToString(ph.Trigger)
	}
	return s
}

func policyHitTriggerToString(ph *flowdata.PolicyHit) string {
	s := fmt.Sprintf("\tKind: %s\n\tName: %s\n\tNamespace: %s\n\tTier: %s\n\tAction: %s\n\tPolicyIndex: %d\n\tRuleIndex: %d\n",
		ph.Kind, ph.Name, ph.Namespace, ph.Tier, ph.Action, ph.PolicyIndex, ph.RuleIndex)
	if ph.Trigger != nil {
		s += "\n\n" + policyHitTriggerToString(ph.Trigger)
	}
	return s
}
