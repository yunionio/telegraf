package procutils

import "github.com/pkg/errors"

func IsRemoteFileExist(path string) (bool, error) {
	out, err := NewRemoteCommandAsFarAsPossible("ls", path).Output()
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if %s exists: %s", path, out)
	}
	return true, nil
}
