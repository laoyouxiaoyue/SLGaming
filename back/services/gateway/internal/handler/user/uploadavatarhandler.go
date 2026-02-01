// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

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

		l := user.NewUploadAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadAvatar(&types.UploadAvatarRequest{Avatar: avatarUrl})
		if err != nil {
			_ = os.Remove(filePath)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			if resp != nil && resp.Code != 0 {
				_ = os.Remove(filePath)
			}
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
