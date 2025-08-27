package module

import (
	"os"
	"strings"
)

func RegX() *GasType {
	var configPath = os.Getenv("GasType_CONFIGFILE")
	var keyPath = os.Getenv("GasType_KEYFILE")
	var certPath = os.Getenv("GasType_CERTFILE")
	var printBannerV = os.Getenv("GasType_PRINTBANNER")
	if printBannerV == "" {
		printBannerV = "true"
	}

	return &GasType{
		configPath:  configPath,
		keyPath:     keyPath,
		certPath:    certPath,
		printBanner: strings.ToLower(printBannerV) == "true",
	}
}
