package grpc

import (
	"context"
	"errors"

	pb "github.com/Karzoug/goph_keeper/common/grpc/server"
	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/delivery/grpc/interceptor/auth"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

func (s *server) ListVaultItems(ctx context.Context, req *pb.ListVaultItemsRequest) (*pb.ListVaultItemsResponse, error) {
	const op = "list vault items"

	email, err := auth.EmailFromContext(ctx)
	if err != nil {
		return nil, pb.ErrEmptyAuthData
	}
	items, err := s.service.ListVaultItems(ctx, email, req.Since)
	if err != nil {
		s.logger.Error(op, sl.Error(err))
		return nil, pb.ErrInternal
	}
	pbItems := make([]*pb.VaultItem, len(items))
	for i := 0; i < len(items); i++ {
		pbItems[i] = &pb.VaultItem{
			Id:              items[i].ID,
			Name:            items[i].Name,
			Itype:           pb.IType(items[i].Type),
			Value:           items[i].Value,
			ServerUpdatedAt: items[i].ServerUpdatedAt,
		}
	}
	return &pb.ListVaultItemsResponse{
		Items: pbItems,
	}, nil
}

func (s *server) SetVaultItem(ctx context.Context, req *pb.SetVaultItemRequest) (*pb.SetVaultItemResponse, error) {
	const op = "set vault item"

	email, err := auth.EmailFromContext(ctx)
	if err != nil {
		return nil, pb.ErrEmptyAuthData
	}
	t, err := s.service.SetVaultItem(ctx, email, vault.Item{
		ID:              req.Item.Id,
		Name:            req.Item.Name,
		Type:            vault.ItemType(req.Item.Itype),
		Value:           req.Item.Value,
		ServerUpdatedAt: req.Item.ServerUpdatedAt,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrVaultItemVersionConflict):
			return nil, pb.ErrVaultItemConflictVersion
		case errors.Is(err, service.ErrVaultItemValueTooBig):
			return nil, pb.ErrVaultItemValueTooBig
		default:
			s.logger.Error(op, sl.Error(err))
			return nil, pb.ErrInternal
		}
	}
	return &pb.SetVaultItemResponse{
		ServerUpdatedAt: t,
	}, nil
}
