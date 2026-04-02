package store
import ("database/sql";"fmt";"os";"path/filepath";"strings";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Site struct{ID string `json:"id"`;Name string `json:"name"`;BaseURL string `json:"base_url"`;CreatedAt string `json:"created_at"`;URLCount int `json:"url_count"`;LastGenerated string `json:"last_generated,omitempty"`}
type URL struct{ID string `json:"id"`;SiteID string `json:"site_id"`;Loc string `json:"loc"`;Lastmod string `json:"lastmod,omitempty"`;Changefreq string `json:"changefreq,omitempty"`;Priority string `json:"priority,omitempty"`;CreatedAt string `json:"created_at"`}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"cartograph.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
for _,q:=range[]string{
`CREATE TABLE IF NOT EXISTS sites(id TEXT PRIMARY KEY,name TEXT NOT NULL,base_url TEXT NOT NULL,created_at TEXT DEFAULT(datetime('now')),last_generated TEXT DEFAULT '')`,
`CREATE TABLE IF NOT EXISTS urls(id TEXT PRIMARY KEY,site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,loc TEXT NOT NULL,lastmod TEXT DEFAULT '',changefreq TEXT DEFAULT 'weekly',priority TEXT DEFAULT '0.5',created_at TEXT DEFAULT(datetime('now')))`,
`CREATE INDEX IF NOT EXISTS idx_urls_site ON urls(site_id)`,
}{if _,err:=db.Exec(q);err!=nil{return nil,fmt.Errorf("migrate: %w",err)}};return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)CreateSite(s *Site)error{s.ID=genID();s.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO sites(id,name,base_url,created_at)VALUES(?,?,?,?)`,s.ID,s.Name,s.BaseURL,s.CreatedAt);return err}
func(d *DB)GetSite(id string)*Site{var s Site;if d.db.QueryRow(`SELECT id,name,base_url,created_at,last_generated FROM sites WHERE id=?`,id).Scan(&s.ID,&s.Name,&s.BaseURL,&s.CreatedAt,&s.LastGenerated)!=nil{return nil};d.db.QueryRow(`SELECT COUNT(*) FROM urls WHERE site_id=?`,id).Scan(&s.URLCount);return &s}
func(d *DB)ListSites()[]Site{rows,_:=d.db.Query(`SELECT id,name,base_url,created_at,last_generated FROM sites ORDER BY name`);if rows==nil{return nil};defer rows.Close();var o []Site;for rows.Next(){var s Site;rows.Scan(&s.ID,&s.Name,&s.BaseURL,&s.CreatedAt,&s.LastGenerated);d.db.QueryRow(`SELECT COUNT(*) FROM urls WHERE site_id=?`,s.ID).Scan(&s.URLCount);o=append(o,s)};return o}
func(d *DB)DeleteSite(id string)error{d.db.Exec(`DELETE FROM urls WHERE site_id=?`,id);_,err:=d.db.Exec(`DELETE FROM sites WHERE id=?`,id);return err}
func(d *DB)AddURL(u *URL)error{u.ID=genID();u.CreatedAt=now();if u.Changefreq==""{u.Changefreq="weekly"};if u.Priority==""{u.Priority="0.5"};_,err:=d.db.Exec(`INSERT INTO urls(id,site_id,loc,lastmod,changefreq,priority,created_at)VALUES(?,?,?,?,?,?,?)`,u.ID,u.SiteID,u.Loc,u.Lastmod,u.Changefreq,u.Priority,u.CreatedAt);return err}
func(d *DB)ListURLs(siteID string)[]URL{rows,_:=d.db.Query(`SELECT id,site_id,loc,lastmod,changefreq,priority,created_at FROM urls WHERE site_id=? ORDER BY loc`,siteID);if rows==nil{return nil};defer rows.Close();var o []URL;for rows.Next(){var u URL;rows.Scan(&u.ID,&u.SiteID,&u.Loc,&u.Lastmod,&u.Changefreq,&u.Priority,&u.CreatedAt);o=append(o,u)};return o}
func(d *DB)DeleteURL(id string)error{_,err:=d.db.Exec(`DELETE FROM urls WHERE id=?`,id);return err}
func(d *DB)GenerateXML(siteID string)string{s:=d.GetSite(siteID);if s==nil{return""};urls:=d.ListURLs(siteID)
var b strings.Builder;b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
for _,u:=range urls{b.WriteString("  <url>\n    <loc>"+s.BaseURL+u.Loc+"</loc>\n");if u.Lastmod!=""{b.WriteString("    <lastmod>"+u.Lastmod+"</lastmod>\n")};b.WriteString("    <changefreq>"+u.Changefreq+"</changefreq>\n    <priority>"+u.Priority+"</priority>\n  </url>\n")}
b.WriteString("</urlset>");d.db.Exec(`UPDATE sites SET last_generated=? WHERE id=?`,now(),siteID);return b.String()}
type Stats struct{Sites int `json:"sites"`;URLs int `json:"urls"`}
func(d *DB)Stats()Stats{var s Stats;d.db.QueryRow(`SELECT COUNT(*) FROM sites`).Scan(&s.Sites);d.db.QueryRow(`SELECT COUNT(*) FROM urls`).Scan(&s.URLs);return s}
