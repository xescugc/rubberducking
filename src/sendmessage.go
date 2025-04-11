package src

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/afero"
)

func SendMessage(ctx context.Context, fs afero.Fs, msg string) error {
	err := initFs(fs)
	if err != nil {
		return fmt.Errorf("could not initialize FS")
	}

	fi, err := fs.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("there is no Duck running!")
		}
		return err
	}
	defer fi.Close()

	var data Data
	err = json.NewDecoder(fi).Decode(&data)
	if err != nil {
		return err
	}

	cmr := CreateMessageRequest{Message: msg}
	b, err := json.Marshal(&cmr)
	if err != nil {
		return err
	}

	err = postToGame(data.ManagerURL+"/messages", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	return nil
}
