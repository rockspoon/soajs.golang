package soajsgo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitMiddleware(t *testing.T) {
	tt := []struct {
		name        string
		config      Config
		envRegAPI   string
		envEnv      string
		expectedErr error
	}{
		{
			name:        "no envs",
			config:      Config{},
			envRegAPI:   "",
			envEnv:      "",
			expectedErr: nil,
		},
		{
			name:        "registry error",
			config:      Config{},
			envRegAPI:   "api",
			envEnv:      "env",
			expectedErr: errors.New("could not create new registry: service name and env code are required"),
		},
	}
	lastEnvRegAPI := os.Getenv(EnvRegistryAPIAddress)
	lastEnvEnv := os.Getenv(EnvEnv)
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.Setenv(EnvRegistryAPIAddress, tc.envRegAPI))
			require.NoError(t, os.Setenv(EnvEnv, tc.envEnv))

			ctx, cancel := context.WithCancel(context.Background())
			_, err := InitMiddleware(ctx, tc.config)
			cancel()
			assert.Equal(t, tc.expectedErr, err)

			assert.NoError(t, os.Setenv(EnvRegistryAPIAddress, lastEnvRegAPI))
			assert.NoError(t, os.Setenv(EnvEnv, lastEnvEnv))
		})
	}
}

func TestRegistry_Middleware(t *testing.T) {
	tt := []struct {
		name            string
		headerInfo      string
		reg             Registry
		expectedSoaData ContextData
	}{
		{
			name:            "bad header",
			headerInfo:      "nil",
			reg:             Registry{},
			expectedSoaData: ContextData{},
		},
		{
			name:            "empty header",
			headerInfo:      "",
			reg:             Registry{},
			expectedSoaData: ContextData{},
		},
		{
			name:            "all ok",
			headerInfo:      `{"device":"iPhone"}`,
			reg:             Registry{Name: "ok"},
			expectedSoaData: ContextData{Device: "iPhone", Reg: Registry{Name: "ok"}},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				soa := r.Context().Value(SoajsKey)
				if soa != nil {
					assert.Equal(t, tc.expectedSoaData, soa.(ContextData))
				} else {
					assert.Nil(t, soa)
				}
				_, _ = w.Write([]byte("ok"))
			})
			req := httptest.NewRequest("", "http://localhost:8080/", nil)
			req.Header.Set(headerDataName, tc.headerInfo)
			rec := httptest.NewRecorder()
			middleware := tc.reg.Middleware(handler)
			middleware.ServeHTTP(rec, req)
		})
	}
}

func TestHeaderData(t *testing.T) {
	tt := []struct {
		name         string
		data         string
		expectedInfo *HeaderInfo
		expectedErr  error
	}{
		{
			name:         "empty header",
			data:         "",
			expectedInfo: nil,
			expectedErr:  nil,
		},
		{
			name:         "bad header",
			data:         "nil",
			expectedInfo: nil,
			expectedErr:  errors.New("unable to parse SOAJS header"),
		},
		{
			name:         "all ok",
			data:         `{"device":"iPhone"}`,
			expectedInfo: &HeaderInfo{Device: "iPhone"},
			expectedErr:  nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("", "http://localhost:8080/", nil)
			req.Header.Set(headerDataName, tc.data)

			info, err := headerData(req)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedInfo, info)
		})
	}
}

func TestHost_Path(t *testing.T) {
	tt := []struct {
		name         string
		host         Host
		args         []string
		expectedPath string
	}{
		{
			name: "1",
			host: Host{
				Host: "localhost",
				Port: 8080,
			},
			args:         []string{"test"},
			expectedPath: "localhost:8080/",
		},
		{
			name: "2",
			host: Host{
				Host: "localhost",
				Port: 8080,
			},
			args:         []string{"CONTROLLER", "v"},
			expectedPath: "localhost:8080/CONTROLLER/",
		},
		{
			name: "3",
			host: Host{
				Host: "localhost",
				Port: 8080,
			},
			args:         []string{"CONTROLLER", "1", "-"},
			expectedPath: "localhost:8080/CONTROLLER/v1/",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.host.Path(tc.args...)
			assert.Equal(t, tc.expectedPath, p)
		})
	}
}
