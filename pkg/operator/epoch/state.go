package epoch


func (e *EpochUpdator) GetEpochState() EpochState {
	return e.currentEpochState
}
