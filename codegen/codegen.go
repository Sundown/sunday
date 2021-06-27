package codegen

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"sundown/sunday/parser"
	"sundown/sunday/util"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

var (
	param  = "Param"
	int_bt types.IntType
	int_t  types.Type
	real_t types.Type
	str_t  types.Type
	bool_t types.Type
)

type State struct {
	module   *ir.Module
	function *ir.Func
	Entry    *ir.Func
	block    *ir.Block
	fns      map[string]*ir.Func
}

func StartCompiler(path string, block *parser.Program) error {
	start_time := time.Now()
	state := State{}
	state.module = ir.NewModule()
	state.fns = make(map[string]*ir.Func)

	int_bt = types.IntType{TypeName: "Int", BitSize: 64}
	int_t = state.module.NewTypeDef("Int", types.I64)
	real_t = state.module.NewTypeDef("Real", types.Double)
	str_t = state.module.NewTypeDef("String", types.I8Ptr)
	bool_t = state.module.NewTypeDef("Bool", types.I1)

	state.BuiltinPuts()
	state.BuiltinDouble()

	for _, he := range block.Expression {
		if he.Expression != nil {
			state.compile(he.Expression)
		}
	}

	for _, he := range block.Expression {
		if he.Directive != nil {
			state.Direct(he.Directive)
		}
	}

	// Generate entry point
	state.function = state.module.NewFunc("main", types.I32)
	state.block = state.function.NewBlock("entry")
	state.block.NewCall(state.Entry)
	state.block.NewRet(constant.NewInt(types.I32, 0))

	ioutil.WriteFile(path, []byte(state.module.String()), 0644)

	fmt.Printf("Compiled %s in %s\n", path, time.Since(start_time).Round(1000))

	return nil
}

func (state *State) compile(expr *parser.Expression) value.Value {
	if expr.FnDecl != nil {
		if takes := GenType(expr.FnDecl.Type.Takes); takes != types.Void {
			state.function = state.module.NewFunc(
				*expr.FnDecl.Ident.Ident,
				GenType(expr.FnDecl.Type.Gives),
				ir.NewParam(param, takes))
		} else {
			state.function = state.module.NewFunc(
				*expr.FnDecl.Ident.Ident,
				GenType(expr.FnDecl.Type.Gives))
		}

		state.block = state.function.NewBlock("entry")

		// Step through and codegen each expression in the function until ";"
		for _, expr := range expr.FnDecl.Block.Expression {
			state.compile(expr)
		}

		if state.function.Sig.RetType == types.Void {
			state.block.NewRet(nil)
		}

		state.fns[*expr.FnDecl.Ident.Ident] = state.function
		// Constructing this function is over so clear state
		state.block = nil
		state.function = nil
	} else if expr.Primary != nil {
		return state.MakePrimary(expr.Primary)
	} else if expr.Application != nil {
		switch *expr.Application.Op.Ident {
		case "Return":
			if state.function.Sig.RetType == types.Void {
				state.block.NewRet(nil)
			} else {
				state.block.NewRet(state.compile(expr.Application.Atoms))
			}

		case "Head":
			vec, vec_type := state.compile_vector(expr.Application.Atoms.Primary.Vec)
			return state.block.NewLoad(
				types.I32,
				state.block.NewGetElementPtr(
					vec_type,
					vec,
					constant.NewInt(types.I32, 0),
					constant.NewInt(types.I32, 0)))
		default:
			fn, err := state.fns[*expr.Application.Op.Ident]
			if !err {
				util.Error("Function not found")
				os.Exit(1)
			}
			return state.block.NewCall(
				fn,
				state.compile(expr.Application.Atoms))
		}
	}

	return nil
}

func (state *State) compile_vector(vector []*parser.Expression) (value.Value, *types.VectorType) {
	elm_type := GenPrimaryType(vector[0].Primary)
	fmt.Println(elm_type)
	vec_type := &types.VectorType{
		TypeName: "",
		Scalable: true,
		Len:      uint64(len(vector)),
		ElemType: elm_type}
	vec := state.block.NewAlloca(vec_type)

	for i, elm := range vector {
		ptr := state.block.NewGetElementPtr(
			vec_type,
			vec,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(i)))
		ptr.InBounds = true

		state.block.NewStore(state.compile(elm), ptr)
	}

	return vec, vec_type
}

func GenType(t *parser.TypeName) types.Type {
	switch *t.Type {
	case "Int":
		return int_t
	case "Real":
		return real_t
	case "Bool":
		return types.I1
	case "Void":
		return types.Void
	case "Str":
		return str_t
	default:
		return types.Void
	}
}

func GenPrimaryType(p *parser.Primary) types.Type {
	if p != nil {
		switch {
		case p.Int != nil:
			return int_t
		case p.Real != nil:
			return real_t
		case p.Bool != nil:
			return bool_t
		case p.String != nil:
			return str_t
		}
	}

	return nil
}
