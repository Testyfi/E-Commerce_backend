package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testify/internal/models"
	"testify/internal/payment/phonepay/payrequest"
	httpClient "testify/internal/utility/http"
	"time"
)

func GetPaymentRequestUrl(w http.ResponseWriter, r *http.Request) {

	var p struct {
		Amount float64 `json:"amount"`
	}

	err := json.NewDecoder(r.Body).Decode(&p)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, "Please send a valid amount", err)
		return
	}

	user, ok := r.Context().Value(models.ContextUser).(models.User)

	if !ok {
		httpClient.RespondError(w, http.StatusBadRequest, "Failed to retrieve user", fmt.Errorf("failed to retrieve user"))
		return
	}
   red:="https://testtify.com/checkoutpage"
   callback:="https://webhook.site/369b0c9d-2c52-4db4-8a26-9a942c882990"
	payload := payrequest.TransactionRequest{
		UID:           user.User_id,
		Amount:        p.Amount * 100,
		MobileNumber:  *user.Phone,
		TransactionID: fmt.Sprintf("ph#%s%s%d", user.User_id[:5], *user.Phone, time.Now().Unix()),
		RedirectURL: &red,
		CallbackURL: &callback,
	}

	transaction, err := payrequest.PayRequest(payload)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	httpClient.RespondSuccess(w, transaction)
}
