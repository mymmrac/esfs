package esfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
)

type Config struct {
	Addr             string           `validate:"hostname_port"`
	Server           *fasthttp.Server `validate:"required"`
	GracefulShutdown bool             `validate:"-"`
	SubDir           string           `validate:"-"`

	Dir string `validate:"omitempty,dir"`

	FileSystem     fs.FS       `validate:"required_without=Dir"`
	TempDir        string      `validate:"omitempty,required_with=FileSystem,dir"`
	TempDirPattern string      `validate:"-"`
	TempFilesPerm  os.FileMode `validate:"required_with=FileSystem"`

	IndexNames         []string                 `validate:"-"`
	GenerateIndexPages bool                     `validate:"-"`
	Compress           bool                     `validate:"-"`
	CompressBrotli     bool                     `validate:"omitempty,excluded_unless=Compress false"`
	AcceptByteRange    bool                     `validate:"-"`
	PathRewrite        fasthttp.PathRewriteFunc `validate:"-"`
	PathRewriteToRoot  bool                     `validate:"-"`
	PathNotFound       fasthttp.RequestHandler  `validate:"-"`
	CacheDuration      time.Duration            `validate:"gte=0"`
}

type Option func(cfg *Config) error

func ServeFS(addr string, fileSystem fs.FS, options ...Option) error {
	op := []Option{
		WithFS(fileSystem),
	}
	return Serve(addr, append(op, options...)...)
}

func ServeDir(addr, dir string, options ...Option) error {
	op := []Option{
		WithDir(dir),
	}
	return Serve(addr, append(op, options...)...)
}

type DiscardLogger struct{}

func (d DiscardLogger) Printf(format string, args ...any) {}

func Serve(addr string, options ...Option) error {
	cfg := &Config{
		Addr: addr,
		Server: &fasthttp.Server{
			Logger: DiscardLogger{},
		},

		TempDirPattern: "esfs-",
		TempFilesPerm:  0o700,

		IndexNames:   []string{"index.html"},
		PathNotFound: func(ctx *fasthttp.RequestCtx) { ctx.NotFound() },
	}

	for _, op := range options {
		if err := op(cfg); err != nil {
			return fmt.Errorf("options: %w", err)
		}
	}

	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		return fmt.Errorf("config validation: %w", err)
	}

	var (
		rootDir    string
		fileSystem fs.FS
	)

	if cfg.FileSystem != nil {
		if cfg.SubDir != "" {
			cfg.FileSystem, err = fs.Sub(cfg.FileSystem, cfg.SubDir)
			if err != nil {
				return fmt.Errorf("fs sub dir: %w", err)
			}
		}

		fileSystem = cfg.FileSystem

		rootDir, err = os.MkdirTemp(cfg.TempDir, cfg.TempDirPattern)
		if err != nil {
			return fmt.Errorf("make temp dir: %w", err)
		}

		err = copyFSToDisk(cfg, cfg.FileSystem, rootDir)
		if err != nil {
			return fmt.Errorf("copy FS to disk: %w", err)
		}

		defer func() {
			err = os.RemoveAll(rootDir)
			if err != nil {
				fmt.Printf("Clean up failed: %s\n", err)
			}
		}()
	} else {
		if cfg.SubDir != "" {
			cfg.Dir = filepath.Join(cfg.Dir, cfg.SubDir)
		}

		rootDir = cfg.Dir
		fileSystem = os.DirFS(rootDir)
	}

	if cfg.PathRewriteToRoot {
		originalPathRewrite := cfg.PathRewrite
		cfg.PathRewrite = func(ctx *fasthttp.RequestCtx) []byte {
			path := ctx.Path()
			if originalPathRewrite != nil {
				path = originalPathRewrite(ctx)
			}

			for _, index := range append([]string{""}, cfg.IndexNames...) {
				resultingPath := strings.TrimPrefix(filepath.Join(string(path), index), "/")
				if !fs.ValidPath(resultingPath) {
					continue
				}

				_, err = fs.Stat(fileSystem, resultingPath)
				if err == nil || !errors.Is(err, fs.ErrNotExist) {
					return path
				}
			}

			return []byte("/")
		}
	}

	fastFS := fasthttp.FS{
		Root:                   rootDir,
		AllowEmptyRoot:         false,
		IndexNames:             cfg.IndexNames,
		GenerateIndexPages:     cfg.GenerateIndexPages,
		Compress:               cfg.Compress,
		CompressBrotli:         cfg.CompressBrotli,
		CompressRoot:           "",
		AcceptByteRange:        cfg.AcceptByteRange,
		PathRewrite:            cfg.PathRewrite,
		PathNotFound:           cfg.PathNotFound,
		CacheDuration:          cfg.CacheDuration,
		CompressedFileSuffix:   "",
		CompressedFileSuffixes: nil,
		CleanStop:              nil,
	}
	cfg.Server.Handler = fastFS.NewRequestHandler()

	fmt.Println("Running...")

	if !cfg.GracefulShutdown {
		return cfg.Server.ListenAndServe(cfg.Addr)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{}, 1)

	go func() {
		<-sigs
		fmt.Println("Stopping...")

		err = cfg.Server.Shutdown()
		if err != nil {
			fmt.Printf("Error: shutdown server: %s", err)
		}

		done <- struct{}{}
	}()

	go func() {
		err = cfg.Server.ListenAndServe(cfg.Addr)
		if err != nil {
			fmt.Printf("Error: start server: %s", err)
			os.Exit(1)
		}
	}()

	<-done
	fmt.Println("Done")

	return nil
}

func copyFSToDisk(cfg *Config, fileSystem fs.FS, rootDir string) error {
	err := fs.WalkDir(fileSystem, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		p := filepath.Join(rootDir, path)

		if entry.IsDir() {
			err = os.MkdirAll(p, cfg.TempFilesPerm)
			if err != nil {
				return err
			}
			return nil
		}

		var (
			f  *os.File
			ef fs.File
		)

		f, err = os.OpenFile(p, os.O_CREATE|os.O_WRONLY, cfg.TempFilesPerm)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		ef, err = fileSystem.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = ef.Close()
		}()

		_, err = io.Copy(f, ef)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
