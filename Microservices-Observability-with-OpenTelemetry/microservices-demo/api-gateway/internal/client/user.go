package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"api-gateway/internal/model"
)

func GetUser(ctx context.Context, userID int) model.User {
	req, _ := http.NewRequestWithContext(
		ctx,
		"GET",
		"http://user-service:3001/users/"+strconv.Itoa(userID),
		nil,
	)

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var user model.User
	_ = json.NewDecoder(resp.Body).Decode(&user)
	return user
}
