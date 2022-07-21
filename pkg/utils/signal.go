package utils

type Signal string

const (
	SignalInt  Signal = "int"
	SignalTerm Signal = "term"
	SignalUsr2 Signal = "usr2"
)
