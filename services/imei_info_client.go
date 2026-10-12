package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
)

type ImeiInfoLookupResult struct {
	Brand          string
	TechnicalModel string
	CommercialName string
	Raw            map[string]interface{}
}

type IImeiInfoClient interface {
	Lookup(ctx context.Context, imei string) (*ImeiInfoLookupResult, error)
}

type imeiInfoClient struct {
	httpClient *http.Client
	baseURL    string
	serviceID  string
	apiKey     string
}

func NewImeiInfoClient() IImeiInfoClient {
	timeout := 15 * time.Second
	configured := strings.TrimSpace(utils.EnvOrDefault("IMEI_INFO_SERVICE", utils.BasicImeiCheckServiceID))
	serviceID := utils.ResolveImeiInfoServiceID()
	if configured != "" && configured != utils.BasicImeiCheckServiceID {
		log.Printf(
			"imei.info: IMEI_INFO_SERVICE=%q is not allowed; using Basic IMEI Check (service %s) only",
			configured,
			utils.BasicImeiCheckServiceID,
		)
	}
	return &imeiInfoClient{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    strings.TrimRight(utils.EnvOrDefault("IMEI_INFO_BASE_URL", "https://dash.imei.info/api"), "/"),
		serviceID:  serviceID,
		apiKey:     utils.EnvOrDefault("IMEI_INFO_API_KEY", ""),
	}
}

func (c *imeiInfoClient) Lookup(ctx context.Context, imei string) (*ImeiInfoLookupResult, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("imei.info API key is not configured")
	}
	if strings.TrimSpace(c.serviceID) == "" {
		return nil, fmt.Errorf("imei.info service id is not configured")
	}

	endpoint := fmt.Sprintf("%s/check/%s/", c.baseURL, url.PathEscape(c.serviceID))
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := reqURL.Query()
	q.Set("imei", imei)
	q.Set("API_KEY", c.apiKey)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("imei.info returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("imei.info invalid json: %w", err)
	}

	brand, tech, commercial := parseImeiInfoPayload(payload)
	if brand == "" && tech == "" && commercial == "" {
		return nil, fmt.Errorf("imei.info returned no device data")
	}

	return &ImeiInfoLookupResult{
		Brand:          brand,
		TechnicalModel: tech,
		CommercialName: commercial,
		Raw:            payload,
	}, nil
}

func parseImeiInfoPayload(payload map[string]interface{}) (brand, technicalModel, commercialName string) {
	brand = firstString(payload,
		"brand", "manufacturer", "brand_name", "maker")
	technicalModel = firstString(payload,
		"model", "technical_model", "model_name", "name")
	commercialName = firstString(payload,
		"commercial_name", "marketing_name", "device_name", "model_name", "name")

	// dash.imei.info async check response: { "status": "Done", "result": { ... } }
	if result, ok := payload["result"].(map[string]interface{}); ok {
		if brand == "" {
			brand = firstString(result, "brand_name", "brand", "manufacturer")
		}
		if technicalModel == "" {
			technicalModel = firstString(result, "technical_model", "model_code", "code_name")
		}
		if commercialName == "" {
			commercialName = firstString(result,
				"model", "commercial_name", "marketing_name", "model_name",
				"model_description", "full_name")
		}
		if brand == "" {
			brand = firstString(result, "manufacturer")
		}
		if technicalModel == "" {
			technicalModel = firstString(result, "code_name", "full_name")
		}
	}

	if data, ok := payload["data"].(map[string]interface{}); ok {
		if brand == "" {
			brand = firstString(data, "brand", "manufacturer", "brand_name")
		}
		if technicalModel == "" {
			technicalModel = firstString(data, "model", "technical_model", "model_name")
		}
		if commercialName == "" {
			commercialName = firstString(data, "commercial_name", "marketing_name", "device_name")
		}
	}
	if commercialName == "" {
		commercialName = technicalModel
	}
	return brand, technicalModel, commercialName
}

func firstString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
