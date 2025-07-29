package usecases

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

// ClientSaver defines an interface for saving secrets.
type ClientSaver interface {
	Save(ctx context.Context, secret *models.Secret) error
}

// Encryptor defines an interface for encrypting data.
type Encryptor interface {
	Encrypt(plaintext []byte) (*models.SecretEncrypted, error)
}

// -------------------- Bankcard --------------------

type ClientBankcardAddUsecase struct {
	saver     ClientSaver
	encryptor Encryptor
}

func NewClientBankcardAddUsecase(saver ClientSaver, encryptor Encryptor) *ClientBankcardAddUsecase {
	return &ClientBankcardAddUsecase{saver: saver, encryptor: encryptor}
}

func (a *ClientBankcardAddUsecase) Execute(
	ctx context.Context,
	token string,
	secretName string,
	bankcard *models.Bankcard,
) error {
	if bankcard == nil {
		return errors.New("request is nil")
	}

	plaintext, err := json.Marshal(bankcard)
	if err != nil {
		return err
	}

	encrypted, err := a.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.Secret{
		SecretName:  secretName,
		SecretType:  models.SecretTypeBankCard,
		SecretOwner: token,
		Ciphertext:  encrypted.Ciphertext,
		AESKeyEnc:   encrypted.AESKeyEnc,
	}

	return a.saver.Save(ctx, secret)
}

// -------------------- Text --------------------

type ClientTextAddUsecase struct {
	saver     ClientSaver
	encryptor Encryptor
}

func NewClientTextAddUsecase(saver ClientSaver, encryptor Encryptor) *ClientTextAddUsecase {
	return &ClientTextAddUsecase{saver: saver, encryptor: encryptor}
}

func (a *ClientTextAddUsecase) Execute(
	ctx context.Context,
	token string,
	secretName string,
	text *models.Text,
) error {
	if text == nil {
		return errors.New("request is nil")
	}

	plaintext, err := json.Marshal(text)
	if err != nil {
		return err
	}

	encrypted, err := a.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.Secret{
		SecretName:  secretName,
		SecretType:  models.SecretTypeText,
		SecretOwner: token,
		Ciphertext:  encrypted.Ciphertext,
		AESKeyEnc:   encrypted.AESKeyEnc,
	}

	return a.saver.Save(ctx, secret)
}

// -------------------- Binary --------------------

type ClientBinaryAddUsecase struct {
	saver     ClientSaver
	encryptor Encryptor
}

func NewClientBinaryAddUsecase(saver ClientSaver, encryptor Encryptor) *ClientBinaryAddUsecase {
	return &ClientBinaryAddUsecase{saver: saver, encryptor: encryptor}
}

func (a *ClientBinaryAddUsecase) Execute(
	ctx context.Context,
	token string,
	secretName string,
	binary *models.Binary,
) error {
	if binary == nil {
		return errors.New("request is nil")
	}

	plaintext, err := json.Marshal(binary)
	if err != nil {
		return err
	}

	encrypted, err := a.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.Secret{
		SecretName:  secretName,
		SecretType:  models.SecretTypeBinary,
		SecretOwner: token,
		Ciphertext:  encrypted.Ciphertext,
		AESKeyEnc:   encrypted.AESKeyEnc,
	}

	return a.saver.Save(ctx, secret)
}

// -------------------- User --------------------

type ClientUserAddUsecase struct {
	saver     ClientSaver
	encryptor Encryptor
}

func NewClientUserAddUsecase(saver ClientSaver, encryptor Encryptor) *ClientUserAddUsecase {
	return &ClientUserAddUsecase{saver: saver, encryptor: encryptor}
}

func (a *ClientUserAddUsecase) Execute(
	ctx context.Context,
	token string,
	secretName string,
	user *models.User,
) error {
	if user == nil {
		return errors.New("request is nil")
	}

	plaintext, err := json.Marshal(user)
	if err != nil {
		return err
	}

	encrypted, err := a.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.Secret{
		SecretName:  secretName,
		SecretType:  models.SecretTypeUser,
		SecretOwner: token,
		Ciphertext:  encrypted.Ciphertext,
		AESKeyEnc:   encrypted.AESKeyEnc,
	}

	return a.saver.Save(ctx, secret)
}
