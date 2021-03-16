package Opcodes


const (
	BR   = iota /* branch */
	ADD                /* add  */
	LD                 /* load */
	ST                 /* store */
	JSR                /* jump register */
	AND                /* bitwise and */
	LDR                /* load register */
	STR                /* store register */
	RTI                /* unused */
	NOT                /* bitwise not */
	LDI                /* load indirect */
	STI                /* store indirect */
	JMP                /* jump */
	RES                /* reserved (unused) */
	LEA                /* load effective address */
	TRAP               /* execute trap */
)
func OperandToString(i int) string {
	return []string{"BR", "ADD", "LD", "ST", "JSR", "AND", "LDR", "STR", "RTI", "NOT", "LDI", "STI", "JMP", "RES", "LEA", "TRAP",}[i]
}
