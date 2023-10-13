package utility

import (
	"fmt"
	"net/http"
	"os"
)

func SMS(message string, variables []string, numbers []string) {
	var varValues string = ""
	for i := 0; i < len(variables); i++ {
		varValues += variables[i] + "%7C"
	}
	fmt.Println(message, varValues, numbers)
	URL := fmt.Sprintf("https://www.fast2sms.com/dev/bulkV2?authorization=%s&route=dlt&sender_id=tstify&message=%s&variables_values=%sflash=0&numbers=%s", os.Getenv("SMS_KEY"), message, varValues, numbers[0])
	resp, err := http.Get(URL)
	fmt.Println(URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Println(resp)
}
