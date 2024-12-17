package common

import (
	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type ResultStatus struct {
	ResultCode     string      `json:"resultCode"`
	ResultMessage  string      `json:"resultMessage"`
	HttpStatusCode int         `json:"httpStatusCode"`
	DetailMessage  string      `json:"detailMessage"`
	Items          interface{} `json:"items"`
}

type ListResultStatus struct {
	ResultCode     string      `json:"resultCode"`
	ResultMessage  string      `json:"resultMessage"`
	HttpStatusCode int         `json:"httpStatusCode"`
	DetailMessage  string      `json:"detailMessage"`
	ItemMetaData   ListCount   `json:"itemMetaData"`
	Items          interface{} `json:"items"`
}

type ListCount struct {
	AllItemCount       int `json:"allItemCount"`
	RemainingItemCount int `json:"remainingItemCount"`
}

func RespOK(c *fiber.Ctx, data interface{}) error {
	resultStatus := ResultStatus{
		ResultCode:     RESULT_STATUS_SUCCESS,
		ResultMessage:  localize(c, "OK"),
		HttpStatusCode: fiber.StatusOK,
		DetailMessage:  localize(c, "OK"),
		Items:          data,
	}
	return c.Status(fiber.StatusOK).JSON(resultStatus)

}

func ListRespOK(c *fiber.Ctx, listCount ListCount, data interface{}) error {
	listResultStatus := ListResultStatus{
		ResultCode:     RESULT_STATUS_SUCCESS,
		ResultMessage:  localize(c, "OK"),
		HttpStatusCode: fiber.StatusOK,
		DetailMessage:  localize(c, "OK"),
		ItemMetaData:   listCount,
		Items:          data,
	}
	return c.Status(fiber.StatusOK).JSON(listResultStatus)

}

func RespErr(c *fiber.Ctx, err error) error {
	log.Errorf("[RespErr Reason]: %s", err.Error())
	resultStatus := ResultStatus{
		ResultCode:     RESULT_STATUS_FAIL,
		ResultMessage:  localize(c, err.Error()),
		HttpStatusCode: fiber.StatusBadRequest,
		DetailMessage:  localize(c, err.Error()),
		Items:          make([]string, 0),
	}
	return c.Status(200).JSON(resultStatus)
}

func RespErrStatus(c *fiber.Ctx, statusCode int, err error) error {
	log.Errorf("[RespErr Reason]: %s", err.Error())
	resultStatus := ResultStatus{
		ResultCode:     RESULT_STATUS_FAIL,
		ResultMessage:  err.Error(),
		HttpStatusCode: statusCode,
		DetailMessage:  localize(c, err.Error()),
		Items:          make([]string, 0),
	}
	return c.Status(statusCode).JSON(resultStatus)
}
func localize(c *fiber.Ctx, msg string) string {
	localize_msg, err := fiberi18n.Localize(c, msg)
	if err != nil {
		return msg
	}
	return localize_msg
}
