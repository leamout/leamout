package worker

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestWorkerMetricsHandler(t *testing.T) {
 w:=httptest.NewRecorder()
 workerMetricsHandler().ServeHTTP(w,httptest.NewRequestWithContext(t.Context(),http.MethodGet,"/metrics",nil))
 if w.Code!=http.StatusOK || !strings.Contains(w.Body.String(),"leamout_worker_running 1") {t.Fatalf("metrics response: %d %s",w.Code,w.Body.String())}
}
