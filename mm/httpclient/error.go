package httpclient

var _ ReplyErr = (*replyErr)(nil)

// ReplyErr 错误响应，当resp.StatusCode != http.StatusOK时用来包装返回的信息。
type ReplyErr interface {
	error
	t()
}

type replyErr struct {
	err  error
	code int
}

func (r *replyErr) Error() string {
	return r.err.Error()
}

func (r *replyErr) t() {}

func newReplyErr(code int, err error) ReplyErr {
	return &replyErr{
		code: code,
		err:  err,
	}
}

// ToReplyErr 尝试将err转换为ReplyErr
func ToReplyErr(err error) (statusCode int, _ error, _ bool) {
	if err == nil {
		return -1, nil, false
	}

	if e, ok := err.(*replyErr); ok {
		return e.code, e.err, true
	}

	return -1, err, false
}
