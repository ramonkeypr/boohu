// confusion idea from: https://crawl.develz.org/tavern/viewtopic.php?f=17&t=24108&sid=cb465fe78aba3b9074a32efc2a835d80#p318813

package main

type status int

const (
	StatusBerserk status = iota
	StatusSlow
	StatusExhausted
	StatusSwift
	StatusAgile
	StatusLignification
	StatusConfusion
	StatusTele
	StatusNausea
	StatusDisabledShield
	StatusCorrosion
)

func (st status) Good() bool {
	switch st {
	case StatusBerserk, StatusSwift, StatusAgile:
		return true
	default:
		return false
	}
}

func (st status) Bad() bool {
	switch st {
	case StatusSlow, StatusConfusion, StatusNausea, StatusDisabledShield:
		return true
	default:
		return false
	}
}

func (st status) String() string {
	switch st {
	case StatusBerserk:
		return "Berserk"
	case StatusSlow:
		return "Slow"
	case StatusExhausted:
		return "Exhausted"
	case StatusSwift:
		return "Swift"
	case StatusLignification:
		return "Lignified"
	case StatusAgile:
		return "Agile"
	case StatusConfusion:
		return "Confused"
	case StatusTele:
		return "Tele"
	case StatusNausea:
		return "Nausea"
	case StatusDisabledShield:
		return "-Shield"
	case StatusCorrosion:
		return "Corroded"
	default:
		// should not happen
		return "unknown"
	}
}
