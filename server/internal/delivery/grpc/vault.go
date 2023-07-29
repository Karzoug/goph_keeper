package grpc

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/delivery/grpc/interceptor/auth"
)

func (s *server) ListVaultItems(ctx context.Context, req *pb.ListVaultItemsRequest) (*pb.ListVaultItemsResponse, error) {
	const op = "list vault items"

	email, err := auth.EmailFromContext(ctx)
	if err != nil {
		return nil, pb.ErrEmptyAuthData
	}
	var since *time.Time
	if req.Since.IsValid() {
		t := req.Since.AsTime()
		since = &t
	}
	items, t, err := s.service.ListVaultItems(ctx, email, since)
	if err != nil {
		s.logger.Error(op, sl.Error(err))
		return nil, pb.ErrInternal
	}
	pbItems := make([]*pb.VaultItem, len(items))
	for i := 0; i < len(items); i++ {
		pbItems[i] = &pb.VaultItem{
			Name:  items[i].Name,
			Itype: pb.IType(items[i].Type),
			Value: items[i].Value,
		}
	}
	return &pb.ListVaultItemsResponse{
		LastUpdate: timestamppb.New(t),
		Items:      pbItems,
	}, nil
}

func (s *server) SetVaultItem(ctx context.Context, req *pb.SetVaultItemRequest) (*pb.SetVaultItemResponse, error) {
	const op = "set vault item"

	email, err := auth.EmailFromContext(ctx)
	if err != nil {
		return nil, pb.ErrEmptyAuthData
	}
	err = s.service.SetVaultItem(ctx, email, vault.Item{
		Name:  req.Item.Name,
		Type:  vault.ItemType(req.Item.Itype),
		Value: req.Item.Value,
	})
	if err != nil {
		s.logger.Error(op, sl.Error(err))
		return nil, pb.ErrInternal
	}
	return &pb.SetVaultItemResponse{}, nil
}
