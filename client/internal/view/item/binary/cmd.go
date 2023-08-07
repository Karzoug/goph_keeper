package binary

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
)

const maxValueSizeInDB = 1024 * 1024 // in bytes

func createCmd(c *client.Client, vitem vault.Item, filename string) tea.Cmd {
	return func() tea.Msg {
		fi, err := os.Stat(filename)
		if err != nil {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "Open file problem.",
			}
		}
		if fi.Size() > maxValueSizeInDB {
			vitem.Type = cvault.BinaryLarge
			// TODO: implement me
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "Not implemented now, sorry.",
			}
		} else {
			vitem.Type = cvault.Binary
		}
		if fi.IsDir() {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "You choose a directory, not a file.",
			}
		}

		value, err := os.ReadFile(filename)
		if err != nil {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "Open file problem.",
			}
		}

		err = c.EncryptAndSetVaultItem(context.TODO(), vitem, vault.Binary{Value: value})
		if err != nil {
			switch {
			case errors.Is(err, client.ErrConflictVersion):
				return item.ConflictVersionSetItemMsg{}
			case errors.Is(err, client.ErrAppInternal) || errors.Is(err, client.ErrUserNeedAuthentication):
				return vc.ErrMsg{
					Time: time.Now(),
					Err:  err.Error(),
				}
			}
		}
		return item.SuccessfulSetItemMsg{}
	}
}

func saveOnDiskCmd(c *client.Client, vitem vault.Item, value vault.Binary, dirname, filename string) tea.Cmd {
	return func() tea.Msg {
		fi, err := os.Stat(dirname)
		if err != nil {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "Open directory problem: ",
			}
		}
		if !fi.IsDir() {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "You choose a file, not a directory.",
			}
		}

		f, err := os.Create(filepath.Join(dirname, filename))
		if err != nil {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "Create file problem",
			}
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		_, err = w.Write(value.Value)
		if err != nil {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  "Write file problem",
			}
		}
		defer w.Flush()

		return item.SuccessfulSetItemMsg{}
	}
}
