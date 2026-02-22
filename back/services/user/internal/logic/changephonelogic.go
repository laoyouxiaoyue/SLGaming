package logic

import (
	"context"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ChangePhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePhoneLogic {
	return &ChangePhoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangePhoneLogic) ChangePhone(in *user.ChangePhoneRequest) (*user.ChangePhoneResponse, error) {
	userID := in.GetUserId()
	oldPhone := strings.TrimSpace(in.GetOldPhone())
	newPhone := strings.TrimSpace(in.GetNewPhone())
	if userID == 0 || oldPhone == "" || newPhone == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, old_phone and new_phone are required")
	}
	if oldPhone == newPhone {
		return nil, status.Error(codes.InvalidArgument, "new_phone must be different from old_phone")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("id = ?", userID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if strings.TrimSpace(u.Phone) != oldPhone {
		return nil, status.Error(codes.InvalidArgument, "old_phone mismatch")
	}

	var count int64
	if err := db.Model(&model.User{}).Where("phone = ? AND id <> ?", newPhone, u.ID).Count(&count).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if count > 0 {
		return nil, status.Error(codes.AlreadyExists, "phone already used")
	}

	if err := db.Model(&u).Update("phone", newPhone).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 将新手机号添加到布隆过滤器
	// 注意：旧手机号无法从布隆过滤器中删除，这是布隆过滤器的设计限制
	// 旧手机号留在过滤器中会略微增加假阳性率，但不会影响正确性
	if l.svcCtx.BloomFilter != nil {
		if err := l.svcCtx.BloomFilter.Phone.Add(l.ctx, newPhone); err != nil {
			l.Logger.Errorf("add new phone to bloom filter failed: %v", err)
			// 不影响主流程，仅记录错误
		} else {
			l.Logger.Infof("new phone added to bloom filter: user_id=%d, phone=%s", userID, newPhone)
		}
	}

	return &user.ChangePhoneResponse{Success: true}, nil
}
