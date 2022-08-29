package compare

import (
	"bytes"
	"image"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpawnMap(t *testing.T) {
	testCase := []struct {
		filename string
	}{
		{
			filename: "invalid_1.jpg",
		},
		{
			filename: "invalid_2.jpg",
		},
		{
			filename: "invalid_3.jpg",
		},
		{
			filename: "invalid_4.jpg",
		},
		{
			filename: "overlay_map_1.jpg",
		},
		{
			filename: "overlay_map_2.jpg",
		},
		{
			filename: "spawn_map_1.jpg",
		},
		{
			filename: "spawn_map_2.jpg",
		},
		{
			filename: "spawn_map_3.jpg",
		},
		{
			filename: "spawn_map_4.jpg",
		},
	}
	for _, c := range testCase {
		t.Run(c.filename, func(t *testing.T) {
			path := getPath(c.filename)
			data, err := ioutil.ReadFile(path)
			require.NoError(t, err)
			img, _, err := image.Decode(bytes.NewReader(data))
			require.NoError(t, err)
			isValid, score, err := IsValidImage(img, SpawnMap)
			require.NoError(t, err)
			if !strings.Contains(c.filename, "spawn") {
				require.False(t, isValid, score)
				return
			}
			require.True(t, isValid, score)
		})
	}
}
