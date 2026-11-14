package tests

import (
	"strconv"
	"testing"
)

func TestPreviewReportsRendererOutageAsServerError(t *testing.T) {
	app, p := setup(t)
	body := map[string]string{"project_id": strconv.FormatInt(p.ID, 10), "subject": "S", "mode": "mdx", "source": "export default () => {"}
	admin(t, app).PostJSON("/api/preview", body).AssertStatus(422)

	t.Setenv("PATH", t.TempDir())
	body["source"] = "Hello"
	body["mode"] = "markdown"
	admin(t, app).PostJSON("/api/preview", body).AssertStatus(500)
}
