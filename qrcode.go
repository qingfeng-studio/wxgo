package wxgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/qingfeng-studio/wxgo/internal/token"
)

const (
	qrCodeCreateAPI = "https://api.weixin.qq.com/cgi-bin/qrcode/create"
	qrCodeShowAPI   = "https://mp.weixin.qq.com/cgi-bin/showqrcode"
)

// QRCodeOption 公众号二维码生成参数
type QRCodeOption struct {
	// SceneID 数字场景值（1~100000）。不填则使用 SceneStr
	SceneID int64
	// SceneStr 字符串场景值（≤64 字节，推荐使用）
	SceneStr string
	// ExpireSeconds 临时二维码有效期（秒，最大 30 天）。永久码忽略此值
	ExpireSeconds int
	// Permanent 是否生成永久二维码。false 为临时码
	Permanent bool
	// Download 是否直接下载二维码图片（微信返回 ticket 时需额外请求）
	Download bool
}

// QRCodeResult 公众号二维码返回结果
type QRCodeResult struct {
	Ticket        string
	ExpireSeconds int
	URL           string
	Image         []byte
	ContentType   string
}

// CreateQRCode 生成公众号二维码
// 根据 Permanent 与 SceneID/SceneStr 选择 action_name，并可选直接拉取图片
func (c *Client) CreateQRCode(ctx context.Context, opt QRCodeOption) (*QRCodeResult, Code, error) {
	actionName, scenePayload, code, err := buildQRCodePayload(opt)
	if err != nil {
		return nil, code, err
	}

	tk, codeToken, err := c.token.GetAccessToken(ctx)
	if err != nil {
		return nil, codeToken, err
	}

	reqURL := qrCodeCreateAPI + "?access_token=" + url.QueryEscape(tk)

	body := map[string]any{
		"action_name": actionName,
		"action_info": map[string]any{
			"scene": scenePayload,
		},
	}
	if !opt.Permanent && opt.ExpireSeconds > 0 {
		body["expire_seconds"] = opt.ExpireSeconds
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, CodeUnknown, fmt.Errorf("marshal qrcode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(raw))
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("create qrcode request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(ctx, req)
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("request qrcode create: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, CodeHTTP, fmt.Errorf("wechat qrcode status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, CodeInvalidResponse, fmt.Errorf("read qrcode response: %w", err)
	}

	var apiResp struct {
		Ticket        string `json:"ticket"`
		ExpireSeconds int    `json:"expire_seconds"`
		URL           string `json:"url"`
		ErrCode       int    `json:"errcode"`
		ErrMsg        string `json:"errmsg"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, CodeInvalidResponse, fmt.Errorf("decode qrcode response: %w", err)
	}

	if apiResp.ErrCode != 0 {
		return nil, CodeAPIError, fmt.Errorf("%w: errcode=%d, errmsg=%s", token.ErrAPIError, apiResp.ErrCode, apiResp.ErrMsg)
	}

	result := &QRCodeResult{
		Ticket:        apiResp.Ticket,
		ExpireSeconds: apiResp.ExpireSeconds,
		URL:           apiResp.URL,
	}

	if !opt.Download || apiResp.Ticket == "" {
		return result, CodeOK, nil
	}

	imgURL := qrCodeShowAPI + "?ticket=" + url.QueryEscape(apiResp.Ticket)
	imgReq, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("create qrcode image request: %w", err)
	}

	imgResp, err := c.http.Do(ctx, imgReq)
	if err != nil {
		return nil, CodeHTTP, fmt.Errorf("download qrcode image: %w", err)
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode < 200 || imgResp.StatusCode >= 300 {
		return nil, CodeHTTP, fmt.Errorf("wechat qrcode image status: %d", imgResp.StatusCode)
	}

	data, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return nil, CodeInvalidResponse, fmt.Errorf("read qrcode image: %w", err)
	}

	result.Image = data
	result.ContentType = imgResp.Header.Get("Content-Type")

	return result, CodeOK, nil
}

func buildQRCodePayload(opt QRCodeOption) (string, map[string]any, Code, error) {
	const maxExpireSeconds = 30 * 24 * 60 * 60 // 30 天

	// 选择场景值
	var scene map[string]any
	switch {
	case opt.SceneStr != "":
		if len(opt.SceneStr) > 64 {
			return "", nil, CodeUnknown, fmt.Errorf("scene_str length must be <=64")
		}
		scene = map[string]any{"scene_str": opt.SceneStr}
	case opt.SceneID != 0:
		if opt.SceneID < 1 || opt.SceneID > 100000 {
			return "", nil, CodeUnknown, fmt.Errorf("scene_id must be in [1,100000]")
		}
		scene = map[string]any{"scene_id": opt.SceneID}
	default:
		return "", nil, CodeUnknown, fmt.Errorf("scene is required (scene_str or scene_id)")
	}

	// 选择 action_name
	isStringScene := opt.SceneStr != ""
	var actionName string
	if opt.Permanent {
		if isStringScene {
			actionName = "QR_LIMIT_STR_SCENE"
		} else {
			actionName = "QR_LIMIT_SCENE"
		}
	} else {
		if opt.ExpireSeconds <= 0 {
			return "", nil, CodeUnknown, fmt.Errorf("expire_seconds is required for temporary qrcode")
		}
		if opt.ExpireSeconds > maxExpireSeconds {
			return "", nil, CodeUnknown, fmt.Errorf("expire_seconds must be <= %d", maxExpireSeconds)
		}
		if isStringScene {
			actionName = "QR_STR_SCENE"
		} else {
			actionName = "QR_SCENE"
		}
	}

	return actionName, scene, CodeOK, nil
}
