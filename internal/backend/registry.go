package backend

import "fmt"

// ValidateRegistryScheme checks a container CLI --scheme value.
// Apple container 1.3.0 removed `auto`; the default is https when the flag is omitted.
func ValidateRegistryScheme(scheme string) error {
	switch scheme {
	case "", "http", "https":
		return nil
	case "auto":
		return fmt.Errorf("registry scheme %q was removed in container CLI 1.3.0; use http or https", scheme)
	default:
		return fmt.Errorf("invalid registry scheme %q (want http or https)", scheme)
	}
}

// ImagePullArgs builds `container image pull` arguments for a compose service image.
func ImagePullArgs(image, platform, scheme string) ([]string, error) {
	if err := ValidateRegistryScheme(scheme); err != nil {
		return nil, err
	}
	args := []string{"image", "pull"}
	if platform != "" {
		args = append(args, "--platform", platform)
	}
	if scheme != "" {
		args = append(args, "--scheme", scheme)
	}
	return append(args, image), nil
}

// RegistryLoginArgs builds `container registry login` arguments.
func RegistryLoginArgs(server, username, scheme string, passwordStdin bool) ([]string, error) {
	if err := ValidateRegistryScheme(scheme); err != nil {
		return nil, err
	}
	args := []string{"registry", "login"}
	if username != "" {
		args = append(args, "--username", username)
	}
	if passwordStdin {
		args = append(args, "--password-stdin")
	}
	if scheme != "" {
		args = append(args, "--scheme", scheme)
	}
	if server != "" {
		args = append(args, server)
	}
	return args, nil
}
