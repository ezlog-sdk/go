package ezlog

const SDK_VERSION = "0.0.2"

const DEFAULT_ENDPOINT = "https://collect.ezlog.cloud/v1"

type ClientOptions struct {
	ServiceKey string
	Endpoint   string
}

var clientOptions *ClientOptions

func Configure(options ClientOptions) error {
	if options.Endpoint == "" {
		options.Endpoint = DEFAULT_ENDPOINT
	}

	clientOptions = &options

	return nil
}
