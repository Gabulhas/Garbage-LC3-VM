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

//TODO: falta operando RET
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
				imm5 := signExtend(instr&0x1F, 5)
				regs.REG[r0] = regs.REG[r1] + imm5

			} else {
				var r2 uint16 = instr & 0x7
				regs.REG[r0] = regs.REG[r1] + regs.REG[r2]
			}

			updateFlags(r0)
			break
		case OPS.AND:

			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			imm_flag := (instr >> 5) & 0x1

			if imm_flag == 1 {
				imm5 := signExtend(instr&0x1F, 5)
				regs.REG[r0] = regs.REG[r1] & imm5

			} else {
				var r2 uint16 = instr & 0x7
				regs.REG[r0] = regs.REG[r1] & regs.REG[r2]
			}

			updateFlags(r0)

			break
		case OPS.NOT:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			regs.REG[r0] = ^regs.REG[r1]
			break
		case OPS.BR:
			n := (instr >> 9) & 0x1
			z := (instr >> 10) & 0x1
			p := (instr >> 11) & 0x1
			PCoffset9 := signExtend(instr&0x1FF, 9)

			if (n&fls.FL_NEG) == 1 || (z&fls.FL_ZRO) == 1 || (p&fls.FL_POS) == 1 {
				regs.REG[regs.R_PC] = regs.REG[regs.R_PC] + PCoffset9
			}
			break
		case OPS.JMP:
			r1 := (instr >> 6) & 0x7
			regs.REG[regs.R_PC] = regs.REG[r1]
			break
		case OPS.JSR:
			regs.REG[regs.R_R7] = regs.REG[regs.R_PC]
			mode := (instr >> 11) & 0x1
			if mode == 0 {
				r1 := (instr >> 6) & 0x7
				regs.REG[regs.R_PC] = regs.REG[r1]
			} else {
				PCoffset11 := signExtend(instr&0x7FF, 11)
				regs.REG[regs.R_PC] = regs.REG[regs.R_PC] + PCoffset11
			}
			break
		case OPS.LD:
			r0 := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x1FF, 9)
			regs.REG[r0] = mem_read(regs.REG[regs.R_PC] + PCoffset9)
			updateFlags(r0)

			break
		case OPS.LDI:
			r0 := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x1FF, 9)
			regs.REG[r0] = mem_read(mem_read(regs.REG[regs.R_PC] + PCoffset9))
			updateFlags(r0)
			break
		case OPS.LDR:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			PCoffset6 := signExtend(instr&0x3F, 6)
			regs.REG[r0] = mem_read(regs.REG[r1] + PCoffset6)
			updateFlags(r0)
			break
		case OPS.LEA:
			r0 := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x3F, 6)
			regs.REG[r0] = regs.REG[regs.R_PC] + PCoffset9
			updateFlags(r0)
			break
		case OPS.ST:
			sr := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x3F, 6)
			mem_write(regs.REG[regs.R_PC]+PCoffset9, regs.REG[sr])
			break
		case OPS.STI:
			sr := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x3F, 6)
			mem_write(mem_read(regs.REG[regs.R_PC] + PCoffset9), regs.REG[sr])
			break
		case OPS.STR:
			sr := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			PCoffset6 := signExtend(instr&0x3F, 6)
			mem_write(regs.REG[r1] + PCoffset6, regs.REG[sr])

			break
		case OPS.TRAP:

			break
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

func signExtend(x uint16, bit_count int) uint16 {

	if (x>>(bit_count-1))&1 == 1 {
		return x | (0xFFFF << bit_count)
	}
	return -1

}

func updateFlags(r uint16) {
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
