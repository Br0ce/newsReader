package tsClient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func post(u *url.URL, r io.Reader, t time.Duration) ([]byte, error) {
	req, err := http.NewRequest("POST", u.String(), r)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "text/plain; charset=utf-8")
	req.Header.Add("Accept-Charset", "utf-8")
	c := &http.Client{Timeout: t}

	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("could not post request to url=%v, status=%v", u.String(), response.Status)
	}

	body := response.Body
	defer func(rc io.ReadCloser) {
		err = rc.Close()
		if err != nil {
			fmt.Printf("could not close body, %v\n", err)
		}

	}(body)

	return io.ReadAll(body)
}
