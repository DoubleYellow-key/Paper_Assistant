package errors

const (
	CodeOK = 0

	CodeBadRequest    = 40001
	CodeNotFound      = 40401
	CodeStateConflict = 40901

	CodeUnauthorized = 40101
	CodeTokenExpired = 40102
	CodeForbidden    = 40301

	CodeRateLimited = 42901

	CodeParseFailed = 50011
	CodeModelFailed = 50021
	CodeInternal    = 50000
)
