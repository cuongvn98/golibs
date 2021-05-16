package facebook_webhook_event

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strings"
)

type FacebookWebhookEventVerification struct {
	token string
	next  http.Handler
}

func New(token string, next http.Handler) *FacebookWebhookEventVerification {
	return &FacebookWebhookEventVerification{token: token, next: next}
}

type body struct {
}

func (f *FacebookWebhookEventVerification) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headerSignature := r.Header.Get("X-Hub-Signature")
	if headerSignature == "" || f.token == "" {
		f.next.ServeHTTP(w, r)
		return
	}

	strSignature := ""
	if len(headerSignature) == 45 && strings.HasPrefix(headerSignature, "sha1=") {
		strSignature = headerSignature[5:]
	}

	signature, err := hex.DecodeString(strSignature)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(err.Error()))
	}

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	hmc := hmac.New(sha1.New, []byte(f.token))
	hash := hmc.Sum(bs)

	if !hmac.Equal(hash, signature) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request"))
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(bs))
	f.next.ServeHTTP(w, r)
}
