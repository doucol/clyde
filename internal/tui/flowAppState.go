package tui

type flowAppState struct {
	sumID, sumRow, flowID, flowRow int
}

func (fas *flowAppState) reset() {
	fas.sumID, fas.sumRow, fas.flowID, fas.flowRow = 0, 0, 0, 0
}

func (fas *flowAppState) setSum(id, row int) {
	fas.sumID, fas.sumRow, fas.flowID, fas.flowRow = id, row, 0, 0
}

func (fas *flowAppState) setFlow(id, row int) {
	fas.flowID, fas.flowRow = id, row
}
