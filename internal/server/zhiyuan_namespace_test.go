package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHTTPZhiYuanMemoryNamespacesRetainOwnership(t *testing.T) {
	t.Setenv("ENGRAM_HTTP_TOKEN", "synthetic-namespace-test")
	st := newServerTestStore(t)
	handler := New(st, 0).Handler()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer synthetic-namespace-test")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	owners := []string{
		"companion://zhiyuan-agent/pet-a", "companion://zhiyuan-agent/pet-b",
		"personal://zhiyuan-agent/user", "workspace-example",
	}
	ids := make([]int64, len(owners))
	for index, owner := range owners {
		sessionID := fmt.Sprintf("owner-%d", index)
		if err := st.CreateSession(sessionID, owner, t.TempDir()); err != nil {
			t.Fatal(err)
		}
		body, _ := json.Marshal(map[string]string{
			"session_id": sessionID, "project": owner, "scope": "project", "type": "decision",
			"title": "namespace fixture", "content": fmt.Sprintf("数据库连接池 private owner %d", index),
		})
		rec := request(http.MethodPost, "/observations", string(body))
		if rec.Code != http.StatusCreated {
			t.Fatalf("save owner %d: %d %s", index, rec.Code, rec.Body.String())
		}
		var saved struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil || saved.ID == 0 {
			t.Fatalf("save receipt: %s, %v", rec.Body.String(), err)
		}
		ids[index] = saved.ID
	}
	for index, owner := range owners {
		for _, path := range []string{
			"/search?q=" + url.QueryEscape("数据库") + "&project=" + url.QueryEscape(owner) + "&scope=project",
			"/observations/recent?project=" + url.QueryEscape(owner) + "&scope=project",
		} {
			rec := request(http.MethodGet, path, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("lookup owner %d: %d %s", index, rec.Code, rec.Body.String())
			}
			var results []struct {
				ID int64 `json:"id"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil || len(results) != 1 || results[0].ID != ids[index] {
				t.Fatalf("owner %d isolation: %s, %v", index, rec.Body.String(), err)
			}
		}
	}
	if rec := request(http.MethodGet, "/search?q=private&project="+url.QueryEscape("companion://zhiyuan-agent/unknown"), ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown owner must not broaden search: %d %s", rec.Code, rec.Body.String())
	}
	if rec := request(http.MethodGet, "/search?q=private&project="+url.QueryEscape("companion://zhiyuan-agent/pet-a/child"), ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("nested namespace must remain invalid: %d %s", rec.Code, rec.Body.String())
	}
	path := fmt.Sprintf("/observations/%d", ids[0])
	if rec := request(http.MethodDelete, path, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing owner assertion: %d %s", rec.Code, rec.Body.String())
	}
	if rec := request(http.MethodDelete, path+"?expected_project="+url.QueryEscape(owners[1]), ""); rec.Code != http.StatusConflict {
		t.Fatalf("incorrect owner assertion: %d %s", rec.Code, rec.Body.String())
	}
	ownedPath := path + "?expected_project=" + url.QueryEscape(owners[0])
	unauthenticated := httptest.NewRecorder()
	handler.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodDelete, ownedPath, nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("missing authentication: %d", unauthenticated.Code)
	}
	for _, target := range []string{ownedPath, ownedPath + "&hard=true"} {
		if rec := request(http.MethodDelete, target, ""); rec.Code != http.StatusOK {
			t.Fatalf("owned soft-then-hard deletion: %d %s", rec.Code, rec.Body.String())
		}
	}
	if rec := request(http.MethodGet, path, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("deleted content still readable: %d", rec.Code)
	}
	if rec := request(http.MethodGet, fmt.Sprintf("/observations/%d", ids[1]), ""); rec.Code != http.StatusOK {
		t.Fatalf("another owner's observation changed: %d", rec.Code)
	}
}
