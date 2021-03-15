package Registers

type Register int16

const (
	R0 Register = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	PC /* program counter */
	COND
	COUNT
)

var REG [COUNT]uint16
