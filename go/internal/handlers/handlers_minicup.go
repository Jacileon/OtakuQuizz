package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ============================================================================
// PAGE — Mini Cup
// ============================================================================

func (h *Handler) MiniCupPage(c *fiber.Ctx) error {
	content := `
<style>
.minicup-wrap{max-width:700px;margin:0 auto;padding:16px}
.minicup-card{background:rgba(15,23,42,.85);border:2px solid #22c55e;border-radius:16px;overflow:hidden}
.minicup-header{text-align:center;padding:24px 16px 12px}
.minicup-header h1{font-size:2rem;font-weight:900;margin:0;letter-spacing:-1px}
.minicup-header p{color:#94a3b8;margin:6px 0 0;font-size:.9rem}
.minicup-menu{padding:16px 24px 24px;display:flex;flex-direction:column;gap:12px}
.minicup-btn{display:flex;align-items:center;gap:10px;padding:16px 20px;border-radius:12px;font-size:1.05rem;font-weight:700;cursor:pointer;border:2px solid transparent;transition:all .2s}
.minicup-btn:hover{transform:scale(1.02)}
.minicup-btn-solo{background:#22c55e;color:#fff}
.minicup-btn-solo:hover{background:#16a34a}
.minicup-btn-local{background:transparent;color:#e2e8f0;border-color:#334155}
.minicup-btn-local:hover{border-color:#22c55e}
.team-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;padding:16px 24px 24px}
.team-card{display:flex;flex-direction:column;align-items:center;gap:4px;padding:14px 8px;border-radius:12px;cursor:pointer;border:2px solid transparent;background:rgba(30,41,59,.6);transition:all .15s}
.team-card:hover{transform:scale(1.05);border-color:#22c55e}
.team-card.disabled{opacity:.3;cursor:not-allowed}
.team-card .flag{font-size:2rem}
.team-card .name{font-size:.8rem;font-weight:600;color:#e2e8f0}
.scoreboard{display:flex;align-items:center;justify-content:space-between;padding:10px 16px;background:rgba(30,41,59,.8);border-radius:12px;margin-bottom:8px}
.scoreboard .team{display:flex;align-items:center;gap:6px}
.scoreboard .team .flag{font-size:1.4rem}
.scoreboard .team .tname{font-weight:700;font-size:.85rem}
.scoreboard .score{font-size:2rem;font-weight:900}
.scoreboard .sep{color:#64748b;margin:0 6px}
.turn-info{display:flex;justify-content:space-between;align-items:center;font-size:.8rem;color:#94a3b8;margin-bottom:6px}
.turn-dot{width:10px;height:10px;border-radius:50%;display:inline-block;margin-right:6px}
.sudden-death{background:#450a0a;color:#ef4444;padding:2px 8px;border-radius:99px;font-weight:700;font-size:.7rem}
.shot-dots{display:flex;gap:4px;padding:6px 0;flex-wrap:wrap}
.shot-dot{width:24px;height:24px;border-radius:6px;display:flex;align-items:center;justify-content:center;font-size:.7rem;font-weight:700;border:1px solid #334155;background:rgba(30,41,59,.6);color:#64748b}
.shot-dot.goal{background:#166534;border-color:#22c55e;color:#4ade80}
.shot-dot.saved{background:#713f12;border-color:#eab308;color:#facc15}
.shot-dot.miss{background:#7f1d1d;border-color:#ef4444;color:#fca5a5}
.game-canvas-wrap{position:relative;aspect-ratio:4/3;border-radius:12px;overflow:hidden;border:2px solid #166534;margin:8px 0}
.game-canvas-wrap.shake{animation:shake .4s ease-out}
@keyframes shake{0%,100%{transform:translateX(0)}20%{transform:translateX(-6px)}40%{transform:translateX(6px)}60%{transform:translateX(-4px)}80%{transform:translateX(4px)}}
.game-canvas-wrap canvas{width:100%;height:100%;display:block;cursor:crosshair;touch-action:none}
.overlay-msg{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,.5);backdrop-filter:blur(4px)}
.overlay-msg span{font-size:3rem;font-weight:900;text-shadow:0 2px 12px rgba(0,0,0,.6)}
.overlay-msg .goal-txt{color:#4ade80}
.overlay-msg .saved-txt{color:#facc15}
.overlay-msg .miss-txt{color:#fca5a5}
.overlay-transition{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;background:rgba(15,23,42,.92)}
.overlay-transition .t-name{font-size:1.6rem;font-weight:900;margin-top:8px}
.overlay-transition .t-role{color:#94a3b8;font-size:.9rem;margin-top:4px}
.game-over-wrap{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;background:rgba(15,23,42,.95)}
.game-over-wrap h2{font-size:1.4rem;font-weight:900;margin:0 0 6px}
.game-over-wrap .final-score{font-size:3rem;font-weight:900}
.game-over-wrap .xp{color:#facc15;font-weight:700;margin-top:8px}
.game-over-wrap .go-btns{display:flex;gap:10px;margin-top:16px}
.game-over-wrap .go-btns button{padding:10px 20px;border-radius:10px;font-weight:700;font-size:.9rem;cursor:pointer;border:none}
.go-replay{background:#22c55e;color:#fff}
.go-quit{background:#334155;color:#e2e8f0}
.game-controls{display:flex;justify-content:space-between;align-items:center;padding:8px 0}
.game-controls button{padding:6px 14px;border-radius:8px;font-size:.8rem;cursor:pointer;border:1px solid #334155;background:transparent;color:#94a3b8;transition:all .15s}
.game-controls button:hover{border-color:#22c55e;color:#22c55e}
.game-controls .quit-btn{color:#ef4444;border-color:#7f1d1d}
.game-controls .quit-btn:hover{background:#7f1d1d;color:#fff}
.keyboard-panel{display:grid;grid-template-columns:repeat(3,1fr);gap:6px;max-width:300px;margin:8px auto}
.keyboard-panel button{padding:10px;border-radius:8px;font-size:.85rem;font-weight:700;cursor:pointer;border:1px solid #334155;background:rgba(30,41,59,.6);color:#e2e8f0;transition:all .1s}
.keyboard-panel button:active{background:#22c55e;color:#fff;border-color:#22c55e}
</style>
<div class="minicup-wrap" id="minicup-app"></div>
<script>
(function(){
var T=[{code:"fr",name:"France",flag:"\uD83C\uDDEB\uD83C\uDDF7",primary:"#002395",secondary:"#FFFFFF"},{code:"br",name:"Brésil",flag:"\uD83C\uDDE7\uD83C\uDDF7",primary:"#009C3B",secondary:"#FFDF00"},{code:"ar",name:"Argentine",flag:"\uD83C\uDDE6\uD83C\uDDF7",primary:"#75AADB",secondary:"#FFFFFF"},{code:"de",name:"Allemagne",flag:"\uD83C\uDDE9\uD83C\uDDEA",primary:"#000000",secondary:"#DD0000"},{code:"es",name:"Espagne",flag:"\uD83C\uDDEA\uD83C\uDDF8",primary:"#AA151B",secondary:"#F1BF00"},{code:"it",name:"Italie",flag:"\uD83C\uDDEE\uD83C\uDDF9",primary:"#0066CC",secondary:"#FFFFFF"},{code:"pt",name:"Portugal",flag:"\uD83C\uDDF5\uD83C\uDDF9",primary:"#006600",secondary:"#FF0000"},{code:"eng",name:"Angleterre",flag:"\uD83C\uDFF4\uDB40\uDC67\uDB40\uDC62\uDB40\uDC65\uDB40\uDC6E\uDB40\uDC67\uDB40\uDC7F",primary:"#FFFFFF",secondary:"#CF081F"},{code:"nl",name:"Pays-Bas",flag:"\uD83C\uDDF3\uD83C\uDDF1",primary:"#FF4F00",secondary:"#FFFFFF"},{code:"be",name:"Belgique",flag:"\uD83C\uDDE7\uD83C\uDDEA",primary:"#000000",secondary:"#FDDA24"},{code:"jp",name:"Japon",flag:"\uD83C\uDDEF\uD83C\uDDF5",primary:"#FFFFFF",secondary:"#BC002D"},{code:"us",name:"USA",flag:"\uD83C\uDDFA\uD83C\uDDF8",primary:"#3C3B6E",secondary:"#FFFFFF"}];
var GW=220,GH=90,BR=10,KW=44,KH=56,GRAV=0.18,SPD=7;
var app=document.getElementById('minicup-app');
var S={scr:'menu',mode:'',step:'a',selA:'',tA:'',tB:'',ph:'aim',sA:0,sB:0,tm:'a',sh:0,shots:[],msg:'',sid:'',skKb:false,xp:0,shk:false};
var ball={x:0,y:0,vx:0,vy:0,z:0,vz:0};
var kp={x:0,tx:0,dl:0,dv:false};
var traj=[],pDown=false,sx=0,sy=0;
var cvs,ctx,afId,shotN=0;
function ft(c){for(var i=0;i<T.length;i++){if(T[i].code===c)return T[i];}return T[0];}
function render(){
if(S.scr==='menu')rMenu();else if(S.scr==='a')rTeam('a');else if(S.scr==='b')rTeam('b');else if(S.scr==='g')rGame();
}
function rMenu(){
app.innerHTML='<div class="minicup-card"><div class="minicup-header"><h1>\uD83C\uDFC6 Mini Cup</h1><p>Dessine la trajectoire de ton tir pour marquer !</p></div><div class="minicup-menu"><button class="minicup-btn minicup-btn-solo" onclick="MC.go(\'solo\')">\uD83C\uDFC6 Solo vs IA</button><button class="minicup-btn minicup-btn-local" onclick="MC.go(\'local\')">\uD83D\uDC65 2 Joueurs local</button></div></div>';
}
function rTeam(st){
var ti=st==='a'?'Choisis ton équipe':'Choisis l\'adversaire';
var sub=st==='a'?'Joueur 1':(S.mode==='local'?'Joueur 2':'IA');
var h='<div class="minicup-card"><div class="minicup-header"><h1>'+ti+'</h1><p>\uD83D\uDEE1\uFE0F '+sub+'</p></div><div class="team-grid">';
for(var i=0;i<T.length;i++){var t=T[i];var d=st==='b'&&t.code===S.selA;h+='<div class="team-card'+(d?' disabled':'')+'" onclick="MC.pick(\''+t.code+'\')"><span class="flag">'+t.flag+'</span><span class="name">'+t.name+'</span></div>';}
h+='</div><div style="padding:0 24px 16px"><button class="minicup-btn minicup-btn-local" onclick="MC.back()">\u2190 Retour</button></div></div>';
app.innerHTML=h;
}
function rGame(){
var tA=ft(S.tA),tB=ft(S.tB);
var isP=S.mode==='solo'?S.tm==='a':true;
var ci=S.tm==='a'?tA:tB;
var mS=S.mode==='solo'?5:10;
var sd=S.shots.length>=mS;
var h='<div class="scoreboard"><div class="team"><span class="flag">'+tA.flag+'</span><span class="tname">'+tA.name+'</span></div><div><span class="score">'+S.sA+'</span><span class="sep">-</span><span class="score">'+S.sB+'</span></div><div class="team"><span class="tname">'+tB.name+'</span><span class="flag">'+tB.flag+'</span></div></div>';
h+='<div class="turn-info"><div><span class="turn-dot" style="background:'+ci.primary+'"></span>'+(isP?'À toi de tirer':'Tour adversaire');
if(sd)h+=' <span class="sudden-death">MORT SUBITE</span>';
h+='</div><span>Tir '+Math.min(S.sh+1,5)+'/5</span></div>';
h+='<div class="shot-dots">';
for(var i=0;i<S.shots.length;i++){var s=S.shots[i];h+='<div class="shot-dot '+(s.r==='goal'?'goal':(s.r==='saved'?'saved':'miss'))+'">'+(s.r==='goal'?'⚽':(s.r==='saved'?'🧤':'❌'))+'</div>';}
h+='</div>';
h+='<div class="game-canvas-wrap'+(S.shk?' shake':'')+'" id="cw"><canvas id="mc-c">';
if(S.ph==='transition'){var rl=isP?'Tir':'Défense';h+='</canvas><div class="overlay-transition"><span style="font-size:3rem">'+ci.flag+'</span><div class="t-name" style="color:'+ci.primary+'">'+ci.name+'</div><div class="t-role">'+rl+'</div></div></div>';}
else if(S.ph==='result'){var cls=S.msg.indexOf('BUT')>-1?'goal-txt':(S.msg.indexOf('Arrêté')>-1?'saved-txt':'miss-txt');h+='</canvas><div class="overlay-msg"><span class="'+cls+'">'+S.msg+'</span></div></div>';}
else if(S.ph==='finished'){var wT=S.sA>S.sB?tA:tB;h+='</canvas><div class="game-over-wrap"><h2>'+wT.flag+' '+wT.name+' remporte la séance !</h2><div class="final-score">'+S.sA+' - '+S.sB+'</div><div class="xp">+'+S.xp+' XP</div><div class="go-btns"><button class="go-replay" onclick="MC.replay()">Rejouer</button><button class="go-quit" onclick="MC.quit()">Quitter</button></div></div></div>';}
else{h+='</canvas></div>';}
h+='<div class="game-controls"><button onclick="MC.kb()">'+(S.skKb?'Mode souris':'Mode clavier')+'</button><div style="color:#64748b;font-size:.75rem">'+(S.ph==='aim'&&!S.skKb?'Dessine vers le but pour tirer':'')+'</div><button class="quit-btn" onclick="MC.quit()">Abandonner</button></div>';
if(S.skKb&&S.ph==='aim'){h+='<div class="keyboard-panel"><button onclick="MC.ksh(-1,-1,1.2)">↖</button><button onclick="MC.ksh(0,-1,1.5)">↑</button><button onclick="MC.ksh(1,-1,1.2)">↗</button><button onclick="MC.ksh(-0.7,0,1)">←</button><button onclick="MC.ksh(0,0,0.5)">●</button><button onclick="MC.ksh(0.7,0,1)">→</button><button onclick="MC.ksh(-1,-0.5,1.3)">↙</button><button onclick="MC.ksh(0,-0.5,1)">↓</button><button onclick="MC.ksh(1,-0.5,1.3)">↘</button></div>';}
app.innerHTML=h;
if(S.scr==='g'){cvs=document.getElementById('mc-c');if(cvs){ctx=cvs.getContext('2d');rszC();setupIn();rsB();startL();}}
}
function rszC(){var w=document.getElementById('cw');if(!cvs||!w)return;var r=w.getBoundingClientRect();var d=window.devicePixelRatio||1;cvs.width=r.width*d;cvs.height=r.height*d;ctx.scale(d,d);cvs.style.width=r.width+'px';cvs.style.height=r.height+'px';}
function gD(){if(!cvs)return{w:800,h:600};var r=cvs.getBoundingClientRect();return{w:r.width,h:r.height};}
function rsB(){var d=gD();ball={x:d.w/2,y:d.h-70,vx:0,vy:0,z:0,vz:0};kp={x:0,tx:0,dl:0,dv:false};traj=[];}
function setupIn(){cvs.onpointerdown=function(e){if(S.ph!=='aim')return;e.preventDefault();var r=cvs.getBoundingClientRect();sx=e.clientX-r.left;sy=e.clientY-r.top;pDown=true;traj=[{x:sx,y:sy}];window.addEventListener('pointermove',pMv);window.addEventListener('pointerup',pUp);window.addEventListener('pointercancel',pUp);};}
function pMv(e){if(!pDown||!cvs)return;var r=cvs.getBoundingClientRect();traj.push({x:e.clientX-r.left,y:e.clientY-r.top});}
function pUp(e){pDown=false;window.removeEventListener('pointermove',pMv);window.removeEventListener('pointerup',pUp);window.removeEventListener('pointercancel',pUp);if(!cvs)return;var r=cvs.getBoundingClientRect();var ex=e.clientX-r.left,ey=e.clientY-r.top;var dx=ex-sx,dy=ey-sy;var dist=Math.sqrt(dx*dx+dy*dy);if(dist<15){traj=[];return;}var pw=Math.min(dist/150,2);var an=Math.atan2(dy,dx);ball.vx=Math.cos(an)*SPD*pw;ball.vy=Math.sin(an)*SPD*pw;ball.vz=pw*6+Math.random()*2;var d=gD();var gx=(d.w-GW)/2;var df=Math.min(0.25+(S.sA+S.sB)*0.07,0.92);var tg=ex-(gx+GW/2);var er=(Math.random()-0.5)*(1-df)*GW*1.2;var rd=Math.random()>df?25:Math.floor(5+Math.random()*8);kp.tx=Math.max(-GW/2+KW/2,Math.min(GW/2-KW/2,tg+er));kp.dl=rd;traj=[];S.ph='shot';}
function draw(){if(!ctx||!cvs)return;var d=gD();ctx.clearRect(0,0,d.w,d.h);ctx.fillStyle='#22c55e';ctx.fillRect(0,0,d.w,d.h);ctx.fillStyle='#16a34a';for(var i=0;i<d.w;i+=120)ctx.fillRect(i,0,60,d.h);ctx.strokeStyle='rgba(255,255,255,.6)';ctx.lineWidth=3;var sy=d.h*.35;ctx.beginPath();ctx.moveTo(d.w*.15,sy);ctx.lineTo(d.w*.85,sy);ctx.stroke();ctx.lineWidth=4;ctx.strokeStyle='#fff';var gly=d.h*.18;ctx.beginPath();ctx.moveTo(d.w*.2,gly);ctx.lineTo(d.w*.8,gly);ctx.stroke();ctx.strokeStyle='rgba(255,255,255,.4)';ctx.lineWidth=2;ctx.strokeRect(d.w*.25,gly,d.w*.5,sy-gly);ctx.fillStyle='#fff';ctx.beginPath();ctx.arc(d.w/2,d.h*.55,4,0,Math.PI*2);ctx.fill();var gx=(d.w-GW)/2,gy=gly-GH;ctx.fillStyle='#e5e7eb';ctx.fillRect(gx-4,gy,4,GH);ctx.fillRect(gx+GW,gy,4,GH);ctx.fillRect(gx-4,gy-4,GW+8,4);ctx.strokeStyle='rgba(200,200,200,.5)';ctx.lineWidth=.8;for(var i=0;i<=GW;i+=12){ctx.beginPath();ctx.moveTo(gx+i,gy);ctx.lineTo(gx+i,gy+GH);ctx.stroke();}for(var i=0;i<=GH;i+=12){ctx.beginPath();ctx.moveTo(gx,gy+i);ctx.lineTo(gx+GW,gy+i);ctx.stroke();}var ksx=gx+GW/2+kp.x,ksy=gy+GH-KH/2;ctx.fillStyle='rgba(0,0,0,.15)';ctx.beginPath();ctx.ellipse(ksx,ksy+KH/2,KW/2,6,0,0,Math.PI*2);ctx.fill();var ti=ft(S.tm==='a'?S.tA:S.tB);ctx.fillStyle=ti.primary;if(kp.dv){ctx.save();ctx.translate(ksx,ksy);var da=(kp.x>0?1:-1)*Math.min(Math.abs(kp.x)*.02,.8);ctx.rotate(da);ctx.fillRect(-KW/2,-KH/2,KW,KH);ctx.fillStyle=ti.secondary;ctx.fillRect(-KW/2-4,-KH/2+4,6,10);ctx.fillRect(KW/2-2,-KH/2+4,6,10);ctx.restore();}else{ctx.fillRect(ksx-KW/2,ksy-KH/2,KW,KH);ctx.fillStyle=ti.secondary;ctx.fillRect(ksx-KW/2-4,ksy-KH/2+4,6,10);ctx.fillRect(ksx+KW/2-2,ksy-KH/2+4,6,10);}var sc=1+ball.z*.008,r=BR*sc;ctx.fillStyle='rgba(0,0,0,.2)';ctx.beginPath();ctx.ellipse(ball.x,ball.y+2,r*.8,r*.25,0,0,Math.PI*2);ctx.fill();ctx.beginPath();ctx.arc(ball.x,ball.y-ball.z,r,0,Math.PI*2);ctx.fillStyle='#fff';ctx.fill();ctx.strokeStyle='#1f2937';ctx.lineWidth=1.5;ctx.stroke();ctx.fillStyle='#1f2937';ctx.beginPath();ctx.arc(ball.x,ball.y-ball.z,r*.4,0,Math.PI*2);ctx.fill();if(S.ph==='aim'&&traj.length>1){ctx.strokeStyle='rgba(255,255,255,.6)';ctx.lineWidth=3;ctx.setLineDash([8,6]);ctx.lineCap='round';ctx.beginPath();ctx.moveTo(traj[0].x,traj[0].y);for(var i=1;i<traj.length;i++)ctx.lineTo(traj[i].x,traj[i].y);ctx.stroke();ctx.setLineDash([]);var la=traj[traj.length-1],pv=traj[traj.length-2]||traj[0];var an=Math.atan2(la.y-pv.y,la.x-pv.x);ctx.fillStyle='rgba(255,255,255,.8)';ctx.beginPath();ctx.moveTo(la.x+Math.cos(an)*12,la.y+Math.sin(an)*12);ctx.lineTo(la.x+Math.cos(an+2.5)*8,la.y+Math.sin(an+2.5)*8);ctx.lineTo(la.x+Math.cos(an-2.5)*8,la.y+Math.sin(an-2.5)*8);ctx.fill();}if(S.shk&&S.ph==='result'){ctx.strokeStyle='rgba(255,255,255,.3)';ctx.lineWidth=2;for(var i=1;i<=3;i++){ctx.beginPath();ctx.arc(gx+GW/2,gy+GH/2,20+i*15,0,Math.PI*2);ctx.stroke();}}}
function startL(){if(afId)cancelAnimationFrame(afId);function lp(){if(S.ph==='shot')phys();draw();afId=requestAnimationFrame(lp);}afId=requestAnimationFrame(lp);}
function phys(){var d=gD();var gx=(d.w-GW)/2,gy=d.h*.18-GH;ball.x+=ball.vx;ball.y+=ball.vy;ball.z+=ball.vz;ball.vz-=GRAV;if(ball.z<0){ball.z=0;ball.vz=-ball.vz*.4;}ball.vx*=.998;ball.vy*=.998;if(kp.dl>0){kp.dl--;}else{kp.dv=true;var dx=kp.tx-kp.x;var sp=3+Math.min((S.sA+S.sB)*.3,4);if(Math.abs(dx)>sp)kp.x+=Math.sign(dx)*sp;else kp.x=kp.tx;}var ksx=gx+GW/2+kp.x,ksy=gy+GH-KH/2;var dk=Math.sqrt(Math.pow(ball.x-ksx,2)+Math.pow(ball.y-ball.z-ksy,2));if(dk<KW/2+BR&&ball.z<KH+10&&ball.y<gy+GH+20){eS('saved');return;}if(ball.y<gy+GH+5&&ball.x>gx+5&&ball.x<gx+GW-5&&ball.z<GH){eS('goal');return;}if(ball.y<0||ball.x<-20||ball.x>d.w+20||ball.y>d.h+20){eS('miss');return;}}
function eS(r){S.ph='result';var m={goal:'⚽ BUT !!',saved:'🧤 Arrêté !',miss:'❌ Raté !'};S.msg=m[r];if(r==='goal'){if(S.tm==='a')S.sA++;else S.sB++;S.shk=true;setTimeout(function(){S.shk=false;},600);}S.shots.push({tm:S.tm,r:r});shotN++;if(S.sid)aRS(S.sid,S.tm,S.sh,r,shotN);setTimeout(function(){chkE(r);},1800);}
function chkE(lr){var tot=S.shots.length;var iS=S.mode==='solo';var cA=S.tm==='a'&&lr==='goal'?S.sA:S.sA;var cB=S.tm==='b'&&lr==='goal'?S.sB:S.sB;if(iS){if(tot>=5&&cA!==cB){fg(cA,cB);return;}if(tot>5&&cA!==cB){fg(cA,cB);return;}}else{if(tot>=10&&cA!==cB){fg(cA,cB);return;}if(tot>10&&cA!==cB){fg(cA,cB);return;}}nxt();}
function nxt(){if(S.mode==='solo'){S.tm='a';S.sh=(S.sh+1)%5;}else{S.tm=S.tm==='a'?'b':'a';if(S.tm==='b')S.sh=(S.sh+1)%5;}S.ph='transition';render();if(afId)cancelAnimationFrame(afId);setTimeout(function(){S.ph='aim';S.msg='';rsB();render();},2200);}
function fg(a,b){S.ph='finished';var w=a>b?(S.mode==='solo'?'player':'team_a'):'team_b';var xp=(a+b)*10;if(w==='player'||w==='team_a')xp+=50;if(a===5&&b===0)xp+=100;S.xp=xp;if(S.sid)aFS(S.sid,a,b,w);}
function aCS(m,a,b){fetch('/api/mini-cup/session',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({mode:m==='solo'?'solo_ai':'local_2v2',team_a:a,team_b:b})}).then(function(r){return r.json();}).then(function(d){S.sid=d.session_id||d.id||'';}).catch(function(){});}
function aRS(s,t,sh,r,o){fetch('/api/mini-cup/session/'+s+'/shot',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({team:t,shooter_index:sh,result:r,shot_order:o})}).catch(function(){});}
function aFS(s,a,b,w){fetch('/api/mini-cup/session/'+s+'/finish',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({score_a:a,score_b:b,winner:w})}).catch(function(){});}
window.MC={go:function(m){S.mode=m;S.scr='a';S.step='a';render();},pick:function(c){if(S.step==='a'){S.selA=c;S.step='b';S.scr='b';render();}else if(c!==S.selA){S.tA=S.selA;S.tB=c;S.scr='g';S.ph='aim';S.sA=0;S.sB=0;S.tm='a';S.sh=0;S.shots=[];shotN=0;S.xp=0;S.shk=false;aCS(S.mode,S.tA,S.tB);render();}},back:function(){S.scr='menu';render();},kb:function(){S.skKb=!S.skKb;render();},ksh:function(dx,dy,pw){if(S.ph!=='aim')return;var an=Math.atan2(dy,dx);ball.vx=Math.cos(an)*SPD*pw;ball.vy=Math.sin(an)*SPD*pw;ball.vz=pw*6;var d=gD();var gx=(d.w-GW)/2;var df=Math.min(0.25+(S.sA+S.sB)*.07,.92);var tg=dx*80;var er=(Math.random()-.5)*(1-df)*GW;var rd=Math.random()>df?25:Math.floor(5+Math.random()*8);kp.tx=Math.max(-GW/2+KW/2,Math.min(GW/2-KW/2,tg+er));kp.dl=rd;S.ph='shot';},replay:function(){S.ph='aim';S.shots=[];S.sA=0;S.sB=0;S.sh=0;S.tm='a';S.msg='';S.xp=0;shotN=0;S.shk=false;rsB();aCS(S.mode,S.tA,S.tB);render();},quit:function(){if(afId)cancelAnimationFrame(afId);S={scr:'menu',mode:'',step:'a',selA:'',tA:'',tB:'',ph:'aim',sA:0,sB:0,tm:'a',sh:0,shots:[],msg:'',sid:'',skKb:false,xp:0,shk:false};render();}};
window.addEventListener('resize',function(){if(cvs)rszC();});
render();
})();
</script>`
	return renderPage(c, "Mini Cup", content)
}

// ============================================================================
// HANDLERS — Mini Cup
// ============================================================================

// POST /api/mini-cup/session
func (h *Handler) CreateMiniCupSession(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	type Request struct {
		Mode  string `json:"mode"`
		TeamA string `json:"team_a"`
		TeamB string `json:"team_b"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Mode != "solo_ai" && req.Mode != "local_2v2" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid mode"})
	}

	sessionID := uuid.New().String()
	data, _ := json.Marshal(map[string]interface{}{
		"id":         sessionID,
		"user_id":    user.ID,
		"mode":       req.Mode,
		"team_a":     req.TeamA,
		"team_b":     req.TeamB,
		"score_a":    0,
		"score_b":    0,
		"status":     "in_progress",
		"xp_earned":  0,
		"created_at": time.Now(),
	})

	body, err := h.db.Insert("mini_cup_sessions", data, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return c.Status(201).JSON(fiber.Map{
		"session_id": sessionID,
		"mode":       req.Mode,
		"team_a":     req.TeamA,
		"team_b":     req.TeamB,
	})
}

// POST /api/mini-cup/session/:id/shot
func (h *Handler) RecordMiniCupShot(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	sessionID := c.Params("id")

	type Request struct {
		Team         string `json:"team"`
		ShooterIndex int    `json:"shooter_index"`
		Result       string `json:"result"`
		ShotOrder    int    `json:"shot_order"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	// Verify session ownership
	sessBody, err := h.db.Select("mini_cup_sessions",
		fmt.Sprintf("id=eq.%s&select=*", sessionID), true)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "session not found"})
	}
	var sessions []map[string]interface{}
	json.Unmarshal(sessBody, &sessions)
	if len(sessions) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "session not found"})
	}
	sessUserID := ""
	if uid, ok := sessions[0]["user_id"].(string); ok {
		sessUserID = uid
	}
	if sessUserID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// Record shot
	shotID := uuid.New().String()
	shotData, _ := json.Marshal(map[string]interface{}{
		"id":            shotID,
		"session_id":    sessionID,
		"team":          req.Team,
		"shooter_index": req.ShooterIndex,
		"result":        req.Result,
		"shot_order":    req.ShotOrder,
		"created_at":    time.Now(),
	})

	_, err = h.db.Insert("mini_cup_shots", shotData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Update score in real-time
	if req.Result == "goal" {
		scoreField := "score_a"
		if req.Team == "b" {
			scoreField = "score_b"
		}
		currentScore := 0
		if s, ok := sessions[0][scoreField].(float64); ok {
			currentScore = int(s)
		}
		updateData, _ := json.Marshal(map[string]interface{}{
			scoreField: currentScore + 1,
		})
		h.db.Update("mini_cup_sessions",
			fmt.Sprintf("id=eq.%s", sessionID), updateData, true)
	}

	return c.Status(201).JSON(fiber.Map{
		"shot_id": shotID,
		"result":  req.Result,
	})
}

// POST /api/mini-cup/session/:id/finish
func (h *Handler) FinishMiniCupSession(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	sessionID := c.Params("id")

	type Request struct {
		ScoreA int    `json:"score_a"`
		ScoreB int    `json:"score_b"`
		Winner string `json:"winner"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	xp := calculateMiniCupXP(req.ScoreA, req.ScoreB, req.Winner)
	now := time.Now()

	updateData, _ := json.Marshal(map[string]interface{}{
		"score_a":     req.ScoreA,
		"score_b":     req.ScoreB,
		"status":      "finished",
		"winner":      req.Winner,
		"xp_earned":   xp,
		"finished_at": now,
	})

	_, err := h.db.Update("mini_cup_sessions",
		fmt.Sprintf("id=eq.%s&user_id=eq.%s", sessionID, user.ID), updateData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"session_id": sessionID,
		"xp_earned":  xp,
		"finished":   true,
	})
}

// GET /api/mini-cup/leaderboard
func (h *Handler) GetMiniCupLeaderboard(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}

	body, err := h.db.Select("mini_cup_leaderboard",
		fmt.Sprintf("select=*&order=total_goals.desc,total_shots.asc&limit=%d", limit), false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var entries []map[string]interface{}
	json.Unmarshal(body, &entries)

	return c.JSON(entries)
}

// GET /api/mini-cup/sessions
func (h *Handler) GetMiniCupSessions(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	body, err := h.db.Select("mini_cup_sessions",
		fmt.Sprintf("user_id=eq.%s&order=created_at.desc&select=*", user.ID), false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var sessions []map[string]interface{}
	json.Unmarshal(body, &sessions)

	return c.JSON(sessions)
}

// ============================================================================
// HELPERS
// ============================================================================

// +10 XP per goal, +50 win, +100 perfect 5/0
func calculateMiniCupXP(scoreA, scoreB int, winner string) int {
	xp := (scoreA + scoreB) * 10
	if winner == "player" || winner == "team_a" {
		xp += 50
	}
	if scoreA == 5 && scoreB == 0 {
		xp += 100
	}
	return xp
}
