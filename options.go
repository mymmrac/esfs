package esfs

import (
	"io/fs"
	"os"
	"time"

	"github.com/valyala/fasthttp"
)

func WithServer(server *fasthttp.Server) Option {
	return func(cfg *Config) error {
		cfg.Server = server
		return nil
	}
}

func WithGracefulShutdown() Option {
	return func(cfg *Config) error {
		cfg.GracefulShutdown = true
		return nil
	}
}

func WithDir(dir string) Option {
	return func(cfg *Config) error {
		cfg.Dir = dir
		return nil
	}
}

func WithFS(fileSystem fs.FS) Option {
	return func(cfg *Config) error {
		cfg.FileSystem = fileSystem
		return nil
	}
}

func WithSubDir(dir string) Option {
	return func(cfg *Config) error {
		cfg.SubDir = dir
		return nil
	}
}

func WithTempDir(dir string) Option {
	return func(cfg *Config) error {
		cfg.TempDir = dir
		return nil
	}
}

func WithTempDirPattern(pattern string) Option {
	return func(cfg *Config) error {
		cfg.TempDirPattern = pattern
		return nil
	}
}

func WithTempFilesPerm(perm os.FileMode) Option {
	return func(cfg *Config) error {
		cfg.TempFilesPerm = perm
		return nil
	}
}

func WithIndexNames(names ...string) Option {
	return func(cfg *Config) error {
		cfg.IndexNames = names
		return nil
	}
}

func WithGenerateIndexPages() Option {
	return func(cfg *Config) error {
		cfg.GenerateIndexPages = true
		return nil
	}
}

func WithCompress() Option {
	return func(cfg *Config) error {
		cfg.Compress = true
		return nil
	}
}

func WithCompressBrotli() Option {
	return func(cfg *Config) error {
		cfg.Compress = true
		cfg.CompressBrotli = true
		return nil
	}
}

func WithAcceptByteRange() Option {
	return func(cfg *Config) error {
		cfg.AcceptByteRange = true
		return nil
	}
}

func WithPathRewrite(rewrite fasthttp.PathRewriteFunc) Option {
	return func(cfg *Config) error {
		cfg.PathRewrite = rewrite
		return nil
	}
}

func WithPathRewriteToRoot() Option {
	return func(cfg *Config) error {
		cfg.PathRewriteToRoot = true
		return nil
	}
}

func WithPathNotFound(handler fasthttp.RequestHandler) Option {
	return func(cfg *Config) error {
		cfg.PathNotFound = handler
		return nil
	}
}

func WithCacheDuration(duration time.Duration) Option {
	return func(cfg *Config) error {
		cfg.CacheDuration = duration
		return nil
	}
}
