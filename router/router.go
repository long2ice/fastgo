package router

import (
	"container/list"
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	"github.com/long2ice/swagin/security"
	"github.com/mcuadros/go-defaults"
)

type Model any
type ErrorHandlerFunc func(ctx *gin.Context, err error, status int)

type Router struct {
	Handlers            *list.List
	Path                string
	Method              string
	Summary             string
	Description         string
	Deprecated          bool
	RequestContentType  string
	ResponseContentType string
	Tags                []string
	API                 gin.HandlerFunc
	Model               Model
	OperationID         string
	Exclude             bool
	Securities          []security.ISecurity
	Response            Response
}

const requestKey = "request"

var validate = validator.New()

func BindModel(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := reflect.New(reflect.TypeOf(model).Elem()).Interface()
		if err := c.ShouldBindHeader(req); err != nil {
			log.Panic(err)
		}
		if err := CookiesParser(c, req); err != nil {
			log.Panic(err)
		}
		if err := c.ShouldBindWith(req, Query); err != nil {
			log.Panic(err)
		}
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			switch c.Request.Header.Get("Content-Type") {
			case binding.MIMEMultipartPOSTForm:
				if err := c.ShouldBindWith(req, binding.FormMultipart); err != nil {
					log.Panic(err)

				}
			case binding.MIMEJSON:
				if err := c.ShouldBindWith(req, binding.JSON); err != nil {
					log.Panic(err)
				}
			case binding.MIMEXML:
				if err := c.ShouldBindWith(req, binding.XML); err != nil {
					log.Panic(err)

				}
			case binding.MIMEPOSTForm:
				if err := c.ShouldBindWith(req, binding.Form); err != nil {
					log.Panic(err)

				}
			case binding.MIMEYAML:
				if err := c.ShouldBindWith(req, binding.YAML); err != nil {
					log.Panic(err)

				}
			case binding.MIMEPROTOBUF:
				if err := c.ShouldBindWith(req, binding.ProtoBuf); err != nil {
					log.Panic(err)

				}
			case binding.MIMEMSGPACK:
				if err := c.ShouldBindWith(req, binding.MsgPack); err != nil {
					log.Panic(err)

				}
			}
		}
		if err := c.ShouldBindUri(req); err != nil {
			log.Panic(err)

		}
		defaults.SetDefaults(req)
		if err := validate.Struct(req); err != nil {
			log.Panic(err)

		}
		c.Set(requestKey, req)
		c.Next()
	}
}

func (router *Router) GetHandlers() []gin.HandlerFunc {
	var handlers []gin.HandlerFunc
	for _, s := range router.Securities {
		handlers = append(handlers, s.Authorize)
	}
	for h := router.Handlers.Front(); h != nil; h = h.Next() {
		if f, ok := h.Value.(gin.HandlerFunc); ok {
			handlers = append(handlers, f)
		}
	}
	handlers = append(handlers, router.API)
	return handlers
}

func NewX(f gin.HandlerFunc, options ...Option) *Router {
	r := &Router{
		Handlers: list.New(),
		Response: make(Response),
		API: func(ctx *gin.Context) {
			f(ctx)
		},
	}
	for _, option := range options {
		option(r)
	}
	return r
}
func New[T Model, F func(c *gin.Context, req T)](f F, options ...Option) *Router {
	var model T
	h := BindModel(&model)
	r := &Router{
		Handlers: list.New(),
		Response: make(Response),
		API: func(ctx *gin.Context) {
			req, ok := ctx.Get(requestKey)
			if !ok {
				log.Panic("Could not get request from context")
			}
			f(ctx, *(req.(*T)))
		},
		Model: model,
	}
	for _, option := range options {
		option(r)
	}

	r.Handlers.PushBack(h)
	return r
}
func (router *Router) WithSecurity(securities ...security.ISecurity) *Router {
	Security(securities...)(router)
	return router
}
func (router *Router) WithResponses(response Response) *Router {
	Responses(response)(router)
	return router
}
func (router *Router) WithHandlers(handlers ...gin.HandlerFunc) *Router {
	Handlers(handlers...)(router)
	return router
}
func (router *Router) WithTags(tags ...string) *Router {
	Tags(tags...)(router)
	return router
}
func (router *Router) WithSummary(summary string) *Router {
	Summary(summary)(router)
	return router
}
func (router *Router) WithDescription(description string) *Router {
	Description(description)(router)
	return router
}
func (router *Router) WithDeprecated() *Router {
	Deprecated()(router)
	return router
}
func (router *Router) WithOperationID(ID string) *Router {
	OperationID(ID)(router)
	return router
}
func (router *Router) WithExclude() *Router {
	Exclude()(router)
	return router
}
func (router *Router) WithContentType(contentType string, contentTypeType ContentTypeType) *Router {
	ContentType(contentType, contentTypeType)(router)
	return router
}
