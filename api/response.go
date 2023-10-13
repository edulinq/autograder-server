package api

import (
)

type APIResponse struct {
    ID string `json:"id"`
    Timestamp string `json:"timestamp"`

    HTTPStatus int `json:"status"`
    Success bool `json:"success"`

    Message string `json:"message"`
    Content any `json:"content"`
}
