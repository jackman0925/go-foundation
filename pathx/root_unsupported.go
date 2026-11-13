//go:build !darwin && !linux && !windows

package pathx

func openRoot(string) (rootHandle, error) {
	return nil, ErrUnsupportedPlatform
}
