package tcp

type renoState struct {
	s *sender
}

func newRenoCC(s *sender) *renoState {
	return &renoState{s: s}
}

func (r *renoState) updateSlowStart(packetsAcked int) int {

	newcwnd := r.s.SndCwnd + packetsAcked
	if newcwnd >= r.s.Ssthresh {
		newcwnd = r.s.Ssthresh
		r.s.SndCAAckCount = 0
	}

	packetsAcked -= newcwnd - r.s.SndCwnd
	r.s.SndCwnd = newcwnd
	return packetsAcked
}

func (r *renoState) updateCongestionAvoidance(packetsAcked int) {

	r.s.SndCAAckCount += packetsAcked
	if r.s.SndCAAckCount >= r.s.SndCwnd {
		r.s.SndCwnd += r.s.SndCAAckCount / r.s.SndCwnd
		r.s.SndCAAckCount = r.s.SndCAAckCount % r.s.SndCwnd
	}
}

func (r *renoState) reduceSlowStartThreshold() {
	r.s.Ssthresh = r.s.Outstanding / 2
	if r.s.Ssthresh < 2 {
		r.s.Ssthresh = 2
	}

}

func (r *renoState) Update(packetsAcked int) {
	if r.s.SndCwnd < r.s.Ssthresh {
		packetsAcked = r.updateSlowStart(packetsAcked)
		if packetsAcked == 0 {
			return
		}
	}	
	r.updateCongestionAvoidance(packetsAcked)
}

func (r *renoState) HandleLossDetected() {
		r.reduceSlowStartThreshold()
}
func (r *renoState) HandleRTOExpired() {
		r.reduceSlowStartThreshold()
	r.s.SndCwnd = 1
}

func (r *renoState) PostRecovery() {

}