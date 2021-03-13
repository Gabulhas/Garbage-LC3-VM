package main

import (
	regs "golang_vm/Registers"
	"math"
	"os"
)

var memory [math.MaxInt16]uint16

func main() {


	if len(os.Args) < 2 {
		/* show usage string */
		fmt.Printf("lc3 [image-file1] ...\n")
		exit(2)
	    }



	const (
		PC_START = 0x3000
	)
	regs.REG[regs.R_PC] = PC_START

	running := 1

	for running == 1 {
		instr := mem_read(regs.REG[R_PC]++)
		op := instr >> 12

		switch op {
		    case OP_ADD:
			{ADD, 6}
			break
		    case OP_AND:
			{AND, 7}
			break
		    case OP_NOT:
			{NOT, 7}
			break
		    case OP_BR:
			{BR, 7}
			break
		    case OP_JMP:
			{JMP, 7}
			break
		    case OP_JSR:
			{JSR, 7}
			break
		    case OP_LD:
			{LD, 7}
			break
		    case OP_LDI:
			{LDI, 6}
			break
		    case OP_LDR:
			{LDR, 7}
			break
		    case OP_LEA:
			{LEA, 7}
			break
		    case OP_ST:
			{ST, 7}
			break
		    case OP_STI:
			{STI, 7}
			break
		    case OP_STR:
			{STR, 7}
			break
		    case OP_TRAP:
			{TRAP, 8}
			break
		    case OP_RES:
		    case OP_RTI:
		    default:
			{BAD OPCODE, 7}
			break;

		}

	}

}
