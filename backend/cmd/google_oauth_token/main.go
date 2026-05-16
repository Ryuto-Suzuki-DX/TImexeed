package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func main() {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))

	if clientID == "" {
		panic("GOOGLE_OAUTH_CLIENT_ID is empty")
	}

	if clientSecret == "" {
		panic("GOOGLE_OAUTH_CLIENT_SECRET is empty")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
		Scopes: []string{
			drive.DriveScope,
		},
	}

	authURL := config.AuthCodeURL(
		"timexeed-google-drive-token",
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)

	fmt.Println("次のURLをブラウザで開いてください。")
	fmt.Println("")
	fmt.Println(authURL)
	fmt.Println("")
	fmt.Println("Googleで許可したあと、http://localhost/?code=... のようなURLに飛びます。")
	fmt.Println("ページ表示に失敗してもOKです。ブラウザのURL欄から code= の値、またはURL全体を貼ってください。")
	fmt.Println("")
	fmt.Print("code またはリダイレクトURLを貼り付け: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	code, err := extractCode(strings.TrimSpace(input))
	if err != nil {
		panic(err)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		panic(err)
	}

	fmt.Println("")
	fmt.Println("取得成功")
	fmt.Println("")
	fmt.Println("GOOGLE_OAUTH_REFRESH_TOKEN=" + token.RefreshToken)
	fmt.Println("")
	fmt.Println("この値を backend/.env に追加してください。")
}

func extractCode(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("code is empty")
	}

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		parsedURL, err := url.Parse(input)
		if err != nil {
			return "", err
		}

		code := strings.TrimSpace(parsedURL.Query().Get("code"))
		if code == "" {
			return "", fmt.Errorf("code query parameter is empty")
		}

		return code, nil
	}

	return input, nil
}
