package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iidesho/bragi/sbragi"
	"github.com/ricochet2200/go-disk-usage/du"
)

var KB = uint64(1024)

func main() {
	mounts, err := os.ReadFile("/proc/mounts")
	sbragi.WithError(err).Fatal("reading mounts")
	lines := strings.Split(string(mounts), "\n")
	for _, l := range lines {
		parts := strings.Split(l, " ")
		if len(parts) < 3 || !strings.HasPrefix(parts[2], "ext") && parts[2] != "zfs" {
			continue
		}
		usage := du.NewDiskUsage(parts[1])
		p := usage.Usage() * 100
		fmt.Printf("%s[%s](%s): %.2f%%\n", parts[0], parts[1], parts[2], p)

		if p < 80 {
			continue
		}

		jsonStr := []byte(
			fmt.Sprintf(
				`{"message": "%s[%s]: %.2f%%", "title": "Low disk space"}`,
				parts[0], parts[1], p,
			),
		)
		req, err := http.NewRequest(
			"POST",
			"http://localhost:8123/api/services/notify/mobile_app_sindres_iphone",
			bytes.NewBuffer(jsonStr),
		)
		sbragi.WithError(err).Fatal("creating request")
		req.Header.Set(
			"Authorization",
			fmt.Sprintf("Bearer %s", os.Getenv("token")),
		)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}
}
