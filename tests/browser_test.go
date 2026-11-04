package tests

import (
	"encoding/json"
	"github.com/AlexGalhardo/vukem-email/app/controllers"
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

// This fixture exists only in go test; production has no authentication bypass.
func TestBrowserSmoke(t *testing.T) {
	if os.Getenv("RUN_BROWSER_TESTS") != "1" {
		t.Skip("set RUN_BROWSER_TESTS=1 and install Chromium to run browser tests")
	}
	app, p := setup(t)
	db, _ := app.DB()
	for _, s := range []models.Subscriber{{ProjectID: p.ID, Email: "ana@example.com", Status: "confirmed"}, {ProjectID: p.ID, Email: "bruno@example.com", Status: "pending"}, {ProjectID: p.ID, Email: "carla@example.com", Status: "confirmed"}} {
		if err := orm.Create(t.Context(), db, &s); err != nil {
			t.Fatal(err)
		}
	}
	h := app.Handler()
	server := httptest.NewServer(h)
	defer server.Close()
	sessions, err := app.Sessions()
	if err != nil {
		t.Fatal(err)
	}
	cookies := sessions.Cookies(t.Context(), map[string]any{"admin": controllers.AdminEmail})
	type browserCookie struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		URL   string `json:"url"`
	}
	var data []browserCookie
	for _, c := range cookies {
		data = append(data, browserCookie{Name: c.Name, Value: c.Value, URL: server.URL})
	}
	payload, _ := json.Marshal(data)
	cmd := exec.CommandContext(t.Context(), "node", "scripts/browser-smoke.mjs")
	cmd.Env = append(os.Environ(), "BROWSER_URL="+server.URL, "BROWSER_COOKIES="+string(payload))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("browser smoke failed: %v\n%s", err, out)
	}
	t.Log(string(out))
}
