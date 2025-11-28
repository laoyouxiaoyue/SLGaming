package logic

import (
	"context"
	"errors"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *user.GetUserRequest) (*user.GetUserResponse, error) {
	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	var err error

	switch {
	case in.GetId() != 0:
		err = db.Where("id = ?", in.GetId()).First(&u).Error
	case in.GetUid() != 0:
		err = db.Where("uid = ?", in.GetUid()).First(&u).Error
	case in.GetPhone() != "":
		err = db.Where("phone = ?", in.GetPhone()).First(&u).Error
	default:
		return nil, status.Error(codes.InvalidArgument, "missing query condition")
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.GetUserResponse{
		User: toUserInfo(&u),
	}, nil
}
