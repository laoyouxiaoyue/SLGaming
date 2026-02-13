// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"SLGaming/back/pkg/avatarmq"
	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// UploadAvatarHandler 上传头像
// @Summary 上传头像
// @Description 上传用户头像，支持 jpg、png、gif、webp 格式，最大 5MB
// @Tags 用户
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像文件"
// @Success 200 {object} types.UploadAvatarResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/avatar [post]
// @Security BearerAuth
func UploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxSizeMB := svcCtx.Config.Upload.MaxSizeMB
		if maxSizeMB <= 0 {
			maxSizeMB = 5
		}
		maxBytes := maxSizeMB * 1024 * 1024
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		if err := r.ParseMultipartForm(maxBytes); err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 400, Msg: "上传文件过大或表单解析失败"},
			})
			return
		}

		file, header, err := r.FormFile("avatar")
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 400, Msg: "缺少头像文件字段 avatar"},
			})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := svcCtx.Config.Upload.AllowedExt
		if len(allowed) == 0 {
			allowed = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
		}
		isAllowed := false
		for _, a := range allowed {
			if strings.ToLower(a) == ext {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 400, Msg: "不支持的文件类型"},
			})
			return
		}

		baseDir := strings.TrimSpace(svcCtx.Config.Upload.LocalDir)
		if baseDir == "" {
			baseDir = "uploads"
		}
		avatarDir := filepath.Join(baseDir, "avatars")
		if err := os.MkdirAll(avatarDir, 0755); err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 500, Msg: "创建上传目录失败"},
			})
			return
		}

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		filePath := filepath.Join(avatarDir, filename)
		out, err := os.Create(filePath)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 500, Msg: "保存文件失败"},
			})
			return
		}
		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			_ = os.Remove(filePath)
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 500, Msg: "写入文件失败"},
			})
			return
		}
		out.Close()

		baseURL := strings.TrimRight(strings.TrimSpace(svcCtx.Config.Upload.BaseURL), "/")
		if baseURL == "" {
			baseURL = "/uploads"
		}
		avatarUrl := fmt.Sprintf("%s/avatars/%s", baseURL, filename)

		userID, err := middleware.GetUserID(r.Context())
		if err != nil {
			_ = os.Remove(filePath)
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
			})
			return
		}

		if svcCtx.EventProducer == nil {
			_ = os.Remove(filePath)
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: 503, Msg: "审核服务暂时不可用"},
			})
			return
		}

		defaultAvatarURL := strings.TrimSpace(svcCtx.Config.Upload.DefaultAvatarURL)
		if defaultAvatarURL == "" {
			defaultAvatarURL = "https://lf-flow-web-cdn.doubao.com/obj/flow-doubao/samantha/logo-icon-white-bg.png"
		}

		requestID := fmt.Sprintf("avatar-%d", time.Now().UnixNano())
		payload := &avatarmq.AvatarModerationPayload{
			UserID:           userID,
			AvatarURL:        avatarUrl,
			DefaultAvatarURL: defaultAvatarURL,
			RequestID:        requestID,
			SubmittedAt:      time.Now().Unix(),
		}
		if err := publishAvatarEvent(r.Context(), svcCtx, payload); err != nil {
			_ = os.Remove(filePath)
			code, msg := utils.HandleError(err, logx.WithContext(r.Context()), "PublishAvatarEvent")
			httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			})
			return
		}

		httpx.OkJsonCtx(r.Context(), w, &types.UploadAvatarResponse{
			BaseResp: types.BaseResp{Code: 0, Msg: "头像已提交审核"},
			Data: types.UploadAvatarData{
				AvatarUrl: defaultAvatarURL,
			},
		})
	}
}

func publishAvatarEvent(ctx context.Context, svcCtx *svc.ServiceContext, payload *avatarmq.AvatarModerationPayload) error {
	if svcCtx.EventProducer == nil {
		return fmt.Errorf("rocketmq producer not initialized")
	}
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := primitive.NewMessage(avatarmq.AvatarEventTopic(), body)
	msg.WithTag(avatarmq.EventTypeAvatarSubmit())
	_, err = svcCtx.EventProducer.SendSync(ctx, msg)
	return err
}
