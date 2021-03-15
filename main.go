package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	fls "golang_vm/ConditionFlags"
	mr "golang_vm/MemoryMappedRegisters"
	OPS "golang_vm/Opcodes"
	regs "golang_vm/Registers"
	trp "golang_vm/Traps"
	"io/ioutil"
	"log"
	"math"
	"os"
)

var memory [math.MaxUint16]uint16
var inputReader *bufio.Reader

//TODO: falta operando RET
func main() {

	if len(os.Args) < 2 {
		log.Fatal("lc3 [image-file1] ...\n")

	}

	for j := 1; j < len(os.Args); j++ {
		readImage(os.Args[j])
	}

	const PC_START = 0x3000
	regs.REG[regs.PC] = PC_START

	inputReader = bufio.NewReader(os.Stdin)
	running := true

	for running {
		instr := memRead(regs.REG[regs.PC])
		regs.REG[regs.PC]++
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
				r2 := instr & 0x7
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
				regs.REG[regs.PC] = regs.REG[regs.PC] + PCoffset9
			}
			break
		case OPS.JMP:
			r1 := (instr >> 6) & 0x7
			regs.REG[regs.PC] = regs.REG[r1]
			break
		case OPS.JSR:
			regs.REG[regs.R7] = regs.REG[regs.PC]
			mode := (instr >> 11) & 0x1
			if mode == 0 {
				r1 := (instr >> 6) & 0x7
				regs.REG[regs.PC] = regs.REG[r1]
			} else {
				PCoffset11 := signExtend(instr&0x7FF, 11)
				regs.REG[regs.PC] = regs.REG[regs.PC] + PCoffset11
			}
			break
		case OPS.LD:
			r0 := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x1FF, 9)
			regs.REG[r0] = memRead(regs.REG[regs.PC] + PCoffset9)
			updateFlags(r0)

			break
		case OPS.LDI:
			r0 := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x1FF, 9)
			regs.REG[r0] = memRead(memRead(regs.REG[regs.PC] + PCoffset9))
			updateFlags(r0)
			break
		case OPS.LDR:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			PCoffset6 := signExtend(instr&0x3F, 6)
			regs.REG[r0] = memRead(regs.REG[r1] + PCoffset6)
			updateFlags(r0)
			break
		case OPS.LEA:
			r0 := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x3F, 6)
			regs.REG[r0] = regs.REG[regs.PC] + PCoffset9
			updateFlags(r0)
			break
		case OPS.ST:
			sr := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x3F, 6)
			memWrite(regs.REG[regs.PC]+PCoffset9, regs.REG[sr])
			break
		case OPS.STI:
			sr := (instr >> 9) & 0x7
			PCoffset9 := signExtend(instr&0x3F, 6)
			memWrite(memRead(regs.REG[regs.PC]+PCoffset9), regs.REG[sr])
			break
		case OPS.STR:
			sr := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			PCoffset6 := signExtend(instr&0x3F, 6)
			memWrite(regs.REG[r1]+PCoffset6, regs.REG[sr])

			break
		case OPS.TRAP:
			trap := instr & 0xFF
			switch trap {
			case trp.GETC:
				charByte, err := inputReader.ReadByte()
				if err != nil {
					log.Fatal(err)
				}
				regs.REG[regs.R0] = uint16(charByte)
				break
			case trp.OUT:
				fmt.Printf("%c", regs.REG[regs.R0])
				break
			case trp.PUTS:

				c := regs.REG[regs.R0]
				var rdchar uint16
				for {
					rdchar = memRead(c)
					if rdchar == 0 {
						break
					}
					fmt.Printf("%c", rdchar)
					c++
				}

				break
			case trp.IN:
				fmt.Printf("Enter a character: ")
				charByte, err := inputReader.ReadByte()
				if err != nil {
					log.Fatal(err)
				}
				regs.REG[regs.R0] = uint16(charByte)
				break
			case trp.PUTSP:

				c := regs.REG[regs.R0]
				var rdchar uint16
				for {
					rdchar = memRead(c)

					char1 := rdchar & 0xff
					if char1 == 0 {
						break
					}
					fmt.Printf("%c", char1)

					char2 := rdchar >> 8

					if char2 == 0 {
						break
					}
					fmt.Printf("%c", char2)
					c++
				}

				break
			case trp.HALT:
				fmt.Printf("HALT")
				inputReader.ReadString('\n')
				running = false
				break
			}

			break
		case OPS.RES:
			break
		case OPS.RTI:
			break
		default:
			running = false
			log.Fatalf("\nBAD OPCODE:%d", op)
			break

		}

	}

}

func signExtend(x uint16, bit_count int) uint16 {

	if (x>>(bit_count-1))&1 == 1 {
		return x | (0xFFFF << bit_count)
	}
	return 0

}

func updateFlags(r uint16) {
	if regs.REG[r] == 0 {
		regs.REG[regs.COND] = fls.FL_ZRO

		/* a 1 in the left-most bit indicates negative */
	} else if regs.REG[r]>>15 == 1 {
		regs.REG[regs.COND] = fls.FL_NEG

	} else {
		regs.REG[regs.COND] = fls.FL_POS
	}
}

func readImage(filename string) {
	fileI, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	readImageFile(fileI)

}

func readImageFile(fileContent []byte) {
	origin := binary.BigEndian.Uint16(fileContent[:2])
	for i := 2; i < len(fileContent); i += 2 {
		memWrite(origin, binary.BigEndian.Uint16(fileContent[i:i+2]))
		origin++
	}
}

func memWrite(address, val uint16) {
	memory[address] = val
}

func memRead(address uint16) uint16 {
	if address == mr.KBSR {
		if checkKey() {
			memory[mr.KBSR] = 0x1 << 15
			charByte, err := inputReader.ReadByte()
			if err != nil {
				log.Fatal(err)
			}
			memory[mr.KBDR] = uint16(charByte)
		}
	} else {
		memory[mr.KBSR] = 0
	}
	return memory[address]
}

func checkKey() bool {
	//TODO: implement
	return true
}
