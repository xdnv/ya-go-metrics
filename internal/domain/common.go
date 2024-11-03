// common configuration parts
package domain

// Transport mode values
const TRANSPORT_HTTP = "http"
const TRANSPORT_GRPC = "grpc"

// Endpoint default
const ENDPOINT = "localhost:8080"

// Default loglevel
const LOGLEVEL = "info"

//structure to be filled by common function to react in http or grpc handler
type HandlerStatus struct {
	Message    string
	Err        error
	HTTPStatus int
}
