package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/nsf/termbox-go"
	fls "golang_vm/ConditionFlags"
	mr "golang_vm/MemoryMappedRegisters"
	OPS "golang_vm/Opcodes"
	regs "golang_vm/Registers"
	trp "golang_vm/Traps"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

var memory [math.MaxUint16]uint16
var inputReader *bufio.Reader
var keyBuffer []rune
var logs []string

//TODO: falta operando RET
func main() {

	if len(os.Args) < 2 {
		log.Fatal("lc3 [image-file1] ...\n")

	}

	for j := 1; j < len(os.Args); j++ {
		readImage(os.Args[j])
	}
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	defer termbox.Close()

	keyBuffer = []rune{}

	listenKeyPresses()

	const PC_START = 0x3000
	regs.REG[regs.PC] = PC_START

	inputReader = bufio.NewReader(os.Stdin)
	running := true

	for running {
		instr := memRead(regs.REG[regs.PC])
		regs.REG[regs.PC]++
		op := instr >> 12

		logs = append(logs, OPS.OperandToString(int(op)))
		switch op {

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
				r2 := instr & 0x7
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

			PCoffset9 := signExtend(instr&0x1FF, 9)
			condFlag := (instr >> 9) & 0x7
			logs = append(logs, fmt.Sprintf(":%d %d", PCoffset9, condFlag))

			if (condFlag & regs.REG[regs.COND]) == 1 {
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
			logs = append(logs, fmt.Sprintf(":%d", trap))
			switch trap {
			case trp.GETC:
				regs.REG[regs.R0] = getChar()
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
				regs.REG[regs.R0] = getChar()
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
				getChar()
				termbox.Flush()
				termbox.Close()
				running = false
				fmt.Println(logs)
				os.Exit(0)
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
			memory[mr.KBSR] = 1 << 15
			memory[mr.KBDR] = uint16(keyBuffer[len(keyBuffer)-1])
		}
	} else {
		memory[mr.KBSR] = 0
	}
	return memory[address]
}

func getChar() uint16 {
	for len(keyBuffer) == 0 {
		time.Sleep(time.Microsecond)
	}
	keyPress := keyBuffer[0]
	fmt.Printf("KEY PRESSED: %c", keyPress)
	keyBuffer = keyBuffer[1:]
	return uint16(keyPress)
}

func listenKeyPresses() {
	go func() {
		for {
			if event := termbox.PollEvent(); event.Type == termbox.EventKey {
				keyBuffer = append(keyBuffer, event.Ch)
				if event.Key == termbox.KeyCtrlC {
					os.Exit(0)
				}
			}
		}
	}()

}

func DebugRegister() {
	for i, val := range regs.REG {
		fmt.Printf("R%d - %d\n", i, val)
	}
	fmt.Println("---------")
}

func checkKey() bool {
	if len(keyBuffer) > 0 {
		return true
	}

	return false
}
