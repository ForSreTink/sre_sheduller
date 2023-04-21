// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.2 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

// Defines values for PostWorkPriority.
const (
	PostWorkPriorityCritical PostWorkPriority = "critical"
	PostWorkPriorityRegular  PostWorkPriority = "regular"
)

// Defines values for PostWorkWorkType.
const (
	PostWorkWorkTypeAutomatic PostWorkWorkType = "automatic"
	PostWorkWorkTypeManual    PostWorkWorkType = "manual"
)

// Defines values for WorkPriority.
const (
	WorkPriorityCritical WorkPriority = "critical"
	WorkPriorityRegular  WorkPriority = "regular"
)

// Defines values for WorkStatus.
const (
	WorkStatusCancelled  WorkStatus = "cancelled"
	WorkStatusInProgress WorkStatus = "in_progress"
	WorkStatusPlanned    WorkStatus = "planned"
)

// Defines values for WorkWorkType.
const (
	WorkWorkTypeAutomatic WorkWorkType = "automatic"
	WorkWorkTypeManual    WorkWorkType = "manual"
)

// Defines values for WorksPriority.
const (
	Critical WorksPriority = "critical"
	Regular  WorksPriority = "regular"
)

// Defines values for WorksStatus.
const (
	WorksStatusCancelled  WorksStatus = "cancelled"
	WorksStatusInProgress WorksStatus = "in_progress"
	WorksStatusPlanned    WorksStatus = "planned"
)

// Defines values for WorksWorkType.
const (
	Automatic WorksWorkType = "automatic"
	Manual    WorksWorkType = "manual"
)

// Defines values for GetscheduleParamsStatuses.
const (
	Cancelled  GetscheduleParamsStatuses = "cancelled"
	InProgress GetscheduleParamsStatuses = "in_progress"
	Planned    GetscheduleParamsStatuses = "planned"
)

// Error defines model for error.
type Error struct {
	Alternative *[]struct {
		Deadline        *time.Time `json:"deadline,omitempty"`
		DurationMinutes *int32     `json:"durationMinutes,omitempty"`
		StartDate       *time.Time `json:"startDate,omitempty"`
		Zones           *[]string  `json:"zones,omitempty"`
	} `json:"alternative,omitempty"`
	ErrorCode *string `json:"errorCode,omitempty"`
	Message   *string `json:"message,omitempty"`
}

// MoveWork defines model for moveWork.
type MoveWork struct {
	DurationMinutes *int32    `json:"durationMinutes,omitempty"`
	StartDate       time.Time `json:"startDate"`
}

// PostWork defines model for postWork.
type PostWork struct {
	Deadline        *time.Time        `json:"deadline,omitempty"`
	DurationMinutes *int32            `json:"durationMinutes,omitempty"`
	Priority        *PostWorkPriority `json:"priority,omitempty"`
	StartDate       *time.Time        `json:"startDate,omitempty"`
	WorkType        *PostWorkWorkType `json:"workType,omitempty"`
	Zones           *interface{}      `json:"zones,omitempty"`
}

// PostWorkPriority defines model for PostWork.Priority.
type PostWorkPriority string

// PostWorkWorkType defines model for PostWork.WorkType.
type PostWorkWorkType string

// ProlongateWork defines model for prolongateWork.
type ProlongateWork struct {
	DurationMinutes *int32 `json:"durationMinutes,omitempty"`
}

// Work defines model for work.
type Work struct {
	Deadline        *time.Time    `json:"deadline,omitempty"`
	DurationMinutes *int32        `json:"durationMinutes,omitempty"`
	Id              *string       `json:"id,omitempty"`
	Priority        *WorkPriority `json:"priority,omitempty"`
	StartDate       *time.Time    `json:"startDate,omitempty"`
	Status          *WorkStatus   `json:"status,omitempty"`
	WorkId          *string       `json:"workId,omitempty"`
	WorkType        *WorkWorkType `json:"workType,omitempty"`
	Zones           *[]string     `json:"zones,omitempty"`
}

// WorkPriority defines model for Work.Priority.
type WorkPriority string

// WorkStatus defines model for Work.Status.
type WorkStatus string

// WorkWorkType defines model for Work.WorkType.
type WorkWorkType string

// Works defines model for works.
type Works = []struct {
	Deadline        *time.Time     `json:"deadline,omitempty"`
	DurationMinutes *int32         `json:"durationMinutes,omitempty"`
	Id              *string        `json:"id,omitempty"`
	Priority        *WorksPriority `json:"priority,omitempty"`
	StartDate       *time.Time     `json:"startDate,omitempty"`
	Status          *WorksStatus   `json:"status,omitempty"`
	WorkId          *string        `json:"workId,omitempty"`
	WorkType        *WorksWorkType `json:"workType,omitempty"`
	Zones           *[]string      `json:"zones,omitempty"`
}

// WorksPriority defines model for Works.Priority.
type WorksPriority string

// WorksStatus defines model for Works.Status.
type WorksStatus string

// WorksWorkType defines model for Works.WorkType.
type WorksWorkType string

// GetscheduleParams defines parameters for Getschedule.
type GetscheduleParams struct {
	// FromDate Starts from
	FromDate *time.Time `form:"fromDate,omitempty" json:"fromDate,omitempty"`

	// ToDate Starts to
	ToDate *time.Time `form:"toDate,omitempty" json:"toDate,omitempty"`

	// Zones List of zones
	Zones *[]string `form:"zones,omitempty" json:"zones,omitempty"`

	// Statuses Statuses of work to get
	Statuses *[]GetscheduleParamsStatuses `form:"statuses,omitempty" json:"statuses,omitempty"`
}

// GetscheduleParamsStatuses defines parameters for Getschedule.
type GetscheduleParamsStatuses string

// AddWorkJSONRequestBody defines body for AddWork for application/json ContentType.
type AddWorkJSONRequestBody = PostWork

// MoveWorkByIdJSONRequestBody defines body for MoveWorkById for application/json ContentType.
type MoveWorkByIdJSONRequestBody = MoveWork

// ProlongateWorkByIdJSONRequestBody defines body for ProlongateWorkById for application/json ContentType.
type ProlongateWorkByIdJSONRequestBody = ProlongateWork

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get schedule with works
	// (GET /schedule)
	Getschedule(w http.ResponseWriter, r *http.Request, params GetscheduleParams)
	// Create new planned work in avialable zone
	// (POST /work)
	AddWork(w http.ResponseWriter, r *http.Request)
	// Get planned work by id
	// (GET /work/{workId})
	GetWorkById(w http.ResponseWriter, r *http.Request, workId string)
	// Cancel planned work by id
	// (PUT /work/{workId}/cancel)
	CancelWorkById(w http.ResponseWriter, r *http.Request, workId string)
	// Move start time and duration for planned work
	// (PUT /work/{workId}/move)
	MoveWorkById(w http.ResponseWriter, r *http.Request, workId string)
	// Prolongate work duration started work
	// (PUT /work/{workId}/prolongate)
	ProlongateWorkById(w http.ResponseWriter, r *http.Request, workId string)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// Getschedule operation middleware
func (siw *ServerInterfaceWrapper) Getschedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetscheduleParams

	// ------------- Optional query parameter "fromDate" -------------

	err = runtime.BindQueryParameter("form", true, false, "fromDate", r.URL.Query(), &params.FromDate)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "fromDate", Err: err})
		return
	}

	// ------------- Optional query parameter "toDate" -------------

	err = runtime.BindQueryParameter("form", true, false, "toDate", r.URL.Query(), &params.ToDate)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "toDate", Err: err})
		return
	}

	// ------------- Optional query parameter "zones" -------------

	err = runtime.BindQueryParameter("form", true, false, "zones", r.URL.Query(), &params.Zones)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "zones", Err: err})
		return
	}

	// ------------- Optional query parameter "statuses" -------------

	err = runtime.BindQueryParameter("form", true, false, "statuses", r.URL.Query(), &params.Statuses)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "statuses", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Getschedule(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// AddWork operation middleware
func (siw *ServerInterfaceWrapper) AddWork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.AddWork(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetWorkById operation middleware
func (siw *ServerInterfaceWrapper) GetWorkById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "workId" -------------
	var workId string

	err = runtime.BindStyledParameter("simple", false, "workId", mux.Vars(r)["workId"], &workId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "workId", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetWorkById(w, r, workId)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// CancelWorkById operation middleware
func (siw *ServerInterfaceWrapper) CancelWorkById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "workId" -------------
	var workId string

	err = runtime.BindStyledParameter("simple", false, "workId", mux.Vars(r)["workId"], &workId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "workId", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CancelWorkById(w, r, workId)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// MoveWorkById operation middleware
func (siw *ServerInterfaceWrapper) MoveWorkById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "workId" -------------
	var workId string

	err = runtime.BindStyledParameter("simple", false, "workId", mux.Vars(r)["workId"], &workId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "workId", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.MoveWorkById(w, r, workId)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// ProlongateWorkById operation middleware
func (siw *ServerInterfaceWrapper) ProlongateWorkById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "workId" -------------
	var workId string

	err = runtime.BindStyledParameter("simple", false, "workId", mux.Vars(r)["workId"], &workId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "workId", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.ProlongateWorkById(w, r, workId)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshallingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshallingParamError) Error() string {
	return fmt.Sprintf("Error unmarshalling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshallingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{})
}

type GorillaServerOptions struct {
	BaseURL          string
	BaseRouter       *mux.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *mux.Router) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r *mux.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options GorillaServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = mux.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.HandleFunc(options.BaseURL+"/schedule", wrapper.Getschedule).Methods("GET")

	r.HandleFunc(options.BaseURL+"/work", wrapper.AddWork).Methods("POST")

	r.HandleFunc(options.BaseURL+"/work/{workId}", wrapper.GetWorkById).Methods("GET")

	r.HandleFunc(options.BaseURL+"/work/{workId}/cancel", wrapper.CancelWorkById).Methods("PUT")

	r.HandleFunc(options.BaseURL+"/work/{workId}/move", wrapper.MoveWorkById).Methods("PUT")

	r.HandleFunc(options.BaseURL+"/work/{workId}/prolongate", wrapper.ProlongateWorkById).Methods("PUT")

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xYX2/jNgz/KgI3YC8+O01vL367dsUQ4Doc1gJ7KIpBtZhEV1nSSXRyWZHvPkjOv4vV",
	"NrmmXbf1KY5NihT5+5GU7qAytTUaNXko78BXY6x5fETnjAsP1hmLjiTG11wROs1JTjD8lYS170oJ5EJJ",
	"HUWGxtWcoATBCd+RrBEyoJlFKMGTk3oE8wxE4zhJo8+lbqhdZKUoNR3310pSE47QBS1P3NEvnPYw9JfR",
	"7fIr1zsiixfcOT6D+fqFufmMFXUlsjZYp0Zgcr0aveej1LfU2rWZ4B/G3Sai2o0RfuW1VQhlP3uWeM0z",
	"cPilkQ4FlFcb+tcJz63xdI/nh8DDnnu1ThonaRZVdVMH/x2OGsUdZFA5SbLiamMja+PfAaupcbeX8eXa",
	"Ws11wxVkwBsyNSdZJa2tELna4BWI6ggyEFU/aLSs3AuzKWBZZ5TRI07PBK+UzemrgYMUybi9LEo8cWr8",
	"pi2ruNYogi2uK1QqPkv9p3Vm5ND7pOkQ1kF6Q4cA4lNKY7Dvn7s1vOX+Vea+W4SkHpo2675y0oY0QgmX",
	"Y+mZRzdBxwROUBmLgg2NYxe/nzFyvLplUrPB5U+eXUp9a4ZDdmFUE9TZaWNzyEDJCrWPW9W8DmY/WF6N",
	"kfXzHmTQOAUljIlsWRTT6TTn8Wtu3KhYqPri4+D07LeLs3f9vJePqVZxD5ICsCDUSBYqr2gUBjxM0PnW",
	"/aO8lx8dBWFjUXMroYTjvJcfQwaW0ziGr1jqhj8jpG4UfkVaGWBTSWPWcicu20I+5DnIrdYKBhyvkdB5",
	"KK+2l7wIsPRs6EwNGeBXq+JMQq7BkAso4UuDbgbZMmZBMsK47fM89vlWfN11dhsU7vGFzG6ekHkmPz5K",
	"T8wMWYvwnXxZin5H402FgRqPPrgQ8svIsACHnRzxC+W0L0+rI1uOX4fIe2sCMcLi/V4v/FRGE+qIXm6t",
	"klXEZfHZh83dbXj1o8MhlPBDsT5RFIvjRBGngFgPtmLTVBX6YaMi89733ndJEmkYySEF04YNTaNFkP75",
	"gP61Z52EgwMdTztqWayWghn4pq65mz1IY+IjH6fmJXuvg2axmomMT1SFU4eckGmcskVeW9hIzfhEcsVv",
	"FEYsd+rEByHiYNdSCD2dGDE7WIxWw30iTPv4vD5OBNjPO6g76oZkDRNWRUuihcsLAOCEC7aI5qsC3T4R",
	"X8Iw4m4NweKunSTmDzaob1a/mTEpUv0pAONkNhCP9aeBWJZBWBS80C/X9W4x22yjZLP6bdf7t7r1hLqV",
	"zO5jeCnaLhNLWJOqYPHzLshpJf/n4Hn5KvYvBexDuHoUs7VpLyyTiD03E2TxSMfCVMm4Fmx56ozHkk2b",
	"HRSfLy7s/kkMH77fr64hE9naK1452zrAM+mZiUtxle8wEbxR8tVScl/iPErT9R3lvWT9tBJpS8DKXvTj",
	"PpJ++uby879F1a2L3UQud43Zy1LRv3HxgFzcNcfbHJzP/w4AAP//LzXAxgYcAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
