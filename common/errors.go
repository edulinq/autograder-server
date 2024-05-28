package common

// Errors that don't present security issues when returned to students.
type SecureError struct {
    Message string
}

func (this *SecureError) Error() string {
    return this.Message;
}
