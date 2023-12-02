package statement

func TernaryF[T any](cond func() bool, ifTrue, ifFalse T) T {

	return Ternary(cond(), ifTrue, ifFalse)

}

func Ternary[T any](cond bool, ifTrue, ifFalse T) T {

	if cond {
		return ifTrue
	} else {
		return ifFalse
	}

}
