package models

// AuthenticationStatus represents the current authentication state of a document
type AuthenticationStatus string

const (
	// AuthenticationStatusUnauthenticated indicates the document has been uploaded but not yet sent for authentication
	AuthenticationStatusUnauthenticated AuthenticationStatus = "unauthenticated"
	
	// AuthenticationStatusAuthenticating indicates the document has been sent for authentication and is awaiting response
	AuthenticationStatusAuthenticating AuthenticationStatus = "authenticating"
	
	// AuthenticationStatusAuthenticated indicates the document has been successfully authenticated by the external service
	AuthenticationStatusAuthenticated AuthenticationStatus = "authenticated"
)

// IsValid checks if the authentication status is one of the valid values
func (s AuthenticationStatus) IsValid() bool {
	switch s {
	case AuthenticationStatusUnauthenticated, 
		 AuthenticationStatusAuthenticating, 
		 AuthenticationStatusAuthenticated:
		return true
	default:
		return false
	}
}

// String returns the string representation of the authentication status
func (s AuthenticationStatus) String() string {
	return string(s)
}
