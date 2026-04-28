package plugin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type testPlugin struct {
	*BasePlugin
	deps     []string
	critical bool
	initErr  error
	startErr error
	routes   []Route
}

func newTestPlugin(name string) *testPlugin {
	return &testPlugin{
		BasePlugin: NewBasePlugin(name, "1.0.0", "test", "test"),
	}
}

func (p *testPlugin) Dependencies() []string { return p.deps }
func (p *testPlugin) Critical() bool         { return p.critical }
func (p *testPlugin) Init(ctx context.Context, config map[string]interface{}) error {
	return p.initErr
}
func (p *testPlugin) Start(ctx context.Context) error {
	return p.startErr
}
func (p *testPlugin) Routes() []Route { return p.routes }

func TestManagerLoadConfigDependencyValidation(t *testing.T) {
	mgr := NewManager()
	cache := newTestPlugin("cache")
	cache.deps = []string{"db"}
	db := newTestPlugin("db")
	if err := mgr.Register(cache); err != nil {
		t.Fatalf("register cache: %v", err)
	}
	if err := mgr.Register(db); err != nil {
		t.Fatalf("register db: %v", err)
	}

	err := mgr.LoadConfig(map[string]PluginConfig{
		"cache": {Enabled: true},
		"db":    {Enabled: false},
	})
	if err == nil {
		t.Fatalf("expect dependency validation error")
	}
}

func TestManagerStartCriticalPluginFailure(t *testing.T) {
	mgr := NewManager()
	core := newTestPlugin("core")
	core.critical = true
	core.startErr = errors.New("boom")
	if err := mgr.Register(core); err != nil {
		t.Fatalf("register core: %v", err)
	}
	if err := mgr.LoadConfig(map[string]PluginConfig{
		"core": {Enabled: true, Critical: true},
	}); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := mgr.Initialize(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := mgr.Start(context.Background()); err == nil {
		t.Fatalf("expect critical plugin start error")
	}
}

func TestManagerApplyRoutesAndMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mgr := NewManager()
	web := newTestPlugin("web")
	web.routes = []Route{
		{
			Method: "GET",
			Path:   "/plugin/ping",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
		},
	}
	if err := mgr.Register(web); err != nil {
		t.Fatalf("register web: %v", err)
	}
	if err := mgr.LoadConfig(map[string]PluginConfig{
		"web": {Enabled: true},
	}); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := mgr.Initialize(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	engine := gin.New()
	mgr.ApplyRoutes(engine)
	req := httptest.NewRequest(http.MethodGet, "/plugin/ping", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	metrics := mgr.GetMetrics()
	if metrics["web"].StartDurationMs < 0 {
		t.Fatalf("unexpected start duration")
	}
}
