package auth

import "encoding/gob"

func init() {
	gob.Register(UserIdent{})
}
