package rpc

import (
	"github.com/zhsyourai/URCF-engine/rpc/shared"
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/zhsyourai/URCF-engine/models"
)

type AccountRPC struct {
	service account.Service
}

func (t *AccountRPC) Register(args *shared.RegisterParam, reply *models.Account) (err error) {
	*reply, err = t.service.Register(args.Id, args.Password, args.Role)
	return
}

func (t *AccountRPC) Verify(args *shared.VerifyParam, reply *models.Account) (err error) {
	*reply, err = t.service.Verify(args.Id, args.Password)
	return
}

func (t *AccountRPC) GetAll(_, reply *[]models.Account) (err error) {
	*reply, err = t.service.GetAll()
	return
}

func (t *AccountRPC) GetByID(id string, reply *models.Account) (err error) {
	*reply, err = t.service.GetByID(id)
	return
}

func (t *AccountRPC) DeleteByID(id string, reply *models.Account) (err error) {
	*reply, err = t.service.DeleteByID(id)
	return
}
