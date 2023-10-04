package util

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"

    "github.com/rs/zerolog/log"
)

// Returns: (body, error)
func Get(uri string) (string, error) {
    body, _, err := GetWithHeaders(uri, make(map[string][]string));
    return body, err;
}

// Returns: (body, headers (response), error)
func GetWithHeaders(uri string, headers map[string][]string) (string, map[string][]string, error) {
    request, err := http.NewRequest("GET", uri, nil);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to create GET request on URL '%s': '%w'.", uri, err);
    }

    for key, values := range headers {
        for _, value := range values {
            request.Header.Add(key, value);
        }
    }

    return doRequest(uri, request, "GET");
}

// Returns: (body, error)
func Post(uri string, form map[string]string) (string, error) {
    body, _, err := PostWithHeaders(uri, form, make(map[string][]string));
    return body, err;
}

// Returns: (body, headers (response), error)
func PostWithHeaders(uri string, form map[string]string, headers map[string][]string) (string, map[string][]string, error) {
    formValues := url.Values{};
    for key, value := range form {
        formValues.Set(key, value);
    }

    request, err := http.NewRequest("POST", uri, strings.NewReader(formValues.Encode()));
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to create POST request on URL '%s': '%w'.", uri, err);
    }

    request.Header.Add("Content-Type", "application/x-www-form-urlencoded");

    for key, values := range headers {
        for _, value := range values {
            request.Header.Add(key, value);
        }
    }

    return doRequest(uri, request, "POST");
}

// Returns: (body, headers (response), error)
func doRequest(uri string, request *http.Request, verb string) (string, map[string][]string, error) {
    client := http.Client{}

    response, err := client.Do(request);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to perform %s request on URL '%s': '%w'.", verb, uri, err);
    }
    defer response.Body.Close();

    rawBody, err := io.ReadAll(response.Body);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to read body from %s on URL '%s': '%w'.", verb, uri, err);
    }
    body := string(rawBody);

    if (response.StatusCode != http.StatusOK) {
        log.Error().Int("code", response.StatusCode).Str("body", body).Str("url", uri).Msg("Got a non-OK status.");
        return "", nil, fmt.Errorf("Got a non-OK status code '%d' from %s on URL '%s': '%w'.", response.StatusCode, verb, uri, err);
    }

    return body, response.Header, nil;
}
