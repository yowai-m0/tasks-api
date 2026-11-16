package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostHandlerBadJson(t *testing.T) {
    srv := &Server{}
    body := fmt.Sprintln(`{"title":}`)
    req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
    rec := httptest.NewRecorder()

    srv.postHandler(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %d", want, rec.Code)
    if rec.Code != want {
        t.Errorf("ожидали %d, получили %d", want, rec.Code)
    }
}

func TestPostHandlerLongTitle(t *testing.T) {
	srv := &Server{}
    longWord := strings.Repeat("a", 101)
    body := fmt.Sprintf(`{"title":"%s"}`, longWord)
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	rec := httptest.NewRecorder()

	srv.postHandler(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %d", want, rec.Code)
	if rec.Code != want {
		t.Errorf("ожидали %d, получили %d", want, rec.Code)
	}
}

func TestPostHandlerIsEmpty(t *testing.T) {
    srv := &Server{}
    req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":""}`))
    rec := httptest.NewRecorder()

    srv.postHandler(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %d", want, rec.Code)
    if rec.Code != want {
        t.Errorf("ожидали %d, получили %d", want, rec.Code)
    }
}

func TestGetHandlerPath(t *testing.T) {
    srv := &Server{}
    req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
    req.SetPathValue("id", "abc")
    rec := httptest.NewRecorder()
    
    srv.getHandlerPath(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %d", want, rec.Code)
    if rec.Code != want {
        t.Errorf("ожидали %d, получили %d", want, rec.Code)
    }
}

func TestDeleteHandler(t *testing.T) {
    srv := &Server{}
    req := httptest.NewRequest(http.MethodDelete, "/tasks/abc", nil)
    req.SetPathValue("id", "abc")
    rec := httptest.NewRecorder()
    
    srv.deleteHandler(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %d", want, rec.Code)
    if rec.Code != want {
        t.Errorf("ожидали %d, получили %d", want, rec.Code)
    }
}

func TestPatchHandler(t *testing.T) {
    srv := &Server{}
    req := httptest.NewRequest(http.MethodPatch, "/tasks/abc", nil)
    req.SetPathValue("id", "abc")
    rec := httptest.NewRecorder()
    
    srv.patchHandler(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %d", want, rec.Code)
    if rec.Code != want {
        t.Errorf("ожидали %d, получили %d", want, rec.Code)
    }
}

func TestPatchHandlerBadJson(t *testing.T) {
    srv := &Server{}
    body := strings.NewReader(`{"title"}`)
    req := httptest.NewRequest(http.MethodPatch, "/tasks/1", body)
    req.SetPathValue("id", "1")
    rec := httptest.NewRecorder()

    srv.patchHandler(rec, req)

    want := http.StatusBadRequest
    t.Logf("ожидали %d, получили %s", want, rec.Body.String())
    if rec.Code != want {
        t.Errorf("ожидали %d, получили %d", want, rec.Code)
    }
}

func TestGetID(t *testing.T) {
    tests := []struct{
        name string
        in string
        want int
        wantErr bool
    }{
        {"положительное", "1", 1, false},
        {"не число", "abc", 0, true},
        {"пустой ID", "", 0, true},
        {"не целый ID", "1.5", 0, true},
        {"отрицательный ID", "-5", 0, true},
        {"положительный результат", "42", 42, false},
        {"нулевой результат", "0", 0, true},
    }

    for _, test := range tests{
        req := httptest.NewRequest(http.MethodGet, "/tasks/x", nil)
        req.SetPathValue("id", test.in)

        id, err := getID(req)
        gotErr := err != nil
        if gotErr != test.wantErr {
            t.Errorf("%s: ошибка: ожидали %t, получили %t", test.name, test.wantErr, gotErr)
        }

        if !test.wantErr && id != test.want {
            t.Errorf("%s: ожидали %d, получили %d", test.name, test.want, id)
        }
    }
}