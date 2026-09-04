package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[int(b)%len(charset)]
	}

	return string(bytes), nil
}

func HttpJsonFromArray[T any](array []T, w http.ResponseWriter) {
	val := reflect.ValueOf(array)

	var modArray []T

	var err error

	if val.Len() == 0 {
		modArray = make([]T, 0)

		err = json.NewEncoder(w).Encode(modArray)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(array)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func HttpJsonFromObject[T any](object T, w http.ResponseWriter) {
	var err error

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func ListFromQueryToResponse[T any](
	query func() ([]T, error),
	r *http.Request,
	w http.ResponseWriter,
) {
	objects, err := query()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	HttpJsonFromArray(objects, w)
}

func ListFromQueryToResponseById[T any](
	query func(string) ([]T, error),
	r *http.Request,
	w http.ResponseWriter,
	id string,
) {
	objects, err := query(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	HttpJsonFromArray(objects, w)
}

func ObjectFromQueryToResponse[T any](
	h func(string) (T, error),
	r *http.Request,
	w http.ResponseWriter,
	id string,
) {
	obj, err := h(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	HttpJsonFromObject(obj, w)
}

func DecodeBody[T any](r *http.Request, w http.ResponseWriter) (T, error) {
	var t T

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return t, err
	}

	return t, nil
}

func Uploader(w http.ResponseWriter, r *http.Request) error {
	// Limit upload size (e.g. 10MB)
	r.ParseMultipartForm(10 << 20) // 10 MB

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)

		return err
	}
	defer file.Close()

	randomString, err := RandomString(8)
	if err != nil {
		return err
	}
	// Create destination file
	ext := filepath.Ext(handler.Filename)
	fileName := fmt.Sprintf("./uploads/%s%s", randomString, ext)

	dst, err := os.Create(fileName)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)

		return err
	}
	defer dst.Close()

	// Copy uploaded content to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)

		return err
	}

	fmt.Fprint(w, strings.Split(fileName, "/")[2])

	return nil
}

func GetUserFromRequest(w http.ResponseWriter, r *http.Request) User {
	AuthCookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)

		return User{}
	}

	tokenString := AuthCookie.Value

	claims := &AuthClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("SECRET")), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)

		return User{}
	}

	user := &User{
		ID:      claims.ID,
		Email:   claims.Email,
		IsAdmin: claims.IsAdmin,
	}

	return *user
}

func ReplacePersianDigits(s string) string {
	persian := []rune("۰۱۲۳۴۵۶۷۸۹")

	english := []rune("0123456789")
	for i, p := range persian {
		s = strings.ReplaceAll(s, string(p), string(english[i]))
	}

	return s
}

func DefaultInput(input string, defaultOutput string) string {
	if input == "" {
		return defaultOutput
	}
	return input
}

func RangeFrom(start, count int) []int {
	nums := make([]int, count)
	for i := range count {
		nums[i] = start + i
	}

	return nums
}

func ValidateBackup(path string) error {
	cmd := exec.Command("tar", "-tzf", path)

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	content := string(out)

	if !strings.Contains(content, "uploads-") {
		return fmt.Errorf("missing uploads archive")
	}

	if !strings.Contains(content, "volume-") {
		return fmt.Errorf("missing volume archive")
	}

	return nil
}

// QueryParams holds the parameters for dynamic queries
type QueryParams struct {
	Sort             string
	SortDirection    string
	Filters          []string
	FilterOperands   []string
	FilterConditions []string
}

// BuildOrderBy generates the ORDER BY clause
func BuildOrderBy(entity string, sort string, sortDirection string) string {
	if sort == "" {
		return ""
	}

	if strings.ToUpper(sortDirection) != "ASC" && strings.ToUpper(sortDirection) != "DESC" {
		sortDirection = "ASC"
	}

	switch sort {
	case "voucher_number":
		return `ORDER BY ` + entity + `.voucher_number::bigint ` + sortDirection
	case "number":
		return `ORDER BY ` + entity + `.number::bigint ` + sortDirection
	case "date":
		return `ORDER BY ` + entity + `.date::date ` + sortDirection
	case "category_name":
		return `ORDER BY categories.name COLLATE "fa-IR-x-icu"` + sortDirection
	case "created_at":
		return `ORDER BY ` + entity + `.created_at::date ` + sortDirection
	case "updated_at":
		return `ORDER BY ` + entity + `.updated_at::date ` + sortDirection
	case "buy_price":
		return `ORDER BY COALESCE(REPLACE(` + entity + `.buy_price,',',''),'')::bigint ` + sortDirection
	case "sell_price":
		return `ORDER BY COALESCE(REPLACE(` + entity + `.sell_price,',',''),'')::bigint ` + sortDirection
	case "code":
		return `ORDER BY COALESCE(` + entity + `.code,'','0')::bigint ` + sortDirection
	case "count":
		return `ORDER BY REPLACE(COALESCE(` + entity + `.count,'0'),',','')::bigint ` + sortDirection
	case "is_service":
		return `ORDER BY COALESCE(` + entity + `.is_service,'FALSE') ` + sortDirection
	case "generated":
		return `ORDER BY COALESCE(` + entity + `.generated,'FALSE') ` + sortDirection
	case "position":
		return `ORDER BY COALESCE(` + entity + `.position,'') ` + sortDirection
	default:
		return fmt.Sprintf(`ORDER BY %s.%s COLLATE "fa-IR-x-icu" %s`, entity, sort, sortDirection)
	}
}

func BuildWhere(
	entitiy string,
	filters, operands, conditions []string,
	startArgIndex int,
) (string, []any) {
	var filterBy strings.Builder

	filterBy.WriteString("WHERE ")

	var args []any

	if len(filters) == 0 {
		return "", args
	}

	argIndex := startArgIndex

	for i, filter := range filters {
		if i >= len(operands) || i >= len(conditions) {
			break
		}

		operand := operands[i]
		condition := conditions[i]

		switch operand {
		case "contains":
			operand = "ILIKE"
			condition = "%" + condition + "%"
		case "notcontains":
			operand = "NOT ILIKE"
			condition = "%" + condition + "%"
		}

		// Add condition value to args array
		args = append(args, condition)

		switch filter {
		case "category_name":
			fmt.Fprintf(&filterBy, `categories.name %s $%d`, operand, argIndex)
		case "is_service":
			fmt.Fprintf(
				&filterBy,
				`COALESCE(%s.is_service,FALSE) %s $%d`,
				entitiy,
				operand,
				argIndex,
			)
		case "number":
			fmt.Fprintf(
				&filterBy,
				`REPLACE(COALESCE(%s.number,'0')::text,',','')::bigint %s $%d::bigint`,
				entitiy,
				operand,
				argIndex,
			)
		case "voucher_number":
			fmt.Fprintf(
				&filterBy,
				`REPLACE(COALESCE(%s.voucher_number,'0')::text,',','')::bigint %s $%d::bigint`,
				entitiy,
				operand,
				argIndex,
			)
		case "debit":
			fmt.Fprintf(
				&filterBy,
				`REPLACE(COALESCE(%s.debit,'0'),',','')::bigint %s $%d::bigint`,
				entitiy,
				operand,
				argIndex,
			)
		case "credit":
			fmt.Fprintf(
				&filterBy,
				`REPLACE(COALESCE(%s.credit,'0'),',','')::bigint %s $%d::bigint`,
				entitiy,
				operand,
				argIndex,
			)
		case "buy_price":
			fmt.Fprintf(&filterBy, `REPLACE(%s.buy_price,',','')::bigint %s $%d::bigint`, entitiy,
				operand,
				argIndex,
			)
		case "sell_price":
			fmt.Fprintf(
				&filterBy,
				`REPLACE(%s.sell_price,',','')::bigint %s $%d::bigint`,
				entitiy,
				operand,
				argIndex,
			)
		case "count":
			fmt.Fprintf(&filterBy, `%s.count::bigint %s $%d::bigint`, entitiy, operand, argIndex)
		case "current_count":
			fmt.Fprintf(&filterBy,
				`%s.count::bigint + COALESCE(stock.stock_change, 0) %s $%d::bigint`,
				entitiy,
				operand,
				argIndex,
			)
		case "created_at":
			fmt.Fprintf(&filterBy, `%s.created_at::date %s $%d::date`, entitiy, operand, argIndex)
		default:
			fmt.Fprintf(&filterBy, `%s.%s %s $%d`, entitiy, filter, operand, argIndex)
		}

		if i < len(filters)-1 {
			filterBy.WriteString(" AND ")
		}

		argIndex++
	}

	return filterBy.String(), args
}
