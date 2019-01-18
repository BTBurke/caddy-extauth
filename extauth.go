package extauth

import "net/http"

func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	return http.StatusOK, nil
}
