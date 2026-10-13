package licensing

import (
	"net/http"

	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler                       { return &Handler{service: service} }
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) { httputil.OK(w, h.service.Status()) }
