package Registers

type Register int16

const (
	R_R0 Register = iota
	R_R1
	R_R2
	R_R3
	R_R4
	R_R5
	R_R6
	R_R7
	R_PC /* program counter */
	R_COND
	R_COUNT
)

var REG [R_COUNT]int16
