package thread

import "github.com/soypat/sdf"

// Unified thread standard.
// Example: UNC 1/4 with external threading would be
//  UTS{D:1.0/4.0, TPI:20, Ext: true}
type UTS struct {
	D   float64
	TPI float64
	// External or internal thread.
	Ext bool
}

var _ Threader = UTS{} // Interface implementation.

func (uts UTS) Parameters() Parameters {
	p := basic{D: uts.D, P: 1.0 / uts.TPI}.Parameters()
	// TODO(soypat) add imperial hex flat-to-flat. See NPT for what that could look like.
	return p
}

func (uts UTS) Thread() (sdf.SDF2, error) {
	return ISO{D: uts.D, P: 1.0 / uts.TPI, Ext: uts.Ext}.Thread()
}
