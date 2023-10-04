package util

import (
    "fmt"
    "io"
    "net/http"

    "github.com/rs/zerolog/log"
)

// Returns: (body, error)
func Get(url string) (string, error) {
    body, _, err := GetWithHeaders(url, make(map[string][]string));
    return body, err;
}

// Returns: (body, headers (response), error)
func GetWithHeaders(url string, headers map[string][]string) (string, map[string][]string, error) {
    request, err := http.NewRequest("GET", url, nil);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to create GET request on URL '%s': '%w'.", url, err);
    }

    for key, values := range headers {
        for _, value := range values {
            request.Header.Add(key, value);
        }
    }

    client := http.Client{}

    response, err := client.Do(request);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to perform GET request on URL '%s': '%w'.", url, err);
    }
    defer response.Body.Close();

    rawBody, err := io.ReadAll(response.Body);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to read body from GET on URK '%s': '%w'.", url, err);
    }
    body := string(rawBody);

    if (response.StatusCode != http.StatusOK) {
        log.Error().Int("code", response.StatusCode).Str("body", body).Str("url", url).Msg("Got a non-OK status.");
        return "", nil, fmt.Errorf("Got a non-OK status code '%d' from GET on URK '%s': '%w'.", response.StatusCode, url, err);
    }

    return body, response.Header, nil;
}
