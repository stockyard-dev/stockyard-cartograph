package server

import (
	"encoding/json"
	"github.com/stockyard-dev/stockyard-cartograph/internal/store"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.mux.HandleFunc("GET /api/sites", s.listSites)
	s.mux.HandleFunc("POST /api/sites", s.createSite)
	s.mux.HandleFunc("GET /api/sites/{id}", s.getSite)
	s.mux.HandleFunc("DELETE /api/sites/{id}", s.deleteSite)
	s.mux.HandleFunc("GET /api/sites/{id}/urls", s.listURLs)
	s.mux.HandleFunc("POST /api/sites/{id}/urls", s.addURL)
	s.mux.HandleFunc("DELETE /api/urls/{id}", s.deleteURL)
	s.mux.HandleFunc("GET /api/sites/{id}/sitemap.xml", s.generateXML)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/cartograph/"})
	})
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(v)
}
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}
func (s *Server) listSites(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"sites": oe(s.db.ListSites())})
}
func (s *Server) createSite(w http.ResponseWriter, r *http.Request) {
	var site store.Site
	json.NewDecoder(r.Body).Decode(&site)
	if site.Name == "" || site.BaseURL == "" {
		we(w, 400, "name and base_url required")
		return
	}
	s.db.CreateSite(&site)
	wj(w, 201, s.db.GetSite(site.ID))
}
func (s *Server) getSite(w http.ResponseWriter, r *http.Request) {
	site := s.db.GetSite(r.PathValue("id"))
	if site == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, site)
}
func (s *Server) deleteSite(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteSite(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}
func (s *Server) listURLs(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"urls": oe(s.db.ListURLs(r.PathValue("id")))})
}
func (s *Server) addURL(w http.ResponseWriter, r *http.Request) {
	var u store.URL
	json.NewDecoder(r.Body).Decode(&u)
	if u.Loc == "" {
		we(w, 400, "loc required")
		return
	}
	u.SiteID = r.PathValue("id")
	s.db.AddURL(&u)
	wj(w, 201, u)
}
func (s *Server) deleteURL(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteURL(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}
func (s *Server) generateXML(w http.ResponseWriter, r *http.Request) {
	xml := s.db.GenerateXML(r.PathValue("id"))
	if xml == "" {
		we(w, 404, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml))
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) { wj(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	wj(w, 200, map[string]any{"status": "ok", "service": "cartograph", "sites": st.Sites, "urls": st.URLs})
}
func oe[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

// ─── personalization (auto-added) ──────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("%s: warning: could not parse config.json: %v", "cartograph", err)
		return
	}
	s.pCfg = cfg
	log.Printf("%s: loaded personalization from %s", "cartograph", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read body"}`, 400)
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		http.Error(w, `{"error":"invalid json"}`, 400)
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		http.Error(w, `{"error":"save failed"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":"saved"}`))
}
