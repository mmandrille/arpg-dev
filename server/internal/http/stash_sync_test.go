package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type recordingStashNotifier struct{ accounts []string }

func (r *recordingStashNotifier) NotifyAccountStashChanged(accountIDs ...string) {
	r.accounts = append(r.accounts, accountIDs...)
}

func serveWithStashSync(t *testing.T, status int, extra ...string) []string {
	t.Helper()
	notifier := &recordingStashNotifier{}
	s := &Server{stash: notifier}
	h := s.withStashSync(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.notifyStashChanged(extra...)
		w.WriteHeader(status)
	}))
	req := httptest.NewRequest(http.MethodPost, "/v0/market/listings", nil)
	req = req.WithContext(withAccount(req.Context(), "acct_caller"))
	h.ServeHTTP(httptest.NewRecorder(), req)
	return notifier.accounts
}

func TestWithStashSyncNotifiesCallerOnSuccess(t *testing.T) {
	got := serveWithStashSync(t, http.StatusOK)
	if len(got) != 1 || got[0] != "acct_caller" {
		t.Fatalf("notified %v, want [acct_caller]", got)
	}
}

func TestWithStashSyncSkipsCallerOnFailure(t *testing.T) {
	if got := serveWithStashSync(t, http.StatusConflict); len(got) != 0 {
		t.Fatalf("4xx response notified %v", got)
	}
}

func TestNotifyStashChangedWithoutHubIsNoop(t *testing.T) {
	s := &Server{stash: stashNotifierFor(nil)}
	if s.stash != nil {
		t.Fatal("a nil hub must not become a typed-nil stash notifier")
	}
	s.notifyStashChanged("acct") // must not panic
}
