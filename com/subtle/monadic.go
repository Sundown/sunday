package subtle

import (
	"sundown/solution/palisade"
	"sundown/solution/prism"
)

func (env Environment) AnalyseMonadic(m *palisade.Monadic) (app prism.MApplication) {
	if m.Verb != nil {
		return env.AnalyseStandardMonadic(m)
	} else if m.Subexpr != nil {
		return env.AnalysePartialMonadic(m)
	} else {
		panic("unreachable")
	}
}

func (env Environment) AnalysePartialMonadic(m *palisade.Monadic) (app prism.MApplication) {
	app = prism.MApplication{
		Operator: env.AnalyseExpression(m.Subexpr).(prism.MonadicFunction),
		Operand:  env.AnalyseExpression(m.Expression),
	}

	tmp := app.Operand.Type()
	resolved_right, err := prism.Delegate(&app.Operator.OmegaType, &tmp)
	if err != nil {
		prism.Panic(*err)
	}

	if prism.PredicateGenericType(app.Operator.Returns) {
		app.Operator.Returns = prism.IntegrateGenericType(resolved_right, app.Operator.Returns)
	}

	return app
}

func (env Environment) AnalyseStandardMonadic(m *palisade.Monadic) (app prism.MApplication) {
	fn := env.FetchMVerb(m.Verb)

	expr := env.AnalyseExpression(m.Expression)

	tmp := expr.Type()
	resolved_right, err := prism.Delegate(&fn.OmegaType, &tmp)

	if !prism.PureMatch(tmp, fn.OmegaType) {
		if !prism.QueryCast(tmp, fn.OmegaType) {
			tmp := tmp
			_, err := prism.Delegate(&fn.OmegaType, &tmp)
			if err != nil {
				prism.Panic(*err)
			}
		} else {
			expr = prism.DelegateCast(expr, fn.OmegaType)
		}
	}

	if prism.PredicateGenericType(fn.Returns) {
		fn.Returns = prism.IntegrateGenericType(fn.OmegaType, fn.Returns)
	}

	if err != nil {
		prism.Panic(*err)
	}

	if prism.PredicateGenericType(fn.Returns) {
		fn.Returns = prism.IntegrateGenericType(resolved_right, fn.Returns)
	}

	if fn.Name.Package == "_" && fn.Name.Name == "Return" {
		if !prism.PrimativeTypeEq(env.CurrentFunctionIR.Type(), fn.Returns) {
			if !prism.PredicateGenericType(env.CurrentFunctionIR.Type()) {
				panic("Return recieves " + fn.Returns.String() + " which does not match determined-function's type " + env.CurrentFunctionIR.Type().String())
			} else {
				panic("Not implemented, pain")
			}
		}
	}

	return prism.MApplication{
		Operator: fn,
		Operand:  expr,
	}
}
