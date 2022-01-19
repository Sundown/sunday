package apotheosis

import (
	"sundown/solution/prism"

	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (state *State) CompileInlineSum(val Value) value.Value {
	typ := val.Type.(prism.VectorType).Type
	lltyp := typ.Realise()

	counter := state.Block.NewAlloca(types.I64)
	state.Block.NewStore(I64(0), counter)

	accum := state.Block.NewAlloca(lltyp)
	state.Block.NewStore(state.DefaultValue(&typ), accum)

	// Body
	// Get elem, add to accum, increment counter, conditional jump to body
	// TODO these can be simplified with the helpers in vector.go
	cond_rhs := state.Block.NewLoad(
		types.I64,
		state.Block.NewGetElementPtr(
			prism.VectorType{typ}.Realise(),
			val.Value,
			I32(0),
			vectorLenOffset))

	ll_body_actual := state.Block.NewLoad(
		types.NewPointer(lltyp),
		state.Block.NewGetElementPtr(
			prism.VectorType{typ}.Realise(),
			val.Value,
			I32(0),
			vectorBodyOffset))

	loopblock := state.CurrentFunction.NewBlock("")
	state.Block.NewBr(loopblock)
	state.Block = loopblock

	// Add to accum
	cur_counter := loopblock.NewLoad(types.I64, counter)

	// Accum <- accum + current element
	loopblock.NewStore(
		state.AgnosticAdd(
			&typ,
			loopblock.NewLoad(lltyp, accum),
			loopblock.NewLoad(
				lltyp,
				loopblock.NewGetElementPtr(
					lltyp,
					ll_body_actual,
					cur_counter))),
		accum)

	// Increment counter
	loopblock.NewStore(
		loopblock.NewAdd(loopblock.NewLoad(types.I64, counter), I64(1)),
		counter)

	cond := loopblock.NewICmp(
		enum.IPredSLT,
		loopblock.NewAdd(cur_counter, I64(1)),
		cond_rhs)

	exitblock := state.CurrentFunction.NewBlock("")
	loopblock.NewCondBr(cond, loopblock, exitblock)
	state.Block = exitblock

	return state.Block.NewLoad(lltyp, accum)
}

func (state *State) CompileInlineProduct(val Value) value.Value {
	typ := val.Type.(prism.VectorType).Type
	lltyp := typ.Realise()

	counter := state.Block.NewAlloca(types.I64)
	state.Block.NewStore(I64(0), counter)

	accum := state.Block.NewAlloca(lltyp)
	state.Block.NewStore(state.Number(&typ, 1), accum)

	// Body
	// Get elem, add to accum, increment counter, conditional jump to body

	cond_rhs := state.Block.NewLoad(
		types.I64,
		state.Block.NewGetElementPtr(
			prism.VectorType{typ}.Realise(),
			val.Value,
			I32(0),
			vectorLenOffset))

	ll_body_actual := state.Block.NewLoad(
		types.NewPointer(lltyp),
		state.Block.NewGetElementPtr(
			prism.VectorType{typ}.Realise(),
			val.Value,
			I32(0),
			vectorBodyOffset))

	loopblock := state.CurrentFunction.NewBlock("")
	state.Block.NewBr(loopblock)
	state.Block = loopblock

	// Add to accum
	cur_counter := loopblock.NewLoad(types.I64, counter)

	// Accum <- accum * current element
	loopblock.NewStore(
		state.AgnosticMult(
			&typ,
			loopblock.NewLoad(lltyp, accum),
			loopblock.NewLoad(
				lltyp,
				loopblock.NewGetElementPtr(
					lltyp,
					ll_body_actual,
					cur_counter))),
		accum)

	cond := loopblock.NewICmp(
		enum.IPredSLT,
		loopblock.NewAdd(cur_counter, I64(1)),
		cond_rhs)

	// Increment counter
	loopblock.NewStore(
		loopblock.NewAdd(loopblock.NewLoad(types.I64, counter), I64(1)),
		counter)

	exitblock := state.CurrentFunction.NewBlock("")
	loopblock.NewCondBr(cond, loopblock, exitblock)
	state.Block = exitblock

	return state.Block.NewLoad(lltyp, accum)

}
