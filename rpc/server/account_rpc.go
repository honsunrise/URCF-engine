package server

import (
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/rpc/shared"
	"github.com/zhsyourai/URCF-engine/services/account"
	"net/rpc"
)

type AccountRPC struct {
	service account.Service
}

func RegisterAccountRPC() error {
	err := rpc.RegisterName("AccountRPC", &AccountRPC{
		service: account.GetInstance(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *AccountRPC) Register(args *shared.RegisterParam, reply *models.Account) (err error) {
	*reply, err = t.service.Register(args.Username, args.Password, args.Role)
	return
}

func (t *AccountRPC) Verify(args *shared.VerifyParam, reply *models.Account) (err error) {
	*reply, err = t.service.Verify(args.Username, args.Password)
	return
}

func (t *AccountRPC) GetAll(_, reply *[]models.Account) (err error) {
	*reply, err = t.service.GetAll()
	return
}

func (t *AccountRPC) GetByID(username string, reply *models.Account) (err error) {
	*reply, err = t.service.GetByUsername(username)
	return
}

func (t *AccountRPC) DeleteByID(username string, reply *models.Account) (err error) {
	*reply, err = t.service.DeleteByUsername(username)
	return
}
