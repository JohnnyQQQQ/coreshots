package compare

import "fmt"

func getPath(imageName string) string {
	return fmt.Sprintf("./samples/%s", imageName)
}
