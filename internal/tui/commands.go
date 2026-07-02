package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/kube"
)

const refreshInterval = 2 * time.Second

type tickMsg time.Time

type flowSumTotalsMsg []*flowdata.FlowSum

type flowSumRatesMsg []*flowdata.FlowSum

type flowsBySumMsg struct {
	sumID int
	flows []*flowdata.FlowData
}

type clusterReadyMsg struct {
	info kube.ClusterNetworkingInfo
}

// installDoneMsg reports the outcome of an attempt to enable Goldmane/Whisker.
// A nil err means Whisker is now available.
type installDoneMsg struct {
	err error
}

// whiskerInstallTimeout bounds how long we wait for the operator to reconcile
// the new Goldmane/Whisker resources and bring the whisker-backend pod up.
const whiskerInstallTimeout = 3 * time.Minute

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type dataProvider interface {
	GetFlowSumTotals() []*flowdata.FlowSum
	GetFlowSumRates() []*flowdata.FlowSum
	GetFlowsBySumID(sumID int) []*flowdata.FlowData
}

func fetchSumTotals(fc dataProvider) tea.Cmd {
	return func() tea.Msg {
		return flowSumTotalsMsg(fc.GetFlowSumTotals())
	}
}

func fetchSumRates(fc dataProvider) tea.Cmd {
	return func() tea.Msg {
		return flowSumRatesMsg(fc.GetFlowSumRates())
	}
}

func fetchFlowsBySum(fc dataProvider, sumID int) tea.Cmd {
	return func() tea.Msg {
		return flowsBySumMsg{sumID: sumID, flows: fc.GetFlowsBySumID(sumID)}
	}
}

// checkClusterReadyCmd inspects the selected cluster and returns a
// clusterReadyMsg. The caller decides what to do if Goldmane isn't available.
func checkClusterReadyCmd(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		cc := cmdctx.CmdCtxFromContext(ctx)
		info := kube.GetClusterNetworkingInfo(ctx, cc.Clientset(), cc.ClientDyn(), cc.GetK8sConfig())
		return clusterReadyMsg{info: info}
	}
}

// installWhiskerCmd creates the Goldmane/Whisker operator resources on the
// selected cluster and waits for Whisker to come up, returning an
// installDoneMsg with the outcome.
func installWhiskerCmd(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		cc := cmdctx.CmdCtxFromContext(ctx)
		if err := kube.InstallGoldmaneWhisker(ctx, cc.ClientDyn()); err != nil {
			return installDoneMsg{err: err}
		}
		if err := kube.WaitForWhiskerAvailable(ctx, cc.Clientset(), "", whiskerInstallTimeout); err != nil {
			return installDoneMsg{err: err}
		}
		return installDoneMsg{}
	}
}
