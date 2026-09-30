package httpapi

import (
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/realtime"
)

// stashNotifier is the one realtime.Hub method the HTTP layer needs (a seam for tests).
type stashNotifier interface {
	NotifyAccountStashChanged(accountIDs ...string)
}

// stashNotifierFor avoids storing a typed-nil *Hub in the interface.
func stashNotifierFor(hub *realtime.Hub) stashNotifier {
	if hub == nil {
		return nil
	}
	return hub
}

// Live stash sync (v488): routes that write an account stash notify the realtime hub so every live
// session holding a player of that account resyncs its stash on the next tick
// (realtime.Hub.NotifyAccountStashChanged).

// withStashSync notifies the caller's account after a successful (2xx) response. Wrap it inside
// requireAuth so the account is on the request context.
func (s *Server) withStashSync(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.status < 200 || rec.status > 299 {
			return
		}
		if accountID, ok := accountFromContext(r.Context()); ok {
			s.notifyStashChanged(accountID)
		}
	})
}

// notifyStashChanged also covers counterpart accounts a route writes (purchase seller, accepted
// offer bidder). Nil-safe for servers built without a realtime hub.
func (s *Server) notifyStashChanged(accountIDs ...string) {
	if s.stash == nil {
		return
	}
	s.stash.NotifyAccountStashChanged(accountIDs...)
}
