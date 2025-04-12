package dtos

/**
* These structs are used to return
* Consistent responses to the user.
* One for valid responses and the other for errors.
 */

type ValidResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Data    interface{} `json:"data"`
}

type ApiErr struct {
	Status  int    `json:"status"`
	Err     string `json:"error"`
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

func (apr *ApiErr) Error() string {
	return apr.Err
}
