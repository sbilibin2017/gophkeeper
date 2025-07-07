package buildinfo

import "fmt"

// BuildInfo содержит сведения о сборке программного обеспечения.
type BuildInfo struct {
	Platform string // Платформа сборки (например, linux/amd64)
	Version  string // Версия сборки (например, v1.2.3)
	Date     string // Дата сборки (например, 2025-07-06)
	Commit   string // Хеш коммита сборки
}

// NewBuildInfo создает новый экземпляр BuildInfo, применяя переданные опции.
// Если какая-либо опция не задана, используются значения по умолчанию "N/A".
func NewBuildInfo(opts ...Opt) *BuildInfo {
	b := &BuildInfo{
		Platform: "N/A",
		Version:  "N/A",
		Date:     "N/A",
		Commit:   "N/A",
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Opt определяет функциональную опцию для настройки BuildInfo.
type Opt func(*BuildInfo)

// WithPlatform задает платформу сборки.
func WithPlatform(platform string) Opt {
	return func(b *BuildInfo) { b.Platform = platform }
}

// WithVersion задает версию сборки.
func WithVersion(version string) Opt {
	return func(b *BuildInfo) { b.Version = version }
}

// WithDate задает дату сборки.
func WithDate(date string) Opt {
	return func(b *BuildInfo) { b.Date = date }
}

// WithCommit задает хеш коммита сборки.
func WithCommit(commit string) Opt {
	return func(b *BuildInfo) { b.Commit = commit }
}

// String возвращает отформатированную строку с информацией о сборке.
// Реализует интерфейс fmt.Stringer.
func (b *BuildInfo) String() string {
	return fmt.Sprintf(
		"Build platform: %s\nBuild version: %s\nBuild date: %s\nBuild commit: %s",
		b.Platform, b.Version, b.Date, b.Commit,
	)
}
