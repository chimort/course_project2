package matching

import (
	"encoding/json"
	"net/http"
)

func (s *Service) HandleAddOnline(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "missing user", http.StatusBadRequest)
		return
	}

	if err := s.AddOnline(user); err != nil {
		http.Error(w, "redis error: "+err.Error(), 500)
		return
	}

	w.Write([]byte("OK"))
}

func (s *Service) HandleListOnline(w http.ResponseWriter, r *http.Request) {
	list, err := s.ListOnline()
	if err != nil {
		http.Error(w, "redis error: "+err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(list)
}
