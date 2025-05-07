package ctxconfig

import (
	"context"

	"fknsrs.biz/p/ytmusic/internal/config"
)

// context registration

var configKey int

func WithConfig(ctx context.Context, c config.Config) context.Context {
	return context.WithValue(ctx, &configKey, c)
}

func GetConfig(ctx context.Context) config.Config {
	if v := ctx.Value(&configKey); v != nil {
		return v.(config.Config)
	}

	return config.Config{}
}

// main interface

func DataFile(ctx context.Context, section, name string) string {
	return GetConfig(ctx).DataFile(section, name)
}
