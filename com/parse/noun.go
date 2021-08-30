package parse

import (
	"sundown/solution/lex"
	"sundown/solution/util"
)

// Tries to find noun in order (defined_namespace or foundation) then package
// order may change in future such that foundation is last
func (state *State) GetNoun(key *lex.Ident) *Atom {
	k := IRIdent(key)
	noun := state.NounDefs[k.AsKey()]

	if noun == nil {
		noun = state.NounDefs[IdentKey{
			Namespace: *state.PackageIdent,
			Ident:     *key.Ident,
		}]

		if noun == nil {
			fn := state.GetFunction(k)
			if fn != nil {
				noun = &Atom{TypeOf: fn.Gives, Function: fn}

			} else {
				util.Error("Identifier \"" + util.Yellow(k.String()) + "\" is not defined in current scope or Foundation.\n" + key.Pos.String()).Exit()
			}
		}
	}

	return noun
}

func (state *State) AnalyseNounDecl(noun *lex.NounDecl) {
	if IsReserved(*noun.Ident) {
		util.Error("Identifier \"" + util.Yellow(*noun.Ident) + "\" is reserved by the compiler.\n" + noun.Pos.String()).Exit()
	}

	var temp *Atom

	if noun.Value.Noun != nil {
		temp = state.GetNoun(noun.Value.Noun)
	} else if noun.Value.Param != nil {
		// ... why
		util.Error("Cannot use \"" + util.Yellow("@") + "\" (parameter figurative) as R-value in definition.\n" + noun.Pos.String()).Exit()
	} else {
		temp = state.AnalyseAtom(noun.Value)
	}

	key := IdentKey{Namespace: *state.PackageIdent, Ident: *noun.Ident}
	if state.NounDefs[key] == nil {
		state.NounDefs[key] = temp
	} else {
		util.Error("Noun \"" + util.Yellow(*noun.Ident) + "\" is already defined as " + util.Yellow(state.NounDefs[key].String()) + ".\n" + noun.Pos.String()).Exit()
	}
}
