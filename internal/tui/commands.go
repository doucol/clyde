package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
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
	info util.ClusterNetworkingInfo
}

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
		info := util.GetClusterNetworkingInfo(ctx, cc.Clientset(), cc.ClientDyn(), cc.GetK8sConfig())
		return clusterReadyMsg{info: info}
	}
}
