package http

import (
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Opt представляет опцию для конфигурации Resty клиента.
type Opt func(*resty.Client)

// New создает новый Resty клиент с указанным базовым URL и опциональной конфигурацией.
//
// Если базовый URL не начинается с "http://" или "https://", автоматически добавляется "http://".
//
// Аргументы:
// - baseURL: базовый URL для HTTP клиента
// - opts: список опций для настройки клиента
//
// Возвращает:
// - *resty.Client: настроенный HTTP клиент
func New(baseURL string, opts ...Opt) *resty.Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	client := resty.New().SetBaseURL(baseURL)
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// RetryPolicy задает политику повторных попыток для HTTP запросов.
type RetryPolicy struct {
	Count   int           // количество повторных попыток
	Wait    time.Duration // время ожидания между попытками
	MaxWait time.Duration // максимальное время ожидания между попытками
}

// WithRetryPolicy возвращает опцию для конфигурации Resty клиента с указанной политикой повторных попыток.
//
// Аргументы:
// - policies: один или несколько объектов RetryPolicy
//
// Настройки включают:
// - Count: количество повторных попыток
// - Wait: базовое время ожидания между попытками
// - MaxWait: максимальное время ожидания между попытками
//
// Если ни одна из настроек не указана, повторные попытки отключаются.
func WithRetryPolicy(policies ...RetryPolicy) Opt {
	return func(c *resty.Client) {
		for _, policy := range policies {
			if policy.Count > 0 || policy.Wait > 0 || policy.MaxWait > 0 {
				if policy.Count > 0 {
					c.SetRetryCount(policy.Count)
				}
				if policy.Wait > 0 {
					c.SetRetryWaitTime(policy.Wait)
				}
				if policy.MaxWait > 0 {
					c.SetRetryMaxWaitTime(policy.MaxWait)
				}
				return
			}
		}

		c.SetRetryCount(0)
		c.SetRetryWaitTime(0)
		c.SetRetryMaxWaitTime(0)
	}
}
