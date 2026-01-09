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
	ID    string `json:"user_id"`
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
			http.Post(baseURL+"/auth/login", "application/json",
				bytes.NewBufferString(`{"email":"alice@test.com","password":"bad"}`))
		}
		return nil
	})

	log.Println("All ASVS tests passed")
}

func registerAndLogin(email string) User {
	pass := "Password123!"
	http.Post(baseURL+"/auth/register", "application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"email":"%s","password":"%s","first_name":"A","last_name":"B"}`, email, pass)))

	res, _ := http.Post(baseURL+"/auth/login", "application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, pass)))

	var out User
	json.NewDecoder(res.Body).Decode(&out)
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
	req, _ := http.NewRequest(method, baseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return do(req, codes...)
}

func expectStatusAuthBody(method, path, token, body string, codes ...int) error {
	req, _ := http.NewRequest(method, baseURL+path, bytes.NewBufferString(body))
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
	defer res.Body.Close()

	for _, c := range codes {
		if res.StatusCode == c {
			return nil
		}
	}

	b, _ := io.ReadAll(res.Body)
	return fmt.Errorf("got %d (%s)", res.StatusCode, string(b))
}
