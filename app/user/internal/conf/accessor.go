package conf

import "os"

// Implement ConfigAccessor interface for Bootstrap

func (b *Bootstrap) GetServiceName() string {
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		return name
	}
	if b != nil && b.Service != nil {
		return b.Service.Name
	}
	return "user"
}

func (b *Bootstrap) GetServiceVersion() string {
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}
	if b != nil && b.Service != nil {
		return b.Service.Version
	}
	return "v1.0.0"
}

func (b *Bootstrap) GetConsulAddress() string {
	if b != nil && b.Registry != nil && b.Registry.Consul != nil {
		return b.Registry.Consul.Address
	}
	return ""
}

func (b *Bootstrap) GetConsulScheme() string {
	if b != nil && b.Registry != nil && b.Registry.Consul != nil {
		return b.Registry.Consul.Scheme
	}
	return ""
}

func (b *Bootstrap) GetOTLPEndpoint() string {
	if b != nil && b.Tracing != nil && b.Tracing.Otlp != nil {
		return b.Tracing.Otlp.Endpoint
	}
	return ""
}

func (b *Bootstrap) GetTracingSampleRate() float64 {
	if b != nil && b.Tracing != nil {
		return b.Tracing.SampleRate
	}
	return 0
}

func (b *Bootstrap) GetLogLevel() string {
	if b != nil && b.Log != nil {
		return b.Log.Level
	}
	return ""
}

func (b *Bootstrap) GetLogEnv() string {
	if b != nil && b.Log != nil {
		return b.Log.Env
	}
	return ""
}

func (b *Bootstrap) GetDatabaseSource() string {
	if b != nil && b.Data != nil && b.Data.Database != nil {
		return b.Data.Database.Source
	}
	return ""
}

func (b *Bootstrap) GetHTTPAddr() string {
	if b != nil && b.Server != nil && b.Server.Http != nil {
		return b.Server.Http.Addr
	}
	return ""
}

func (b *Bootstrap) GetHTTPTimeout() string {
	if b != nil && b.Server != nil && b.Server.Http != nil {
		return b.Server.Http.Timeout
	}
	return ""
}

func (b *Bootstrap) GetGRPCAddr() string {
	if b != nil && b.Server != nil && b.Server.Grpc != nil {
		return b.Server.Grpc.Addr
	}
	return ""
}

func (b *Bootstrap) GetGRPCTimeout() string {
	if b != nil && b.Server != nil && b.Server.Grpc != nil {
		return b.Server.Grpc.Timeout
	}
	return ""
}
