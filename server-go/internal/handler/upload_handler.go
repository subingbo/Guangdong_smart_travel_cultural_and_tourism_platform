package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
)

const maxUploadBytes = 5 << 20

var uploadExtension = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type UploadHandler struct {
	dir string
}

func NewUploadHandler(dir string) *UploadHandler { return &UploadHandler{dir: dir} }

// Upload stores one image under <dir>/<kind>/YYYY/MM/<random>.<ext>.
func (h *UploadHandler) Upload(c *gin.Context) {
	if _, ok := userIDRequired(c); !ok {
		return
	}
	kind := c.DefaultQuery("kind", "blog")
	if kind != "blog" && kind != "avatar" {
		common.Fail(c, common.BadRequest("kind must be 'blog' or 'avatar'"))
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+512*1024)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			common.Fail(c, common.BadRequest("file is too large (max 5MB)"))
			return
		}
		common.Fail(c, common.BadRequest("file is required"))
		return
	}
	defer file.Close()
	if header.Size > maxUploadBytes {
		common.Fail(c, common.BadRequest("file is too large (max 5MB)"))
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		common.Fail(c, common.Internal("cannot read upload"))
		return
	}
	if len(data) == 0 {
		common.Fail(c, common.BadRequest("file is empty"))
		return
	}
	if len(data) > maxUploadBytes {
		common.Fail(c, common.BadRequest("file is too large (max 5MB)"))
		return
	}

	ext, ok := uploadExtension[http.DetectContentType(data)]
	if !ok {
		common.Fail(c, common.BadRequest("only jpeg, png, webp or gif images are allowed"))
		return
	}

	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		common.Fail(c, common.Internal("cannot generate file name"))
		return
	}
	sub := time.Now().Format("2006/01")
	targetDir := filepath.Join(h.dir, kind, filepath.FromSlash(sub))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		common.Fail(c, common.Internal("cannot create upload directory"))
		return
	}
	name := hex.EncodeToString(random) + ext
	if err := os.WriteFile(filepath.Join(targetDir, name), data, 0o644); err != nil {
		common.Fail(c, common.Internal("cannot save upload"))
		return
	}

	common.OK(c, gin.H{"url": "/uploads/" + kind + "/" + sub + "/" + name})
}
