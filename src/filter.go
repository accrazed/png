package png

import (
	"fmt"
	"math"
)

type FilterFunc func(orig, a, b, c int) int

var (
	// None
	FilterFuncNone FilterFunc = func(orig, a, b, c int) int {
		return orig
	}
	// Sub
	FilterFuncSub FilterFunc = func(orig, a, b, c int) int {
		return orig - a
	}
	// Up
	FilterFuncUp FilterFunc = func(orig, a, b, c int) int {
		return orig - b
	}
	// Average
	FilterFuncAverage FilterFunc = func(orig, a, b, c int) int {
		return (orig - int(math.Floor((float64(a+b))/2))) % 0xFF
	}
	// Paeth
	FilterFuncPaeth FilterFunc = func(orig, a, b, c int) int {
		p := a + b - c
		pa := abs(p - a)
		pb := abs(p - b)
		pc := abs(p - c)

		if pa <= pb && pa <= pc {
			return a
		}

		if pb <= pc {
			return b
		}

		return c
	}
)

type Filterer interface {
	Filter(*Image) error // TODO: figure out signature
	Unfilter(*Image) error
}

type AdaptiveFilter struct {
	Width uint32
}

func (af *AdaptiveFilter) Filter(img *Image) error {
	if af.Width == 0 {
		return fmt.Errorf("invalid adaptive filter image width of 0")
	}

	idOut := make([]byte, len(img.data))

	var filt int
	for i, orig := range img.data {
		// check if byte is filter designator
		if i%int(af.Width+1) == 0 {
			filt = int(orig) % len(filts)
			idOut[i] = orig
			continue
		}

		upper := uint32(i) <= 4*af.Width
		left := uint32(i)%(4*af.Width) == 0

		a, b, c := 0, 0, 0
		if !left {
			a = int(img.data[i-4])
		}
		if !upper && !left {
			b = int(img.data[i-int(af.Width)*4])
		}
		if !upper {
			c = int(img.data[i-int(af.Width)*4-4])
		}

		idOut[i] = byte(filts[filt](int(orig), a, b, c))
	}

	return nil
}

func (af *AdaptiveFilter) Unfilter(id *Image) error { return nil }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var _ Filterer = (*AdaptiveFilter)(nil)
