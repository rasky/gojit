package amd64

import (
	"fmt"
	"unsafe"
)

func (a *Assembler) Inc(o Operand) {
	o.Rex(a, Register{})
	a.byte(0xff)
	o.ModRM(a, Register{})
}

func (a *Assembler) Dec(o Operand) {
	o.Rex(a, Register{})
	a.byte(0xff)
	o.ModRM(a, Register{1, 0})
}

func (a *Assembler) Incb(o Operand) {
	o.Rex(a, Register{})
	a.byte(0xfe)
	o.ModRM(a, Register{})
}

func (a *Assembler) Decb(o Operand) {
	o.Rex(a, Register{})
	a.byte(0xfe)
	o.ModRM(a, Register{1, 0})
}

func (asm *Assembler) arithmeticImmReg(insn *Instruction, src Imm, dst Register) {
	if insn.imm_r != 0 {
		asm.rex(false, false, false, dst.Val > 7)
		asm.byte(insn.imm_r | (dst.Val & 7))
	} else {
		asm.rex(dst.Bits == 64, false, dst.Val > 7, false)
		asm.byte(insn.imm_rm.op)
		asm.modrm(MOD_REG, insn.imm_rm.sub, dst.Val&7)
	}
}

func (asm *Assembler) arithmeticRegReg(insn *Instruction, src Register, dst Register) {
	if insn.r_rm != 0 {
		dst.Rex(asm, src)
		asm.byte(insn.r_rm)
		dst.ModRM(asm, src)
	} else {
		src.Rex(asm, dst)
		asm.byte(insn.rm_r)
		src.ModRM(asm, dst)
	}
}

func (asm *Assembler) Arithmetic(insn *Instruction, src, dst Operand) {
	switch s := src.(type) {
	case Imm:
		if dr, ok := dst.(Register); ok {
			asm.arithmeticImmReg(insn, s, dr)
		} else {
			dst.Rex(asm, Register{insn.imm_rm.sub, 0})
			asm.byte(insn.imm_rm.op)
			dst.ModRM(asm, Register{insn.imm_rm.sub, 0})
		}
		asm.int32(uint32(s.Val))
		return
	case Register:
		if dr, ok := dst.(Register); ok {
			asm.arithmeticRegReg(insn, s, dr)
		} else {
			dst.Rex(asm, s)
			asm.byte(insn.r_rm)
			dst.ModRM(asm, s)
		}
		return
	}
	// if the LHS is neither an immediate nor a register, the rhs
	// must be a register
	dr, ok := dst.(Register)
	if !ok {
		panic(fmt.Sprintf("arithmetic: %#v/%#v not supported!", src, dst))
	}

	src.Rex(asm, dr)
	asm.byte(insn.rm_r)
	src.ModRM(asm, dr)
}

func (a *Assembler) Add(src, dst Operand) {
	a.Arithmetic(InstAdd, src, dst)
}

func (a *Assembler) And(src, dst Operand) {
	a.Arithmetic(InstAnd, src, dst)
}

func (a *Assembler) Cmp(src, dst Operand) {
	a.Arithmetic(InstCmp, src, dst)
}

func (a *Assembler) Mov(src, dst Operand) {
	a.Arithmetic(InstMov, src, dst)
}

func (a *Assembler) Or(src, dst Operand) {
	a.Arithmetic(InstOr, src, dst)
}

func (a *Assembler) Lea(src, dst Operand) {
	a.Arithmetic(InstLea, src, dst)
}

func (a *Assembler) Sub(src, dst Operand) {
	a.Arithmetic(InstSub, src, dst)
}

func (a *Assembler) Test(src, dst Operand) {
	a.Arithmetic(InstTest, src, dst)
}

func (a *Assembler) Xor(src, dst Operand) {
	a.Arithmetic(InstXor, src, dst)
}

func (a *Assembler) Ret() {
	a.byte(0xc3)
}

func (a *Assembler) Call(dst Operand) {
	if _, ok := dst.(Imm); ok {
		panic("can't call(Imm); use CallRel instead.")
	} else {
		a.byte(0xff)
		dst.ModRM(a, Register{0x2, 64})
	}
}

func (a *Assembler) CallRel(dst uintptr) {
	a.byte(0xe8)
	a.rel32(dst)
}

// Clobbers RDX
func (a *Assembler) CallFunc(f func()) {
	addr := *(*uintptr)(unsafe.Pointer(&f))
	a.Lea(PCRel{addr}, Rdx)
	a.Call(Indirect{Rdx, 0, 64})
}