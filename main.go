package main

import (
	fls "golang_vm/ConditionFlags"
	OPS "golang_vm/Opcodes"
	regs "golang_vm/Registers"
	"log"
	"math"
	"os"
)

var memory [math.MaxInt16]uint16

func main() {

	if len(os.Args) < 2 {
		log.Fatal("lc3 [image-file1] ...\n")

	}

	for j := 1; j < len(os.Args); j++ {

		if !read_image(os.Args[j]) {
			log.Fatalf("failed to load image: %s\n", os.Args[j])
		}
	}

	const (
		PC_START = 0x3000
	)
	regs.REG[regs.R_PC] = PC_START

	running := 1

	for running == 1 {
		instr := mem_read(regs.REG[regs.R_PC] + 1)
		op := instr >> 12

		switch op {
		//ADD - takes two values and stores them in one register
		//if imm_flag (immediate mode) it takes sums with the value
		//else it takes a register as an argument and adds it

		//Like, imm mode: ADD R2 R0 14
		//Like, normal mode: ADD R2 R0 R1

		case OPS.ADD:

			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			imm_flag := (instr >> 5) & 0x1

			if imm_flag == 1 {
				imm5 := sign_extend(instr&0x1F, 5)
				regs.REG[r0] = regs.REG[r1] + imm5

			} else {
				var r2 uint16 = instr & 0x7
				regs.REG[r0] = regs.REG[r1] + regs.REG[r2]
			}

			update_flags(r0)
			break
		case OPS.AND:
			break
		case OPS.NOT:
			break
		case OPS.BR:
			break
		case OPS.JMP:
			break
		case OPS.JSR:
			break
		case OPS.LD:
			break
		case OPS.LDI:
			r0 := (instr >> 9) & 0x7
			pc_offset := sign_extend(instr&0x1FF, 9)

			regs.REG[r0] = mem_read(mem_read(regs.REG[regs.R_PC] + pc_offset))
			update_flags(r0)
			break
		case OPS.LDR:
			break
		case OPS.LEA:
			break
		case OPS.ST:
			break
		case OPS.STI:
		case OPS.STR:
		case OPS.TRAP:
		case OPS.RES:
			break
		case OPS.RTI:
			break
		default:
			log.Fatal("BAD OPCODE:" + op)
			break

		}

	}

}

func sign_extend(x uint16, bit_count int) uint16 {

	if (x>>(bit_count-1))&1 == 1 {
		return x | (0xFFFF << bit_count)
	}
	return -1

}

func update_flags(r uint16) {
	if regs.REG[r] == 0 {
		regs.REG[regs.R_COND] = fls.FL_ZRO

		/* a 1 in the left-most bit indicates negative */
	} else if regs.REG[r]>>15 == 1 {
		regs.REG[regs.R_COND] = fls.FL_NEG

	} else {
		regs.REG[regs.R_COND] = fls.FL_POS
	}
}

func read_image(filename string) bool {

}
