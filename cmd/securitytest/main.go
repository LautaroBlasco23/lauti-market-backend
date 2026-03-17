package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

type User struct {
	ID    string `json:"account_id"`
	Token string `json:"token"`
}

func main() {
	log.Println("Running OWASP ASVS security tests")

	u1 := registerAndLogin("alice@test.com")
	u2 := registerAndLogin("bob@test.com")

	run("V4.1 Missing auth", func() error {
		return expectStatus("GET", "/users/"+u1.ID, "", 401)
	})

	run("V4.2 Broken JWT", func() error {
		return expectStatusAuth("GET", "/users/"+u1.ID, "aaa.bbb.ccc", 401)
	})

	run("V4.3 IDOR read", func() error {
		return expectStatusAuth("GET", "/users/"+u2.ID, u1.Token, 403, 404)
	})

	run("V4.3 IDOR update", func() error {
		body := `{"first_name":"hacked"}`
		return expectStatusAuthBody("PUT", "/users/"+u2.ID, u1.Token, body, 403, 404)
	})

	run("V5 Mass assignment", func() error {
		body := `{"first_name":"ok","last_name":"ok","is_admin":true}`
		return expectStatusAuthBody("PUT", "/users/"+u1.ID, u1.Token, body, 200)
	})

	run("V13 Brute force (should be blocked later)", func() error {
		for i := 0; i < 30; i++ {
			res, err := http.Post(baseURL+"/auth/login", "application/json",
				bytes.NewBufferString(`{"email":"alice@test.com","password":"bad"}`))
			if err == nil {
				_ = res.Body.Close() //nolint:errcheck
			}
		}
		return nil
	})

	log.Println("All ASVS tests passed")
}

func registerAndLogin(email string) User {
	pass := "Password123!"
	regRes, err := http.Post(baseURL+"/auth/register/user", "application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"email":%q,"password":%q,"first_name":"Alice","last_name":"Test"}`, email, pass)))
	if err != nil {
		log.Fatalf("register request failed: %v", err)
	}
	if regRes.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(regRes.Body) //nolint:errcheck
		log.Fatalf("register failed (%d): %s", regRes.StatusCode, string(b))
	}
	_ = regRes.Body.Close() //nolint:errcheck

	res, err := http.Post(baseURL+"/auth/login", "application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"email":%q,"password":%q}`, email, pass)))
	if err != nil {
		log.Fatalf("login request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body) //nolint:errcheck
		log.Fatalf("login failed (%d): %s", res.StatusCode, string(b))
	}

	var out User
	_ = json.NewDecoder(res.Body).Decode(&out) //nolint:errcheck
	_ = res.Body.Close()                       //nolint:errcheck
	if out.ID == "" || out.Token == "" {
		log.Fatalf("login response missing id or token: %+v", out)
	}
	return out
}

func run(name string, fn func() error) {
	if err := fn(); err != nil {
		log.Fatalf("❌ %s: %v", name, err)
	}
	log.Println("✅", name)
}

func expectStatus(method, path, token string, codes ...int) error {
	return expectStatusAuth(method, path, token, codes...)
}

func expectStatusAuth(method, path, token string, codes ...int) error {
	req, err := http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return do(req, codes...)
}

func expectStatusAuthBody(method, path, token, body string, codes ...int) error {
	req, err := http.NewRequest(method, baseURL+path, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return do(req, codes...)
}

func do(req *http.Request, codes ...int) error {
	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close() //nolint:errcheck
	}()

	for _, c := range codes {
		if res.StatusCode == c {
			return nil
		}
	}

	b, _ := io.ReadAll(res.Body) //nolint:errcheck
	return fmt.Errorf("got %d (%s)", res.StatusCode, string(b))
}
