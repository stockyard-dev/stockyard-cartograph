package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Cartograph</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{padding:1.5rem;max-width:960px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(2,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center}
.search{flex:1;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.site{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.6rem}
.site-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.site-name{font-size:.85rem;font-weight:700}
.site-url{font-size:.6rem;color:var(--blue);margin-top:.1rem}
.site-meta{font-size:.55rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.6rem;align-items:center}
.site-actions{display:flex;gap:.3rem;flex-shrink:0}
.urls{margin-top:.4rem;padding-left:.5rem;border-left:2px solid var(--bg3)}
.url-item{font-size:.6rem;color:var(--cd);padding:.15rem 0;display:flex;justify-content:space-between;align-items:center}
.url-item:hover{color:var(--cream)}
.url-pri{font-size:.5rem;color:var(--cm)}
.btn{font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}
.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:460px;max-width:92vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1><span>&#9670;</span> CARTOGRAPH</h1><div style="display:flex;gap:.4rem"><button class="btn btn-p" onclick="openSiteForm()">+ Site</button></div></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar"><input class="search" id="search" placeholder="Search sites and URLs..." oninput="render()"></div>
<div id="list"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()"><div class="modal" id="mdl"></div></div>
<script>
var A='/api',sites=[],urls={},editId=null,editType='';
async function load(){var r=await fetch(A+'/sites').then(function(r){return r.json()});sites=r.sites||[];
for(var i=0;i<sites.length;i++){var u=await fetch(A+'/sites/'+sites[i].id+'/urls').then(function(r){return r.json()}).catch(function(){return{urls:[]}});urls[sites[i].id]=u.urls||[];}
renderStats();render();}
function renderStats(){var totalUrls=0;Object.values(urls).forEach(function(u){totalUrls+=u.length});
document.getElementById('stats').innerHTML='<div class="st"><div class="st-v">'+sites.length+'</div><div class="st-l">Sites</div></div><div class="st"><div class="st-v">'+totalUrls+'</div><div class="st-l">URLs</div></div>';}
function render(){var q=(document.getElementById('search').value||'').toLowerCase();var f=sites;
if(q)f=f.filter(function(s){return(s.name||'').toLowerCase().includes(q)||(s.base_url||'').toLowerCase().includes(q)});
if(!f.length){document.getElementById('list').innerHTML='<div class="empty">No sites. Add one to generate sitemaps.</div>';return;}
var h='';f.forEach(function(s){
h+='<div class="site"><div class="site-top"><div style="flex:1">';
h+='<div class="site-name">'+esc(s.name)+'</div>';
h+='<div class="site-url">'+esc(s.base_url)+'</div>';
h+='</div><div class="site-actions">';
h+='<button class="btn btn-sm" onclick="openUrlForm(''+s.id+'')">+ URL</button>';
h+='<button class="btn btn-sm" onclick="delSite(''+s.id+'')" style="color:var(--red)">&#10005;</button>';
h+='</div></div>';
h+='<div class="site-meta"><span>'+s.url_count+' URLs</span>';
if(s.last_generated)h+='<span>Generated: '+ft(s.last_generated)+'</span>';
h+='</div>';
var su=urls[s.id]||[];
if(su.length){h+='<div class="urls">';su.forEach(function(u){
h+='<div class="url-item"><span>'+esc(u.loc)+'</span><span class="url-pri">'+(u.priority||'0.5')+' | '+(u.changefreq||'weekly')+'</span></div>';
});h+='</div>';}
h+='</div>';});
document.getElementById('list').innerHTML=h;}
async function delSite(id){if(!confirm('Delete site and all URLs?'))return;await fetch(A+'/sites/'+id,{method:'DELETE'});load();}
function openSiteForm(){editType='site';editId=null;
document.getElementById('mdl').innerHTML='<h2>NEW SITE</h2><div class="fr"><label>Name *</label><input id="f-name" placeholder="My Website"></div><div class="fr"><label>Base URL *</label><input id="f-url" placeholder="https://example.com"></div><div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submitSite()">Create</button></div>';
document.getElementById('mbg').classList.add('open');}
function openUrlForm(siteId){editType='url';editId=siteId;
document.getElementById('mdl').innerHTML='<h2>ADD URL</h2><div class="fr"><label>Location *</label><input id="f-loc" placeholder="/about"></div><div class="row2"><div class="fr"><label>Priority</label><input id="f-pri" placeholder="0.5"></div><div class="fr"><label>Change Freq</label><input id="f-freq" placeholder="weekly"></div></div><div class="fr"><label>Last Modified</label><input id="f-mod" type="date"></div><div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submitUrl()">Add</button></div>';
document.getElementById('mbg').classList.add('open');}
function closeModal(){document.getElementById('mbg').classList.remove('open');}
async function submitSite(){var name=document.getElementById('f-name').value.trim();var url=document.getElementById('f-url').value.trim();
if(!name||!url){alert('Name and URL required');return;}
await fetch(A+'/sites',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:name,base_url:url})});closeModal();load();}
async function submitUrl(){var loc=document.getElementById('f-loc').value.trim();if(!loc){alert('Location required');return;}
await fetch(A+'/sites/'+editId+'/urls',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({loc:loc,priority:document.getElementById('f-pri').value.trim(),changefreq:document.getElementById('f-freq').value.trim(),lastmod:document.getElementById('f-mod').value})});closeModal();load();}
function ft(t){if(!t)return'';try{return new Date(t).toLocaleDateString('en-US',{month:'short',day:'numeric'})}catch(e){return t;}}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML;}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal();});
load();
</script></body></html>`
