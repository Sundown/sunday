package prism

import "github.com/llir/llvm/ir/types"

func EqType(a, b Type) bool {
	if a.Kind() != b.Kind() {
		return false
	}

	switch a.Kind() {
	case TypeKindAtomic:
		return a.(AtomicType).ID == b.(AtomicType).ID
	case TypeKindVector:
		return EqType(a.(VectorType).ElementType, b.(VectorType).ElementType)
	case TypeKindStruct:
		// TODO
	}

	return false
}

func (f Function) Type() Type {
	return f.Returns
}

func (m Monadic) Type() Type {
	return m.Operator.Type()
}

func (d Dyadic) Type() Type {
	return d.Operator.Type()
}

func (a Application) Type() Type {
	return a.Operator.Type()
}

func (i Int) Type() Type {
	return IntType
}

func (r Real) Type() Type {
	return RealType
}

func (c Char) Type() Type {
	return CharType
}

func (b Bool) Type() Type {
	return BoolType
}

func (s String) Type() Type {
	return StringType
}

var (
	IntType = AtomicType{
		ID:           TypeInt,
		WidthInBytes: 8,
		Name:         ParseIdent("Int"),
		Actual:       types.I64,
	}
	RealType = AtomicType{
		ID:           TypeReal,
		WidthInBytes: 8,
		Name:         ParseIdent("Real"),
		Actual:       types.Double,
	}
	CharType = AtomicType{
		ID:           TypeChar,
		WidthInBytes: 1,
		Name:         ParseIdent("Char"),
		Actual:       types.I8,
	}
	StringType = AtomicType{
		ID:           TypeString,
		WidthInBytes: 12, // TODO
		Name:         ParseIdent("String"),
		Actual:       types.I8Ptr,
	}
	BoolType = AtomicType{
		ID:           TypeBool,
		WidthInBytes: 1,
		Name:         ParseIdent("Bool"),
		Actual:       types.I1,
	}
	VoidType = AtomicType{
		ID:           TypeVoid,
		WidthInBytes: 0,
		Name:         ParseIdent("Void"),
		Actual:       types.Void,
	}
)
