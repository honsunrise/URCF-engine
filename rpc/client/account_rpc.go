package client

import (
	"github.com/zhsyourai/URCF-engine/rpc/shared"
	"github.com/zhsyourai/URCF-engine/models"
	"net/rpc"
)

type AccountRPC struct {
	client *rpc.Client
}

const AccountRPCName = "AccountRPC"

func NewAccountRPC(address string) (*AccountRPC, error) {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}
	return &AccountRPC{
		client: client,
	}, nil
}

func (t *AccountRPC) Register(id string, password string, role []string) (reply models.Account, err error) {
	param := &shared.RegisterParam{
		Id: id,
		Password: password,
		Role: role,
	}
	err = t.client.Call(AccountRPCName + ".Register", param, &reply)
	return
}

func (t *AccountRPC) Verify(args *shared.VerifyParam, reply *models.Account) (err error) {
	return
}

func (t *AccountRPC) GetAll(_, reply *[]models.Account) (err error) {
	return
}

func (t *AccountRPC) GetByID(id string, reply *models.Account) (err error) {
	return
}

func (t *AccountRPC) DeleteByID(id string, reply *models.Account) (err error) {
	return
}
