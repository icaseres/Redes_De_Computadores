package tcp

import (

	"math"
	"time"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

)

type cubicState struct {
		stack.TCPCubicState
		numCongestionEvents int
		s *sender

}

func newCubicCC(s *sender) *cubicState {
		return &cubicState{
			TCPCubicState: stack.TCPCubicState{
						T:

s.ep.stack.Clock().NowMonotonic(),
				Beta: 0.7,
				C: 0.4,
		},
		s: s,
	}
}


func (c *cubicState) enterCongestionAvoidance() {
				if c.numCongestionEvents == 0 {

					c.K = 0
					c.T = c.s.ep.stack.Clock().NowMonotonic()
					c.WLastMax = c.WMax
					c.WMax = float64(c.s.SndCwnd)
				}
}


func (c *cubicState) updateSlowStart(packetsAcked int) int {

				newcwnd := c.s.SndCwnd + packetsAcked
				enterCA := false
				if newcwnd >= c.s.Ssthresh {
							newcwnd = c.s.Ssthresh
							c.s.SndCAAckCount = 0
							enterCA = true

				}
				packetsAcked -= newcwnd - c.s.SndCwnd
				c.s.SndCwnd = newcwnd
				if enterCA {
							c.enterCongestionAvoidance()

				}
				return packetsAcked

}

func (c *cubicState) Update(packetsAcked int) {
				if c.s.SndCwnd < c.s.Ssthresh {
						packetsAcked = c.updateSlowStart(packetsAcked)

						if packetsAcked == 0 {
									return
						}
				} else {

						c.s.rtt.Lock()
						srtt := c.s.rtt.TCPRTTState.SRTT
						c.s.rtt.Unlock()
						c.s.SndCwnd = c.getCwnd(packetsAcked,c.s.SndCwnd, srtt)
				}

}


func (c *cubicState) cubicCwnd(t float64) float64 {
					return c.C*math.Pow(t, 3.0) + c.WMax

}

func (c *cubicState) getCwnd(packetsAcked, sndCwnd int, srtt time.Duration) int {
				elapsed :=
c.s.ep.stack.Clock().NowMonotonic().Sub(c.T)
				elapsedSeconds := elapsed.Seconds()
				c.WC = c.cubicCwnd(elapsedSeconds - c.K)
				c.WEst = c.WMax*c.Beta + (3.0*((1.0-c.Beta)/(1.0+c.Beta)))*(elapsedSeconds/srtt.Seconds())


				if c.WC < c.WEst && float64(sndCwnd) < c.WEst {
						return int(c.WEst)

				}


				tEst := (elapsed + srtt).Seconds()
				wtRtt := c.cubicCwnd(tEst - c.K)


				cwnd := float64(sndCwnd)
				for i := 0; i < packetsAcked; i++ {
							cwnd += (wtRtt - cwnd) / cwnd

				}
				return int(cwnd)

}

func (c *cubicState) HandleLossDetected() {
			c.numCongestionEvents++
			c.T = c.s.ep.stack.Clock().NowMonotonic()
			c.WLastMax = c.WMax
			c.WMax = float64(c.s.SndCwnd)
			c.fastConvergence()
			c.reduceSlowStartThreshold()

}

func (c *cubicState) HandleRTOExpired() {
			c.T = c.s.ep.stack.Clock().NowMonotonic()
			c.numCongestionEvents = 0
			c.WLastMax = c.WMax
			c.WMax = float64(c.s.SndCwnd)
			c.fastConvergence()

			c.reduceSlowStartThreshold()

			c.s.SndCwnd = 1

}


func (c *cubicState) fastConvergence() {

				if c.WMax < c.WLastMax {
							c.WLastMax = c.WMax
							c.WMax = c.WMax * (1.0 + c.Beta) / 2.0
				} else {

							c.WLastMax = c.WMax
				}
				c.K = math.Cbrt(c.WMax * (1 - c.Beta) / c.C)

}

func (c *cubicState) PostRecovery() {
				c.T = c.s.ep.stack.Clock().NowMonotonic()
}

func (c *cubicState) reduceSlowStartThreshold() {
				c.s.Ssthresh = int(math.Max(float64(c.s.SndCwnd)*c.Beta, 2.0))
}