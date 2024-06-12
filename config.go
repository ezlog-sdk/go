package ezlog

const SDK_VERSION = "0.0.1"

const DEFAULT_ENDPOINT = "https//collect.ezlog.cloud"

type ClientOptions struct {
	ServiceKey string
	Endpoint   string
}

var clientOptions *ClientOptions

func Config(options ClientOptions) error {
	if options.Endpoint == "" {
		options.Endpoint = DEFAULT_ENDPOINT
	}

	clientOptions = &options

	return nil
}
