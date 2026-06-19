package httpresponse

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

func JSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}

func JSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message: message,
	})
}

func JSONAppError(c *gin.Context, err error) {
	statusCode := StatusCodeFromError(err)
	if statusCode >= http.StatusInternalServerError {
		JSONError(c, statusCode, "Erro interno do servidor")
		return
	}

	JSONError(c, statusCode, err.Error())
}

func StatusCodeFromError(err error) int {
	return apperror.StatusCode(err)
}

func AbortJSONError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{
		Message: message,
	})
}

func BindAndValidateJSON(c *gin.Context, validate *validator.Validate, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		JSONError(c, http.StatusBadRequest, "Corpo da requisicao invalido")
		return false
	}

	if validate == nil {
		validate = validator.New()
	}

	if err := validate.Struct(req); err != nil {
		JSONValidationError(c, err, req)
		return false
	}

	return true
}

func JSONValidationError(c *gin.Context, err error, req interface{}) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		JSONError(c, http.StatusBadRequest, "Validacao falhou")
		return
	}

	fieldNames := jsonFieldNames(req)
	details := make([]ErrorDetail, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		fieldName := fieldNames[validationErr.Field()]
		if fieldName == "" {
			fieldName = strings.ToLower(validationErr.Field())
		}

		details = append(details, ErrorDetail{
			Field:   fieldName,
			Message: validationMessage(validationErr, fieldNames),
		})
	}

	c.JSON(http.StatusBadRequest, ErrorResponse{
		Message: "Validacao falhou",
		Errors:  details,
	})
}

func jsonFieldNames(req interface{}) map[string]string {
	reqType := reflect.TypeOf(req)
	if reqType == nil {
		return map[string]string{}
	}

	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	if reqType.Kind() != reflect.Struct {
		return map[string]string{}
	}

	fieldNames := make(map[string]string, reqType.NumField())
	for index := 0; index < reqType.NumField(); index++ {
		field := reqType.Field(index)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldNames[field.Name] = strings.Split(jsonTag, ",")[0]
	}

	return fieldNames
}

func validationMessage(fieldError validator.FieldError, fieldNames map[string]string) string {
	switch fieldError.Tag() {
	case "required":
		return "e obrigatorio"
	case "email":
		return "deve ser um e-mail valido"
	case "min":
		return fmt.Sprintf("deve ter pelo menos %s caracteres", fieldError.Param())
	case "max":
		return fmt.Sprintf("deve ter no maximo %s caracteres", fieldError.Param())
	case "oneof":
		return fmt.Sprintf("deve ser um dos valores: %s", fieldError.Param())
	case "eqfield":
		targetField := fieldNames[fieldError.Param()]
		if targetField == "" {
			targetField = strings.ToLower(fieldError.Param())
		}

		return fmt.Sprintf("deve ser igual a %s", targetField)
	default:
		return "e invalido"
	}
}
