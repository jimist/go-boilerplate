package core

import (
	"context"
	"time"

	"git.cafebazaar.ir/arcana261/golang-boilerplate/internal/pkg/cache"
	"git.cafebazaar.ir/arcana261/golang-boilerplate/internal/pkg/errors"
	"git.cafebazaar.ir/arcana261/golang-boilerplate/internal/pkg/provider"
	"git.cafebazaar.ir/arcana261/golang-boilerplate/pkg/postview"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	cacheExpireTime = 1 * time.Minute
)

type core struct {
	provider provider.PostProvider
	cache    cache.PostCache
}

func New(provider provider.PostProvider, cache cache.PostCache) postview.PostViewServer {
	return &core{
		provider: provider,
		cache:    cache,
	}
}

func (c *core) GetPost(ctx context.Context, request *postview.GetPostRequest) (*postview.GetPostResponse, error) {
	post, ok, err := c.cache.Get(ctx, request.Token)
	if err != nil {
		logrus.WithError(err).WithFields(map[string]interface{}{
			"token": request.Token,
		}).Error("failed to load data from cache")
	}

	if !ok {
		post, err = c.provider.GetPost(ctx, request.Token)
		if err != nil {
			if xerrors.Is(err, provider.ErrNotFound) {
				return nil, status.Error(codes.NotFound, "post not found")
			}

			return nil, errors.WrapWithExtra(err, "failed to acquire post", map[string]interface{}{
				"request": request,
			})
		}

		err = c.cache.Set(ctx, post, cacheExpireTime)
		if err != nil {
			logrus.WithError(err).WithFields(map[string]interface{}{
				"token": request.Token,
			}).Error("failed to set data in cache")
		}
	}

	return &postview.GetPostResponse{
		Post: post,
	}, nil
}
