package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"encoding/gob"

	"github.com/pioz/faker"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/common/model/vault"
)

type SyncVaultSuite struct {
	commonTestSuite

	token_2      string
	conn_2       *grpc.ClientConn
	grpcClient_2 pb.GophKeeperServiceClient
}

// SetupSuite bootstraps suite dependencies
func (suite *SyncVaultSuite) SetupSuite() {
	suite.commonTestSuite.SetupSuite()

	// second client
	config := &tls.Config{
		ServerName:         suite.host,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
	}
	addr := suite.host + ":" + suite.port
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(config)), grpc.WithBlock())
	if err != nil {
		suite.Require().Error(err, "gRPC dial error")
		return
	}
	suite.conn_2 = conn
	suite.grpcClient_2 = pb.NewGophKeeperServiceClient(conn)
}

func (suite *SyncVaultSuite) TestVault() {
	ctx, cancel := context.WithTimeout(context.Background(), testsTimeout)
	defer cancel()

	itemIDs := make([]string, 3)
	itemIDs[0] = xid.New().String()
	itemIDs[1] = xid.New().String()
	itemIDs[2] = xid.New().String()

	var (
		serverUpdatedAtForTextItem int64
		lastUpdateFirstClient      int64
	)

	suite.Run("add vault items", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		// text data
		respSet, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:    itemIDs[0],
				Name:  faker.String(),
				Itype: pb.IType(vault.Text),
				Value: []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)
		serverUpdatedAtForTextItem = respSet.ServerUpdatedAt

		// binary data
		b := make([]byte, 50*1024)
		_, err = rand.Read(b)
		suite.Require().NoError(err)
		respSet, err = suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:    itemIDs[1],
				Name:  faker.String(),
				Itype: pb.IType(vault.Binary),
				Value: b,
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)

		// encrypted card data
		type Card struct {
			Holder  string
			Expired string
			Number  string
			CSC     string
		}
		c := Card{
			Holder:  faker.FirstName() + " " + faker.LastName(),
			Expired: faker.DigitsWithSize(2) + "/" + faker.DigitsWithSize(2),
			Number:  faker.DigitsWithSize(15),
			CSC:     faker.DigitsWithSize(4),
		}
		bb := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(bb)
		err = enc.Encode(c)
		suite.Require().NoError(err)

		secret := make([]byte, 32)
		_, err = rand.Read(b)
		suite.Require().NoError(err)
		data := encrypt(bb.Bytes(), secret)

		respSet, err = suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:    itemIDs[2],
				Name:  faker.String(),
				Itype: pb.IType(vault.Card),
				Value: data,
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)

		lastUpdateFirstClient = respSet.ServerUpdatedAt
	})

	suite.Run("login into empty second client & sync", func() {
		resp, err := suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: suite.email,
			Hash:  suite.authHash,
		})
		suite.Require().NoError(err, "gRPC user login with mail verification code error", err)
		suite.Assert().NotEqual(0, len(resp.Token), "Token not found in response")

		suite.token_2 = resp.Token

		ctx := newContextWithAuthData(ctx, suite.token_2)

		respList, err := suite.grpcClient_2.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: 0,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 3, "Wrong number of items in response")

		lastUpdatedAt := respList.Items[0].ServerUpdatedAt
		for i := 0; i < 3; i++ {
			suite.Assert().Contains(itemIDs, respList.Items[i].Id)
			if lastUpdatedAt < respList.Items[i].ServerUpdatedAt {
				lastUpdatedAt = respList.Items[i].ServerUpdatedAt
			}
		}

		respList, err = suite.grpcClient_2.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: lastUpdatedAt,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 0, "Wrong number of items in response")
	})

	itemIDs = append(itemIDs, xid.New().String())

	suite.Run("add/edit vault items in second client & sync", func() {
		ctx := newContextWithAuthData(ctx, suite.token_2)

		// edit text data
		_, err := suite.grpcClient_2.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemIDs[0],
				Name:            faker.String(),
				Itype:           pb.IType(vault.Text),
				Value:           []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
				ServerUpdatedAt: serverUpdatedAtForTextItem,
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)

		// add new binary data
		b := make([]byte, 50*1024)
		_, err = rand.Read(b)
		suite.Require().NoError(err)
		_, err = suite.grpcClient_2.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:    itemIDs[3],
				Name:  faker.String(),
				Itype: pb.IType(vault.Binary),
				Value: b,
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)
	})

	suite.Run("sync first client", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		respList, err := suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: lastUpdateFirstClient,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 2, "Wrong number of items in response")
	})

}

func encrypt(plain, secret []byte) []byte {
	aes, err := aes.NewCipher(secret)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		panic(err)
	}
	cipher := gcm.Seal(nonce, nonce, plain, nil)
	return cipher
}
