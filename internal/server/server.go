package server
import ("encoding/json";"log";"net/http";"github.com/stockyard-dev/stockyard-cartograph/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("GET /api/sites",s.listSites);s.mux.HandleFunc("POST /api/sites",s.createSite);s.mux.HandleFunc("GET /api/sites/{id}",s.getSite);s.mux.HandleFunc("DELETE /api/sites/{id}",s.deleteSite)
s.mux.HandleFunc("GET /api/sites/{id}/urls",s.listURLs);s.mux.HandleFunc("POST /api/sites/{id}/urls",s.addURL);s.mux.HandleFunc("DELETE /api/urls/{id}",s.deleteURL)
s.mux.HandleFunc("GET /api/sites/{id}/sitemap.xml",s.generateXML)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listSites(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"sites":oe(s.db.ListSites())})}
func(s *Server)createSite(w http.ResponseWriter,r *http.Request){var site store.Site;json.NewDecoder(r.Body).Decode(&site);if site.Name==""||site.BaseURL==""{we(w,400,"name and base_url required");return};s.db.CreateSite(&site);wj(w,201,s.db.GetSite(site.ID))}
func(s *Server)getSite(w http.ResponseWriter,r *http.Request){site:=s.db.GetSite(r.PathValue("id"));if site==nil{we(w,404,"not found");return};wj(w,200,site)}
func(s *Server)deleteSite(w http.ResponseWriter,r *http.Request){s.db.DeleteSite(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)listURLs(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"urls":oe(s.db.ListURLs(r.PathValue("id")))})}
func(s *Server)addURL(w http.ResponseWriter,r *http.Request){var u store.URL;json.NewDecoder(r.Body).Decode(&u);if u.Loc==""{we(w,400,"loc required");return};u.SiteID=r.PathValue("id");s.db.AddURL(&u);wj(w,201,u)}
func(s *Server)deleteURL(w http.ResponseWriter,r *http.Request){s.db.DeleteURL(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)generateXML(w http.ResponseWriter,r *http.Request){xml:=s.db.GenerateXML(r.PathValue("id"));if xml==""{we(w,404,"not found");return};w.Header().Set("Content-Type","application/xml");w.Write([]byte(xml))}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"cartograph","sites":st.Sites,"urls":st.URLs})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
