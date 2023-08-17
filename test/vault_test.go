package main

import (
	"context"
	"time"

	"github.com/pioz/faker"
	"github.com/rs/xid"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/common/model/vault"
)

type VaultSuite struct {
	commonTestSuite
}

func (suite *VaultSuite) TestVault() {
	ctx, cancel := context.WithTimeout(context.Background(), testsTimeout)
	defer cancel()

	itemId := xid.New().String()
	var setServerUpdatedAt int64

	suite.Run("add vault item", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		now := time.Now().UnixMicro()

		respSet, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:    itemId,
				Name:  faker.String(),
				Itype: pb.IType(vault.Text),
				Value: []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)
		setServerUpdatedAt = respSet.ServerUpdatedAt
		suite.Assert().LessOrEqual(now, setServerUpdatedAt, "returned server update time must be equal or greater than time of request")

		respList, err := suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: now - 1, // -1 to avoid case server update time equal time of request
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 1, "returned wrong number of vault items")

		respList, err = suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: setServerUpdatedAt,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 0, "returned wrong number of vault items")
	})

	suite.Run("update vault item", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		updatedName := faker.String()

		respSet, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemId,
				Name:            updatedName,
				Itype:           pb.IType(vault.Text),
				Value:           []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
				ServerUpdatedAt: setServerUpdatedAt,
			},
		})
		suite.Require().NoError(err, "gRPC update vault item error", err)
		setServerUpdatedAt = respSet.ServerUpdatedAt

		respList, err := suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: setServerUpdatedAt - 1,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Require().Len(respList.Items, 1, "returned wrong number of vault items")
		suite.Assert().Equal(updatedName, respList.Items[0].Name, "returned wrong vault item name")
	})

	suite.Run("update vault item with conflict version", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		_, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemId,
				Name:            faker.String(),
				Itype:           pb.IType(vault.Text),
				Value:           []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
				ServerUpdatedAt: setServerUpdatedAt - 10, // client has old version
			},
		})
		suite.Assert().ErrorIs(err, pb.ErrVaultItemConflictVersion)
	})

	suite.Run("delete vault item", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		_, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemId,
				Name:            "",
				Itype:           pb.IType(vault.Text),
				Value:           nil,
				ServerUpdatedAt: setServerUpdatedAt, // client has old version
				IsDeleted:       true,
			},
		})
		suite.Assert().NoError(err)
	})
}
