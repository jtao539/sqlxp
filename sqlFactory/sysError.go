package sqlFactory

import "errors"

var UpdateError = errors.New("this update affected rows is 0")

var InsertError = errors.New("this insert affected rows is 0")

var DeleteError = errors.New("this delete affected rows is 0")
