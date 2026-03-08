package server_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kazhuravlev/just"
	"github.com/kazhuravlev/lrpc/ctypes"
	"github.com/kazhuravlev/lrpc/server"
)

func ExampleNew() {
	router := just.Must(server.New(server.NewOptions(server.WithName("example_server"))))

	{
		type AuthReq struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		type AuthResp struct {
			Token string `json:"token"`
		}

		errUnableToAuth := errors.New("unable to auth")
		authHandler := func(ctx context.Context, id ctypes.ID, req AuthReq) (*AuthResp, error) {
			if req.Username == "root" && req.Password == "admin" {
				return &AuthResp{Token: "some-token"}, nil
			}

			return nil, fmt.Errorf("bad credentials: %w", errUnableToAuth)
		}

		server.RegisterHandler(
			router,
			"auth-v1",
			authHandler,
			map[error]ctypes.ErrorCode{
				errUnableToAuth: 334455,
			},
		)
	}

	{
		type LogoutReq struct {
			Token string `json:"token"`
		}

		logoutHandler := func(ctx context.Context, id ctypes.ID, req LogoutReq) error {
			if req.Token == "some-token" {
				return nil
			}

			return errors.New("unable to logout")
		}

		server.RegisterHandlerNoResponse(router, "logout-v1", logoutHandler, nil)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/lrpc/{method}", router.HTTPHandler())

	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	go func() {
		time.Sleep(time.Second)
		_ = srv.Shutdown(context.Background()) //nolint:errcheck
	}()
	_ = srv.ListenAndServe() //nolint:errcheck

	// Output:
}
