package auth

import "github.com/alexedwards/argon2id"

var Argon2IdParams = &argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  4,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}
