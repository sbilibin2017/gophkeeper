package services

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// ClientInstallHTTP скачивает бинарный клиент с HTTP-сервера и сохраняет его в файл.
// Имя файла формируется на основе текущей ОС и архитектуры.
// После успешного скачивания устанавливает права на выполнение (кроме Windows).
// Возвращает ошибку, если возникли проблемы с загрузкой или сохранением файла.
func ClientInstallHTTP(ctx context.Context, httpClient *resty.Client) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	fileName, err := generateClientBinaryFileName(goos, goarch)
	if err != nil {
		return fmt.Errorf("не удалось определить имя файла клиента: %w", err)
	}

	platform := fmt.Sprintf("%s-%s", goos, goarch)

	resp, err := httpClient.R().
		SetContext(ctx).
		SetOutput(fileName).
		Get(fmt.Sprintf("/clients/%s", platform))
	if err != nil {
		return fmt.Errorf("ошибка при загрузке клиента по HTTP: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("получен ошибочный HTTP статус: %d", resp.StatusCode())
	}

	if goos != "windows" {
		if err := os.Chmod(fileName, 0755); err != nil {
			return fmt.Errorf("не удалось установить права на выполнение файла: %w", err)
		}
	}

	return nil
}

// ClientInstallGRPC скачивает бинарный клиент с gRPC-сервера и сохраняет его в файл.
// Имя файла формируется на основе текущей ОС и архитектуры.
// Возвращает ошибку, если сервер вернул ошибку, данные пустые, или не удалось сохранить файл.
func ClientInstallGRPC(ctx context.Context, grpcClient pb.ClientInstallServiceClient) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	fileName, err := generateClientBinaryFileName(goos, goarch)
	if err != nil {
		return fmt.Errorf("не удалось определить имя файла клиента: %w", err)
	}

	req := &pb.InstallRequest{
		Os:   goos,
		Arch: goarch,
	}

	resp, err := grpcClient.DownloadClient(ctx, req)
	if err != nil {
		return fmt.Errorf("ошибка вызова gRPC метода: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("ошибка сервера: %s", resp.Error)
	}

	if len(resp.BinaryData) == 0 {
		return fmt.Errorf("получены пустые бинарные данные")
	}

	err = os.WriteFile(fileName, resp.BinaryData, 0755)
	if err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}

	return nil
}

// generateClientBinaryFileName возвращает имя файла бинарного клиента
// для указанной операционной системы и архитектуры.
// Поддерживаются платформы: windows-amd64, linux-amd64, darwin-amd64, darwin-arm64.
// В случае неподдерживаемой платформы возвращает ошибку.
func generateClientBinaryFileName(goos, goarch string) (string, error) {
	platform := fmt.Sprintf("%s-%s", goos, goarch)

	switch platform {
	case "windows-amd64":
		return "client-windows-amd64.exe", nil
	case "linux-amd64":
		return "client-linux-amd64", nil
	case "darwin-amd64":
		return "client-darwin-amd64", nil
	case "darwin-arm64":
		return "client-darwin-arm64", nil
	default:
		return "", fmt.Errorf("неподдерживаемая платформа: %s", platform)
	}
}
