package binary

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const maxValueSizeInDB = 1024 * 1024 // in bytes

func (v *View) createCmd(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return e.Wrap("open file problem", err)
	}
	if fi.Size() > maxValueSizeInDB {
		v.item.Type = cvault.BinaryLarge
		// TODO: implement me
		return errors.New("not implemented now, sorry")
	} else {
		v.item.Type = cvault.Binary
	}
	if fi.IsDir() {
		return errors.New("you choose a directory, not a file")
	}

	value, err := os.ReadFile(path)
	if err != nil {
		return e.Wrap("open file problem", err)
	}

	ctx, cancel := context.WithTimeout(v.baseContext, vc.StandartTimeout)
	defer cancel()

	err = v.client.EncryptAndSetVaultItem(ctx, v.item, vault.Binary{Value: value})
	if err != nil {
		return err
	}
	return nil
}

func (v *View) saveOnDiskCmd(path, filename string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return e.Wrap("open directory problem", err)
	}
	if !fi.IsDir() {
		path = filepath.Dir(path)
		if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
			return errors.New("you choose a file, not a directory")
		}
	}

	f, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return e.Wrap("create file problem", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.Write(v.value.Value)
	if err != nil {
		return e.Wrap("Write file problem", err)
	}
	defer w.Flush()

	return nil
}
