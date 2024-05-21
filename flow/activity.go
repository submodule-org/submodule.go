package flow

type Executable[T any, R any] interface {
	Execute(FlowContext, T) (R, error)
}

func ExecuteActivity[T any, R any](ctx FlowContext, a Activity[T, R], param T) (r R, e error) {
	v, e := ctx.GetFlow().resolveActivity(a)

	if e != nil {
		return r, e
	}

	return v.(R), nil
}

func ExecuteExecutable[T any, R any](ctx FlowContext, e Executable[T, R], param T) (R, error) {
	return e.Execute(ctx, param)
}
