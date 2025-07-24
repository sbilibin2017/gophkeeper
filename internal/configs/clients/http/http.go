// Пакет http предоставляет обёртку над resty-клиентом
// с возможностью настройки политики повторов и авторизации.
package http

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// Opt определяет функцию, которая настраивает *resty.Client и может вернуть ошибку.
// Используется для модульной конфигурации клиента.
type Opt func(*resty.Client) error

// New создаёт и возвращает новый экземпляр resty.Client с заданным базовым URL и опциями.
// Опции передаются в виде слайса функций типа Opt.
func New(baseURL string, opts ...Opt) (*resty.Client, error) {
	client := resty.New().SetBaseURL(baseURL)

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// RetryPolicy описывает параметры политики повторных попыток HTTP-запросов.
type RetryPolicy struct {
	Count   int           // Количество повторных попыток
	Wait    time.Duration // Время ожидания между попытками
	MaxWait time.Duration // Максимальное время ожидания между попытками
}

// WithRetryPolicy возвращает опцию Opt, которая применяет первую валидную политику повторов из переданных.
// Если ни одна из политик невалидна, опция не делает изменений.
func WithRetryPolicy(policies ...RetryPolicy) Opt {
	return func(c *resty.Client) error {
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
				break
			}
		}
		return nil
	}
}

// WithAuthToken возвращает опцию Opt, которая устанавливает Bearer-токен авторизации
// в заголовки HTTP-запросов. Используется первый непустой токен из переданных.
func WithAuthToken(tokens ...string) Opt {
	return func(c *resty.Client) error {
		for _, token := range tokens {
			if token != "" {
				c.SetAuthToken(token)
				break
			}
		}
		return nil
	}
}
