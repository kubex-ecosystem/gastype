package cli

import (
	"os"
	"strings"
)

func GetDescriptions(descriptionArg []string, _ bool) map[string]string {
	var description, banner string
	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	} else {
		description = ""
	}

	banner = `   ______          ______               
  / ________ _____/_  ____  ______  ___ 
 / / __/ __ ´/ ___// / / / / / __ \/ _ \
/ /_/ / /_/ (__  )/ / / /_/ / /_/ /  __/
\____/\__,_/____//_/  \__, / .___/\___/
                   /____/_/
`
	return map[string]string{"banner": banner, "description": description}
}
