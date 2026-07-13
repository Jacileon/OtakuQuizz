package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"otaku-quiz-africa/internal/database"
)

func generateToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ============================================================
// PERSONAL QUIZ — CRUD
// ============================================================

func (h *Handler) PersonalQuizCreate(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	if c.Method() == "POST" {
		return h.personalQuizCreatePost(c)
	}

	checkBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("creator_id=eq.%s&status=eq.active&select=id", user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) > 0 {
		return c.Redirect("/personal-quiz/" + database.DBValue(existing[0]["id"]) + "/edit")
	}

	nickname := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		nickname = *user.Nickname
	}

	return h.renderVivatokCreate(c, nickname)
}

func (h *Handler) renderVivatokCreate(c *fiber.Ctx, nickname string) error {
	premade := []map[string]interface{}{
		{"q": "Quel contenu envahit le plus ton roll photo ?", "o": []string{"🤳 Selfies", "😂 Memes", "📲 Screenshots", "🍕 Nourriture", "🐶 Animaux", "👕 Tenues"}},
		{"q": "Quelle tendance ne testerais-tu jamais ?", "o": []string{"🕺 Danse TikTok", "🌶️ Défi piment", "💇 Coiffure bizarre", "🤡 Vidéos farces", "🏆 Challenges", "✂️ Mode DIY"}},
		{"q": "C'est quoi ton objectif de célébrité ?", "o": []string{"🎤 Musique", "⚽ Sport", "🎭 Comédie", "🎬 TikTok", "📺 YouTube", "🎨 Art"}},
		{"q": "C'est quoi ton plus grand flex ?", "o": []string{"👟 Sneakers", "📱 Tech", "📈 Abonnés", "📝 Notes", "🎭 Talent", "✈️ Voyages"}},
		{"q": "Quel rôle aurais-tu dans un film d'horreur ?", "o": []string{"💪 Héros", "💀 Première victime", "🤣 Comic relief", "🔪 Vilain secret", "😎 Sidekick", "🏃 Survivant"}},
		{"q": "C'est quoi ton excuse préférée ?", "o": []string{"📚 Devoirs", "👨‍👩‍👧 Famille", "🤒 Malade", "💸 Pas d'argent", "📵 Batterie", "🤔 Oubli"}},
		{"q": "Quelle matière sauterais-tu en cours ?", "o": []string{"🔢 Maths", "📜 Histoire", "🏃 Sport", "🧪 Science", "📝 Anglais", "🎨 Art"}},
		{"q": "Tu postes tes meilleures photos où ?", "o": []string{"📸 Instagram", "📱 BeReal", "🎥 TikTok", "👻 Snapchat", "🐦 Twitter", "🙈 Nulle part"}},
		{"q": "Quel emoji décrit ton vibe ?", "o": []string{"😎 Cool", "🥺 Mignon", "🤪 Sauvage", "😇 Innocent", "🙄 Agacé", "🔥 Feu"}},
		{"q": "Tu texterais qui à 3h du mat ?", "o": []string{"👯 Bestie", "💘 Crush", "😴 Personne", "👪 Famille", "🙄 Ex", "🎭 Groupe"}},
		{"q": "C'est quoi ton plus grand cringe ?", "o": []string{"💃 Parents qui dansent", "📱 Anciens posts", "📸 Photos de classe", "😖 Gênant public", "💌 Lettres d'amour", "👶 Phases d'enfance"}},
		{"q": "Quel lookalike célèbre tu voudrais ?", "o": []string{"🎤 Star pop", "🎬 Star cinéma", "🏆 Athlète", "📱 YouTuber", "💅 Mannequin", "✨ Personnage animé"}},
		{"q": "C'est quoi la recherche la plus bizarre de ton historique ?", "o": []string{"👽 Théories complot", "🎵 Paroles", "🧠 Faits random", "👀 Potins stars", "🎬 Théories fans", "🔨 Astuces DIY"}},
		{"q": "Quel style musical tu écoutes le plus ?", "o": []string{"🎵 Pop", "🎤 Hip-hop/Rap", "🎸 Rock", "🎧 Électro", "🎹 Indie", "🤠 Country"}},
		{"q": "Quel est le statut de ta batterie là ?", "o": []string{"🔋 100%", "📶 À moitié", "⚡ Low power", "😱 Presque mort", "💀 Toujours à plat", "🔌 En charge"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai déguisé mon chien pour un goûter thé ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait semblant d'être malade pour sauter l'école ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai créé un langage secret avec un ami ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai envoyé une note anonyme à mon crush ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai trop ri à une blague pas drôle de mon crush ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait semblant d'aimer un cadeau que je détestais ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait un blanc devant mon crush ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai essayé une danse TikTok seul(e) et j'ai raté ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai chanté sous la douche en pensant être seul(e) ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai mangé quelque chose de tellement épicé que j'ai pleuré ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai envoyé un message gênant à la mauvaise personne ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai marché dans une vitre en pensant que c'était ouvert ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai ri si fort dans une pièce silencieuse que tout le monde a regardé ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai déjà secrètement aimé un ami ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai pris un selfie avec une célébrité ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai porté mes vêtements à l'envers sans m'en rendre compte ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai convaincu quelqu'un d'une histoire totalement inventée ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai eu tellement peur en voyant un film que j'ai dormi avec la lumière ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai veillé toute la nuit pour bingewatcher une série ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai entré dans la mauvaise classe en toute confiance ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait semblant d'aimer des chansons pour être dans le coup ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "Quel film voudrais-tu être la star ?", "o": []string{"🎬 Action", "😂 Comédie", "👻 Horreur", "💖 Romance", "👽 Sci-Fi", "🧙‍♂️ Fantaisie", "🎨 Animation"}},
		{"q": "Quel personnage de fiction aimerais-tu comme ami ?", "o": []string{"🧙‍♂️ Harry Potter", "🏹 Katniss", "🦾 Iron Man", "🕷️ Spider-Man", "❄️ Elsa", "📚 Hermione", "⚡ Pikachu"}},
		{"q": "C'est quoi ton snack de minuit préféré ?", "o": []string{"🍕 Pizza", "🍿 Popcorn", "🍫 Chocolat", "🍓 Fruit", "🍪 Biscuits", "🍨 Glace", "🍟 Chips"}},
		{"q": "Quel animal de compagnie inhabituel voudrais-tu ?", "o": []string{"🐰 Lapin", "🐹 Hamster", "🐢 Tortue", "🦔 Hérisson", "🦜 Perroquet", "🦎 Lézard"}},
		{"q": "Quelle créature mythique te fascine le plus ?", "o": []string{"🐉 Dragon", "🦄 Licorne", "🔥 Phénix", "🧜‍♀️ Sirène", "🌕🐺 Loup-garou"}},
		{"q": "Quel sport extrême voudrais-tu essayer ?", "o": []string{"🪂 Parachute", "🏞️ Élastique", "🧗 Escalade", "🚣 Rafting", "🪁 Parapente", "🏂 Snowboard", "🤸 Parkour"}},
		{"q": "Quelle fête à thème organiserais-tu ?", "o": []string{"🦸‍♂️ Super-héros", "🏖️ Plage", "🚀 Spatiale", "🎤 Années 2000", "🎭 Bal masqué", "🎵 TikTok", "🎆 Néon"}},
		{"q": "Quel gadget tech utilises-tu le plus ?", "o": []string{"📱 Smartphone", "💻 PC", "📲 Tablette", "⌚ Montre", "🎮 Console", "📖 E-reader", "🔊 Enceinte"}},
		{"q": "Quelle forme d'art aimerais-tu apprendre ?", "o": []string{"🎨 Peinture", "🗿 Sculpture", "💻 Art numérique", "🏺 Poterie", "🩰 Danse", "🎹 Musique", "📷 Photo"}},
		{"q": "Que ferais-tu avec une machine à voyager dans le temps ?", "o": []string{"🦖 Dinosaures", "👑 Histoire", "🌆 Futur", "🏛️ Civilisations", "🎤 Concerts", "🚨 Catastrophes", "💎 Trésors"}},
		{"q": "Quel aliment inhabituel voudrais-tu essayer ?", "o": []string{"🦗 Insectes frits", "🐙 Poulpe", "🍈 Durian", "🐌 Escargot", "🥚 Œuf centenaire", "🐊 Alligator", "🥗 Algues"}},
		{"q": "Quel animal sauvage voudrais-tu voir de près ?", "o": []string{"🦁 Lion", "🐬 Dauphin", "🦅 Aigle", "🐼 Panda", "🐘 Éléphant", "🦈 Requin", "🦘 Kangourou"}},
		{"q": "Ta boisson préférée par temps chaud ?", "o": []string{"🍋 Limonade", "🍹 Thé glacé", "🥤 Smoothie", "🥤 Soda", "☕ Café glacé", "💧 Eau", "🍇 Punch"}},
		{"q": "Ta glace préférée ?", "o": []string{"🍦 Vanille", "🍫 Chocolat", "🍓 Fraise", "🍬 Chewing-gum", "🍪 Pâte à cookie", "🍯 Caramel salé"}},
		{"q": "Ta garniture de pizza idéale ?", "o": []string{"🍕 Pepperoni", "🧀 Fromage supp.", "🌭 Saucisse", "🍖 Jambon", "🍗 Poulet", "🫒 Olives"}},
	}

	questionsJSON, _ := json.Marshal(premade)
	questionsStr := strings.ReplaceAll(string(questionsJSON), "</", "<\\/")

	content := `
<div class="vtk-page" style="align-items:flex-start;padding-top:30px">
<div class="vtk-card" style="max-width:520px">

<div class="vtk-hero">
<div style="font-size:3rem;margin-bottom:8px">🏆</div>
<h1 style="font-size:1.6rem">Que savent tes amis de toi ?</h1>
<p style="color:rgba(255,255,255,.4);margin:4px 0 20px">Réponds aux questions et découvre qui te connaît le mieux !</p>
</div>

<div class="vtk-steps" style="display:flex;gap:12px;justify-content:center;margin-bottom:24px">
<div style="flex:1;text-align:center;padding:10px;background:rgba(255,255,255,.04);border-radius:12px;border:1px solid rgba(255,255,255,.08)"><div style="font-size:1.5rem">📝</div><div style="font-size:.75rem;color:rgba(255,255,255,.5);margin-top:4px">Crée ton défi</div></div>
<div style="flex:1;text-align:center;padding:10px;background:rgba(255,255,255,.04);border-radius:12px;border:1px solid rgba(255,255,255,.08)"><div style="font-size:1.5rem">📤</div><div style="font-size:.75rem;color:rgba(255,255,255,.5);margin-top:4px">Envoie à tes amis</div></div>
<div style="flex:1;text-align:center;padding:10px;background:rgba(255,255,255,.04);border-radius:12px;border:1px solid rgba(255,255,255,.08)"><div style="font-size:1.5rem">🏆</div><div style="font-size:.75rem;color:rgba(255,255,255,.5);margin-top:4px">Vérifie les réponses</div></div>
</div>

<div id="name-screen">
<div style="background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:16px;padding:24px;text-align:center;margin-bottom:20px">
<p style="margin:0 0 12px;font-size:.95rem">Entre ton prénom pour commencer</p>
<input type="text" id="creator-name" value="` + nickname + `" placeholder="Ton prénom..." class="vtk-input" style="text-align:center;font-size:1.1rem;font-weight:600;margin-bottom:4px">
<p style="color:rgba(255,255,255,.3);font-size:.75rem;margin:0">Ce nom sera utilisé dans le lien de partage</p>
</div>
<button class="vtk-btn-play" onclick="startCreate()">Commencer 👉</button>
</div>

<div id="quiz-area" style="display:none">
<div class="vtk-progress" id="vtk-progress"></div>
<div id="question-container"></div>
</div>

<div id="share-screen" style="display:none;text-align:center">
<div style="font-size:3rem;margin-bottom:8px">🎉</div>
<h2 style="margin:0 0 4px">Super, ` + nickname + ` !</h2>
<p style="color:rgba(255,255,255,.4);margin:0 0 20px">Ton défi est prêt ! Partage-le avec tes amis 👇</p>
<div id="share-link-area" style="background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:16px;padding:16px;margin-bottom:16px">
<div style="display:flex;gap:8px;align-items:center;background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:12px;padding:4px 14px">
<input type="text" id="share-link" readonly style="flex:1;padding:10px 4px;background:transparent;border:none;color:#fff;font-size:.85rem;outline:none;text-align:center">
<button class="vtk-btn-icon" onclick="copyShareLink()">📋</button>
</div>
</div>
<div id="share-buttons" style="display:flex;gap:8px;flex-wrap:wrap;justify-content:center;margin-bottom:16px"></div>
<p style="color:rgba(255,255,255,.3);font-size:.8rem">Vérifie qui a répondu sur la page de résultats !</p>
<a id="results-link" href="#" class="vtk-btn-ghost" style="font-size:.9rem;display:inline-block;margin-top:4px">📊 Voir les résultats →</a>
</div>

</div>
</div>

<style>
@keyframes vtkFadeIn{from{opacity:0;transform:translateY(16px)}to{opacity:1;transform:translateY(0)}}
@keyframes vtkPulse{0%{transform:scale(1)}50%{transform:scale(1.03)}100%{transform:scale(1)}}
@keyframes vtkPop{0%{transform:scale(.8);opacity:0}100%{transform:scale(1);opacity:1}}
@keyframes vtkSlideIn{from{opacity:0;transform:translateX(40px)}to{opacity:1;transform:translateX(0)}}
@keyframes fadeIn{from{opacity:0;transform:translateY(10px)}to{opacity:1;transform:translateY(0)}}
.vtk-question{display:none;animation:fadeIn .3s ease}
.vtk-question.active{display:block}
.vtk-q-num{color:rgba(255,255,255,.3);font-size:.8rem;margin-bottom:8px}
.vtk-q-text{font-size:1.15rem;font-weight:600;margin:0 0 16px;line-height:1.4}
.vtk-opt-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin:16px 0}
.vtk-opt-card{display:flex;flex-direction:column;align-items:center;justify-content:center;padding:16px 8px;border-radius:16px;border:1px solid rgba(255,255,255,.08);background:rgba(255,255,255,.04);color:#fff;cursor:pointer;transition:all .2s;font-family:inherit;aspect-ratio:1}
.vtk-opt-card:hover{border-color:rgba(102,126,234,.4);background:rgba(102,126,234,.08);transform:translateY(-2px)}
.vtk-opt-card:active{transform:scale(.95)}
.vtk-opt-card.selected{border-color:#667eea;background:rgba(102,126,234,.15);animation:vtkPulse .3s ease}
.vtk-opt-emoji{font-size:2.2rem;margin-bottom:6px}
.vtk-opt-label{font-size:.8rem;font-weight:600;text-align:center;line-height:1.2}
.vtk-btn-skip{background:none;border:none;color:rgba(255,255,255,.3);font-size:.85rem;cursor:pointer;margin-top:8px;padding:8px 16px;border-radius:8px;transition:color .2s}
.vtk-btn-skip:hover{color:rgba(255,255,255,.6)}
.vtk-sbtn{display:inline-flex;align-items:center;gap:8px;padding:10px 20px;border-radius:12px;border:none;color:#fff;font-size:.9rem;font-weight:600;cursor:pointer;text-decoration:none;transition:opacity .2s}
.vtk-sbtn:hover{opacity:.9}
</style>
<script id="pq-data" type="application/json">` + questionsStr + `</script>
<script>
var premade=JSON.parse(document.getElementById('pq-data').textContent);
var current=0;var answers=[];

function startCreate(){
var name=document.getElementById('creator-name').value.trim();
if(!name){alert('Entre ton prénom !');return;}
document.getElementById('name-screen').style.display='none';
document.getElementById('quiz-area').style.display='block';
renderQuestion();
}

function renderQuestion(){
if(current>=premade.length){finishCreate();return;}
var q=premade[current];
var h='<div class="vtk-question active"><p class="vtk-q-num">Question '+(current+1)+'/'+premade.length+'</p><p class="vtk-q-text">'+q.q+'</p>';
h+='<div class="vtk-opt-grid">';
q.o.forEach(function(o,i){
var parts=o.match(/^(\S+)\s+(.+)$/);
var emoji=parts?parts[1]:'❓';
var label=parts?parts[2]:o;
h+='<button class="vtk-opt-card" onclick="selectOpt(this,'+i+')"><span class="vtk-opt-emoji">'+emoji+'</span><span class="vtk-opt-label">'+label+'</span></button>';
});
h+='</div>';
h+='<div style="text-align:center;margin-top:12px"><button class="vtk-btn-skip" onclick="skipQuestion()">Sauter →</button></div></div>';
document.getElementById('question-container').innerHTML=h;
var pct=((current)/premade.length*100);
document.getElementById('vtk-progress').innerHTML='<div class="vtk-bar"><div class="vtk-bar-fill" style="width:'+pct+'%"></div></div><p style="color:rgba(255,255,255,.3);font-size:.78rem;margin-top:6px">'+(current+1)+'/'+premade.length+'</p>';
}

function selectOpt(el,i){document.querySelectorAll('.vtk-opt-card').forEach(function(e){e.classList.remove('selected')});el.classList.add('selected');answers[current]=i;setTimeout(function(){current++;renderQuestion();},200);}

function skipQuestion(){answers[current]=-1;current++;renderQuestion();}

function finishCreate(){
document.getElementById('quiz-area').style.display='none';
document.getElementById('share-screen').style.display='block';
var name=document.getElementById('creator-name').value.trim();
fetch('/personal-quiz/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({nickname:name,answers:answers})}).then(function(r){return r.json()}).then(function(d){
if(d.success){
var link=window.location.origin+'/quiz/personal/'+d.token;
document.getElementById('share-link').value=link;
document.getElementById('results-link').href='/quiz/personal/'+d.token+'/results';
var btns=document.getElementById('share-buttons');
var shares=[
{label:'WhatsApp',icon:'https://vivatok.com/meill/images/whats2.png',url:'https://wa.me/?text='+encodeURIComponent('🏆 Devine qui est le/la plus fort(e) ! 🔥\n\n'+name+' a créé un quiz sur lui/elle ! 😍\nRéponds et découvre si tu le/la connais vraiment !\n\n👉 '+link)},
{label:'Snapchat',icon:'https://vivatok.com/meill/images/snapchat3.png',url:'#'},
{label:'Instagram',icon:'https://vivatok.com/meill/images/insta_novo.png',url:'#'},
{label:'Copier',icon:'',url:'javascript:copyFinalLink()'},
];
shares.forEach(function(s){
var a=document.createElement('a');a.href=s.url;a.target='_blank';a.className='vtk-sbtn';
if(s.icon){var img=document.createElement('img');img.src=s.icon;img.style.width='24px';img.style.height='24px';a.appendChild(img);}
var span=document.createElement('span');span.textContent=s.label;a.appendChild(span);
btns.appendChild(a);
});
}else alert(d.error||'Erreur lors de la création');
});
}

function copyShareLink(){
var t=document.getElementById('share-link');
navigator.clipboard.writeText(t.value).then(function(){var el=document.querySelector('.vtk-toast');if(!el){el=document.createElement('div');el.className='vtk-toast';document.querySelector('.vtk-card').appendChild(el);}el.textContent='✅ Lien copié !';el.style.display='block';setTimeout(function(){el.style.display='none';},2000);});
}
function copyFinalLink(){copyShareLink();}
</script>`

	return renderPage(c, "Crée ton défi", content)
}

func (h *Handler) personalQuizCreatePost(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	type Request struct {
		Nickname string `json:"nickname"`
		Answers  []int  `json:"answers"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Requête invalide"})
	}

	premade := []map[string]interface{}{
		{"q": "Quel contenu envahit le plus ton roll photo ?", "o": []string{"🤳 Selfies", "😂 Memes", "📲 Screenshots", "🍕 Nourriture", "🐶 Animaux", "👕 Tenues"}},
		{"q": "Quelle tendance ne testerais-tu jamais ?", "o": []string{"🕺 Danse TikTok", "🌶️ Défi piment", "💇 Coiffure bizarre", "🤡 Vidéos farces", "🏆 Challenges", "✂️ Mode DIY"}},
		{"q": "C'est quoi ton objectif de célébrité ?", "o": []string{"🎤 Musique", "⚽ Sport", "🎭 Comédie", "🎬 TikTok", "📺 YouTube", "🎨 Art"}},
		{"q": "C'est quoi ton plus grand flex ?", "o": []string{"👟 Sneakers", "📱 Tech", "📈 Abonnés", "📝 Notes", "🎭 Talent", "✈️ Voyages"}},
		{"q": "Quel rôle aurais-tu dans un film d'horreur ?", "o": []string{"💪 Héros", "💀 Première victime", "🤣 Comic relief", "🔪 Vilain secret", "😎 Sidekick", "🏃 Survivant"}},
		{"q": "C'est quoi ton excuse préférée ?", "o": []string{"📚 Devoirs", "👨‍👩‍👧 Famille", "🤒 Malade", "💸 Pas d'argent", "📵 Batterie", "🤔 Oubli"}},
		{"q": "Quelle matière sauterais-tu en cours ?", "o": []string{"🔢 Maths", "📜 Histoire", "🏃 Sport", "🧪 Science", "📝 Anglais", "🎨 Art"}},
		{"q": "Tu postes tes meilleures photos où ?", "o": []string{"📸 Instagram", "📱 BeReal", "🎥 TikTok", "👻 Snapchat", "🐦 Twitter", "🙈 Nulle part"}},
		{"q": "Quel emoji décrit ton vibe ?", "o": []string{"😎 Cool", "🥺 Mignon", "🤪 Sauvage", "😇 Innocent", "🙄 Agacé", "🔥 Feu"}},
		{"q": "Tu texterais qui à 3h du mat ?", "o": []string{"👯 Bestie", "💘 Crush", "😴 Personne", "👪 Famille", "🙄 Ex", "🎭 Groupe"}},
		{"q": "C'est quoi ton plus grand cringe ?", "o": []string{"💃 Parents qui dansent", "📱 Anciens posts", "📸 Photos de classe", "😖 Gênant public", "💌 Lettres d'amour", "👶 Phases d'enfance"}},
		{"q": "Quel lookalike célèbre tu voudrais ?", "o": []string{"🎤 Star pop", "🎬 Star cinéma", "🏆 Athlète", "📱 YouTuber", "💅 Mannequin", "✨ Personnage animé"}},
		{"q": "C'est quoi la recherche la plus bizarre de ton historique ?", "o": []string{"👽 Théories complot", "🎵 Paroles", "🧠 Faits random", "👀 Potins stars", "🎬 Théories fans", "🔨 Astuces DIY"}},
		{"q": "Quel style musical tu écoutes le plus ?", "o": []string{"🎵 Pop", "🎤 Hip-hop/Rap", "🎸 Rock", "🎧 Électro", "🎹 Indie", "🤠 Country"}},
		{"q": "Quel est le statut de ta batterie là ?", "o": []string{"🔋 100%", "📶 À moitié", "⚡ Low power", "😱 Presque mort", "💀 Toujours à plat", "🔌 En charge"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai déguisé mon chien pour un goûter thé ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait semblant d'être malade pour sauter l'école ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai créé un langage secret avec un ami ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai envoyé une note anonyme à mon crush ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai trop ri à une blague pas drôle de mon crush ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait semblant d'aimer un cadeau que je détestais ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait un blanc devant mon crush ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai essayé une danse TikTok seul(e) et j'ai raté ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai chanté sous la douche en pensant être seul(e) ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai mangé quelque chose de tellement épicé que j'ai pleuré ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai envoyé un message gênant à la mauvaise personne ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai marché dans une vitre en pensant que c'était ouvert ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai ri si fort dans une pièce silencieuse que tout le monde a regardé ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai déjà secrètement aimé un ami ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai pris un selfie avec une célébrité ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai porté mes vêtements à l'envers sans m'en rendre compte ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai convaincu quelqu'un d'une histoire totalement inventée ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai eu tellement peur en voyant un film que j'ai dormi avec la lumière ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai veillé toute la nuit pour bingewatcher une série ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai entré dans la mauvaise classe en toute confiance ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "🎯 Vrai ou 🦄 Faux : J'ai fait semblant d'aimer des chansons pour être dans le coup ?", "o": []string{"🎯 Vrai", "🦄 Faux"}},
		{"q": "Quel film voudrais-tu être la star ?", "o": []string{"🎬 Action", "😂 Comédie", "👻 Horreur", "💖 Romance", "👽 Sci-Fi", "🧙‍♂️ Fantaisie", "🎨 Animation"}},
		{"q": "Quel personnage de fiction aimerais-tu comme ami ?", "o": []string{"🧙‍♂️ Harry Potter", "🏹 Katniss", "🦾 Iron Man", "🕷️ Spider-Man", "❄️ Elsa", "📚 Hermione", "⚡ Pikachu"}},
		{"q": "C'est quoi ton snack de minuit préféré ?", "o": []string{"🍕 Pizza", "🍿 Popcorn", "🍫 Chocolat", "🍓 Fruit", "🍪 Biscuits", "🍨 Glace", "🍟 Chips"}},
		{"q": "Quel animal de compagnie inhabituel voudrais-tu ?", "o": []string{"🐰 Lapin", "🐹 Hamster", "🐢 Tortue", "🦔 Hérisson", "🦜 Perroquet", "🦎 Lézard"}},
		{"q": "Quelle créature mythique te fascine le plus ?", "o": []string{"🐉 Dragon", "🦄 Licorne", "🔥 Phénix", "🧜‍♀️ Sirène", "🌕🐺 Loup-garou"}},
		{"q": "Quel sport extrême voudrais-tu essayer ?", "o": []string{"🪂 Parachute", "🏞️ Élastique", "🧗 Escalade", "🚣 Rafting", "🪁 Parapente", "🏂 Snowboard", "🤸 Parkour"}},
		{"q": "Quelle fête à thème organiserais-tu ?", "o": []string{"🦸‍♂️ Super-héros", "🏖️ Plage", "🚀 Spatiale", "🎤 Années 2000", "🎭 Bal masqué", "🎵 TikTok", "🎆 Néon"}},
		{"q": "Quel gadget tech utilises-tu le plus ?", "o": []string{"📱 Smartphone", "💻 PC", "📲 Tablette", "⌚ Montre", "🎮 Console", "📖 E-reader", "🔊 Enceinte"}},
		{"q": "Quelle forme d'art aimerais-tu apprendre ?", "o": []string{"🎨 Peinture", "🗿 Sculpture", "💻 Art numérique", "🏺 Poterie", "🩰 Danse", "🎹 Musique", "📷 Photo"}},
		{"q": "Que ferais-tu avec une machine à voyager dans le temps ?", "o": []string{"🦖 Dinosaures", "👑 Histoire", "🌆 Futur", "🏛️ Civilisations", "🎤 Concerts", "🚨 Catastrophes", "💎 Trésors"}},
		{"q": "Quel aliment inhabituel voudrais-tu essayer ?", "o": []string{"🦗 Insectes frits", "🐙 Poulpe", "🍈 Durian", "🐌 Escargot", "🥚 Œuf centenaire", "🐊 Alligator", "🥗 Algues"}},
		{"q": "Quel animal sauvage voudrais-tu voir de près ?", "o": []string{"🦁 Lion", "🐬 Dauphin", "🦅 Aigle", "🐼 Panda", "🐘 Éléphant", "🦈 Requin", "🦘 Kangourou"}},
		{"q": "Ta boisson préférée par temps chaud ?", "o": []string{"🍋 Limonade", "🍹 Thé glacé", "🥤 Smoothie", "🥤 Soda", "☕ Café glacé", "💧 Eau", "🍇 Punch"}},
		{"q": "Ta glace préférée ?", "o": []string{"🍦 Vanille", "🍫 Chocolat", "🍓 Fraise", "🍬 Chewing-gum", "🍪 Pâte à cookie", "🍯 Caramel salé"}},
		{"q": "Ta garniture de pizza idéale ?", "o": []string{"🍕 Pepperoni", "🧀 Fromage supp.", "🌭 Saucisse", "🍖 Jambon", "🍗 Poulet", "🫒 Olives"}},
	}

	token := strings.ToLower(strings.ReplaceAll(req.Nickname, " ", ""))
	if token == "" {
		token = generateToken(8)
	}
	title := fmt.Sprintf("Connais-tu bien %s ?", req.Nickname)

	data, _ := json.Marshal(map[string]interface{}{
		"creator_id": user.ID,
		"token":      token,
		"title":      title,
		"status":     "active",
	})
	body, err := h.db.Insert("personal_quizzes", data, true)
	if err != nil || len(body) < 10 {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur création quiz"})
	}
	var rows []map[string]interface{}
	json.Unmarshal(body, &rows)
	if len(rows) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur création quiz"})
	}
	quizID := database.DBValue(rows[0]["id"])

	alreadySkipped := make(map[int]bool)
	for qi, ansIdx := range req.Answers {
		if ansIdx < 0 || ansIdx >= len(premade[qi]["o"].([]string)) {
			alreadySkipped[qi] = true
			continue
		}
		q := premade[qi]
		qText := q["q"].(string)
		opts := q["o"].([]string)
		correctAnswer := opts[ansIdx]

		optionsJSON, _ := json.Marshal(opts)
		qData, _ := json.Marshal(map[string]interface{}{
			"quiz_id":        quizID,
			"question_text":  qText,
			"type":           "qcm",
			"correct_answer": correctAnswer,
			"options":        string(optionsJSON),
			"sort_order":     qi,
		})
		h.db.Insert("personal_quiz_questions", qData, true)
	}

	return c.JSON(fiber.Map{"success": true, "token": token})
}

func (h *Handler) PersonalQuizEdit(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	quizID := c.Params("id")

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&creator_id=eq.%s&select=*", quizID, user.ID), true)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)
	if len(quizzes) == 0 {
		return c.Status(404).SendString("Quiz introuvable")
	}
	quiz := quizzes[0]

	title, _ := quiz["title"].(string)
	token, _ := quiz["token"].(string)
	status, _ := quiz["status"].(string)

	questionsBody, _ := h.db.Select("personal_quiz_questions",
		fmt.Sprintf("quiz_id=eq.%s&select=*&order=sort_order.asc", quizID), true)
	var questions []map[string]interface{}
	json.Unmarshal(questionsBody, &questions)

	questionsHTML := ""
	for i, q := range questions {
		qID := database.DBValue(q["id"])
		qText := database.DBValue(q["question_text"])
		qType := database.DBValue(q["type"])
		qAnswer := database.DBValue(q["correct_answer"])
		qOptions := database.DBValue(q["options"])

		opts := []string{}
		if qOptions != "null" && qOptions != "" {
			json.Unmarshal([]byte(qOptions), &opts)
		}

		gridCards := ""
		for _, o := range opts {
			parts := strings.SplitN(o, " ", 2)
			emoji := "❓"
			label := o
			if len(parts) == 2 {
				emoji = parts[0]
				label = parts[1]
			}
			isSelected := ""
			if o == qAnswer {
				isSelected = " selected"
			}
			gridCards += fmt.Sprintf(`<button type="button" class="vtk-opt-card%s" onclick="setCorrectCard(this,'%s','%s')"><span class="vtk-opt-emoji">%s</span><span class="vtk-opt-label">%s</span></button>`,
				isSelected, qID, htmlAttr(o), emoji, label)
		}

		if qType == "vrai_faux" {
			gridCards = ""
			for _, vf := range []string{"🎯 Vrai", "🦄 Faux"} {
				isSelected := ""
				if vf == qAnswer || strings.TrimPrefix(vf, "🎯 ") == qAnswer || strings.TrimPrefix(vf, "🦄 ") == qAnswer {
					isSelected = " selected"
				}
				parts := strings.SplitN(vf, " ", 2)
				gridCards += fmt.Sprintf(`<button type="button" class="vtk-opt-card%s" onclick="setCorrectCard(this,'%s','%s')"><span class="vtk-opt-emoji">%s</span><span class="vtk-opt-label">%s</span></button>`,
					isSelected, qID, htmlAttr(parts[1]), parts[0], parts[1])
			}
		}

		questionsHTML += fmt.Sprintf(`
<div class="vtk-ques-card" data-id="%s" data-correct="%s">
<div class="vtk-ques-header">
<span class="vtk-ques-num">#%d</span>
<span class="vtk-ques-type">%s</span>
<button type="button" class="vtk-ques-del" onclick="deleteQuestion('%s')">✕</button>
</div>
<p style="font-size:1rem;font-weight:600;margin:0 0 12px;line-height:1.4">%s</p>
<div class="vtk-opt-grid">%s</div>
</div>`, qID, htmlAttr(qAnswer), i+1, qType, qID, qText, gridCards)
	}

	content := `
<script>
var QUIZ_ID='` + quizID + `';
var CSRF='` + token + `';

function copyShareLink(){
    var fullLink=window.location.origin+'/quiz/personal/'+document.getElementById('share-token').value;
    var tmp=document.createElement('textarea');
    tmp.value=fullLink;
    tmp.style.position='fixed';tmp.style.opacity='0';
    document.body.appendChild(tmp);
    tmp.select();document.execCommand('copy');
    document.body.removeChild(tmp);
    alert('Lien copié !');
}

function onTokenChange(){
    document.getElementById('save-link-btn').style.display=document.getElementById('share-token').value!='` + token + `'?'inline-block':'none';
}

function saveShareLink(){
    var token=document.getElementById('share-token').value.trim();
    if(!token){alert('Le lien ne peut pas être vide');return;}
    fetch('/api/personal-quiz/'+QUIZ_ID,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({token:token})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){document.getElementById('save-link-btn').style.display='none';alert('Lien mis à jour !');}
        else alert(d.error||'Erreur');
    });
}

function addQuestion(){
	fetch('/api/personal-quiz/'+QUIZ_ID+'/questions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({question_text:'Nouvelle question',type:'qcm',correct_answer:'',options:['Option 1','Option 2','Option 3','Option 4']})}).then(function(r){return r.json()}).then(function(d){if(d.success)location.reload();});
}

function deleteQuestion(qid){
	if(!confirm('Supprimer cette question ?'))return;
	fetch('/api/personal-quiz/'+QUIZ_ID+'/questions/'+qid,{method:'DELETE'}).then(function(r){return r.json()}).then(function(d){if(d.success)location.reload();});
}

function updateQuestion(qid,field,value){
	fetch('/api/personal-quiz/'+QUIZ_ID+'/questions/'+qid,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({[field]:value})});
}

function renderQCMOpts(qid,opts,correct){}

function renderVFOpts(qid,correct){}

function toggleOpts(qid,type){}

function addOptRow(qid,qtype,val,isCorrect){}

function setCorrect(qid,val){
	document.querySelector('.vtk-ques-card[data-id="'+qid+'"]').setAttribute('data-correct',val);
	updateQuestion(qid,'correct_answer',val);
}

function setCorrectQCM(qid,radio){}

function saveOptions(el,qid){}

function setCorrectCard(el,qid,val){
	var card=el.closest('.vtk-ques-card');
	card.querySelectorAll('.vtk-opt-card').forEach(function(b){b.classList.remove('selected')});
	el.classList.add('selected');
	card.setAttribute('data-correct',val);
	updateQuestion(qid,'correct_answer',val);
}

function checkAllAnswered(){
	var qs=document.querySelectorAll('.vtk-ques-card');
	for(var i=0;i<qs.length;i++){
		if(!qs[i].getAttribute('data-correct')){
			alert('Question #'+(i+1)+' n\'a pas de bonne réponse attribuée.');
			return false;
		}
	}
	return true;
}
function archiveQuiz(){
	if(!checkAllAnswered())return;
	fetch('/api/personal-quiz/'+QUIZ_ID+'/archive',{method:'POST'}).then(function(r){return r.json()}).then(function(d){if(d.success)window.location='/dashboard';else if(d.error)alert(d.error);});
}
function publishQuiz(){
	if(!checkAllAnswered())return;
	fetch('/api/personal-quiz/'+QUIZ_ID,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'active'})}).then(function(r){return r.json()}).then(function(d){if(d.success)location.reload();else if(d.error)alert(d.error);});
}
function saveQuiz(){
	var btn=document.getElementById('save-btn');
	btn.textContent='💾 Sauvegarde...';btn.disabled=true;
	setTimeout(function(){
		btn.textContent='✅ Sauvegardé !';btn.style.borderColor='#22c55e';btn.style.color='#22c55e';
		setTimeout(function(){btn.textContent='💾 Sauvegarder';btn.disabled=false;btn.style.borderColor='';btn.style.color='';},2000);
	},300);
}
</script>
<div class="vtk-page" style="align-items:flex-start;padding-top:40px">
<div class="vtk-card" style="max-width:600px;text-align:left">
<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px">
<h1 style="margin:0;font-size:1.4rem">📝 ` + title + `</h1>
` + func() string {
	if status == "archived" { return `<span style="background:rgba(34,197,94,.15);color:#22c55e;padding:4px 12px;border-radius:20px;font-size:.8rem">Publié</span>` }
	return `<span style="background:rgba(255,255,255,.08);color:rgba(255,255,255,.5);padding:4px 12px;border-radius:20px;font-size:.8rem">Brouillon</span>`
}() + `
</div>
<div style="background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:16px;padding:16px;margin-bottom:20px">
<label style="font-size:.85rem;color:rgba(255,255,255,.5);margin-bottom:8px;display:block">🔗 Lien de partage</label>
<div style="display:flex;gap:8px;align-items:center;background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:12px;padding:4px 4px 4px 14px">
<span style="color:rgba(255,255,255,.35);font-size:.85rem;white-space:nowrap;font-family:monospace">/quiz/personal/</span>
<input type="text" id="share-token" value="` + token + `" oninput="onTokenChange()" style="flex:1;padding:8px 4px;background:transparent;border:none;color:#fff;font-size:.9rem;outline:none;font-family:monospace">
<button class="vtk-btn-icon" onclick="copyShareLink()">📋</button>
</div>
<button class="vtk-btn-ghost" id="save-link-btn" style="display:none;margin-top:8px" onclick="saveShareLink()">💾 Enregistrer le token</button>
</div>
<div id="questions-list">
` + questionsHTML + `
</div>
<button class="vtk-btn-ghost" style="margin-top:12px;padding:12px;border:1px dashed rgba(255,255,255,.15);border-radius:12px;width:100%;font-size:.9rem;text-align:center" onclick="addQuestion()">+ Ajouter une question</button>
<div style="display:flex;gap:8px;margin-top:20px">
<button class="vtk-btn-ghost" id="save-btn" onclick="saveQuiz()" style="flex:1;padding:12px;border:1px solid rgba(255,255,255,.15);border-radius:12px;font-size:.9rem;text-align:center">💾 Sauvegarder</button>
` + func() string {
    if status == "archived" { return `<button class="vtk-btn-form" onclick="publishQuiz()" style="flex:1">Publier le quiz</button>` }
    return `<button class="vtk-btn-form" onclick="archiveQuiz()" style="flex:1">Archiver le quiz</button>`
}() + `
</div>
</div>
</div>
<style>
@keyframes vtkPulse{0%{transform:scale(1)}50%{transform:scale(1.03)}100%{transform:scale(1)}}
.vtk-ques-card{background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:16px;padding:16px;margin-bottom:10px}
.vtk-ques-header{display:flex;align-items:center;gap:8px;margin-bottom:10px}
.vtk-ques-num{font-weight:700;color:#667eea;font-size:.85rem}
.vtk-ques-type{font-size:.75rem;color:rgba(255,255,255,.4);background:rgba(255,255,255,.06);padding:2px 10px;border-radius:10px}
.vtk-ques-del{margin-left:auto;background:none;border:none;color:#ef4444;cursor:pointer;font-size:1rem;padding:4px 8px;border-radius:8px}
.vtk-ques-del:hover{background:rgba(239,68,68,.1)}
.vtk-opt-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin-top:8px}
.vtk-opt-card{display:flex;flex-direction:column;align-items:center;justify-content:center;padding:12px 6px;border-radius:14px;border:1px solid rgba(255,255,255,.08);background:rgba(255,255,255,.04);color:#fff;cursor:pointer;transition:all .2s;font-family:inherit;aspect-ratio:1}
.vtk-opt-card:hover{border-color:rgba(102,126,234,.4);background:rgba(102,126,234,.08);transform:translateY(-2px)}
.vtk-opt-card:active{transform:scale(.95)}
.vtk-opt-card.selected{border-color:#22c55e;background:rgba(34,197,94,.15);animation:vtkPulse .3s ease}
.vtk-opt-emoji{font-size:1.8rem;margin-bottom:4px}
.vtk-opt-label{font-size:.7rem;font-weight:600;text-align:center;line-height:1.2}
.vtk-btn-icon{background:rgba(255,255,255,.08);border:none;color:#fff;padding:8px 12px;border-radius:8px;cursor:pointer;font-size:1rem;transition:background .2s}
.vtk-btn-icon:hover{background:rgba(255,255,255,.15)}
.vtk-btn-ghost{background:none;border:none;color:rgba(255,255,255,.5);font-size:.82rem;cursor:pointer;padding:6px 0;transition:color .2s;font-family:inherit}
.vtk-btn-ghost:hover{color:rgba(255,255,255,.8)}
.vtk-btn-form{background:linear-gradient(135deg,#667eea,#764ba2);border:none;color:#fff;font-weight:600;cursor:pointer;transition:opacity .2s;font-family:inherit;padding:12px;border-radius:12px;font-size:.9rem;text-align:center}
.vtk-btn-form:hover{opacity:.9}
</style>`

	return renderPage(c, "Mon quiz personnel", content)
}

// ============================================================
// PERSONAL QUIZ — PLAY (public link)
// ============================================================

func (h *Handler) PersonalQuizPlay(c *fiber.Ctx) error {
	token := c.Params("token")

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("token=eq.%s&status=eq.active&select=*", token), false)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)
	if len(quizzes) == 0 {
		return c.Status(404).SendString("Quiz introuvable ou archivé")
	}
	quiz := quizzes[0]
	quizID := database.DBValue(quiz["id"])
	creatorID := database.DBValue(quiz["creator_id"])
	title, _ := quiz["title"].(string)

	creatorBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=nickname,username,avatar_url", creatorID), false)
	var creators []map[string]interface{}
	json.Unmarshal(creatorBody, &creators)
	creatorName := "Anonyme"
	creatorAvatar := ""
	if len(creators) > 0 {
		if nn, ok := creators[0]["nickname"].(string); ok && nn != "" {
			creatorName = nn
		} else if un, ok := creators[0]["username"].(string); ok {
			creatorName = un
		}
		if av, ok := creators[0]["avatar_url"].(string); ok {
			creatorAvatar = av
		}
	}
	avatarHTML := `<div class="pq-creator-avatar" id="creator-avatar">` + creatorName[:1] + `</div>`
	if creatorAvatar != "" {
		avatarHTML = `<img class="pq-creator-avatar" src="` + htmlAttr(creatorAvatar) + `" alt="">`
	}

	user := h.getUserFromSession(c)

	if user == nil {
		sess, _ := h.store.Get(c)
		sess.Set("redirect_after_login", "/quiz/personal/"+token)
		sess.Save()
		guestAvatar := `<div class="vtk-avatar" style="margin:0 auto">` + creatorName[:1] + `</div>`
		if creatorAvatar != "" {
			guestAvatar = `<img class="vtk-avatar-img" src="` + htmlAttr(creatorAvatar) + `" alt="" style="margin:0 auto">`
		}
		return renderPage(c, title, `
<div class="vtk-page">
<div class="vtk-card">
<div class="vtk-hero">
` + guestAvatar + `
<h1>` + title + `</h1>
<p class="vtk-hero-sub">Quiz de ` + creatorName + `</p>
</div>
<div style="background:rgba(255,255,255,.04);border-radius:16px;padding:20px;margin:16px 0">
<p style="font-size:1.05rem;margin:0 0 6px">🔒 Connecte-toi pour jouer</p>
<p style="color:rgba(255,255,255,.4);font-size:.85rem;margin:0">Un clic suffit pour rejoindre le quiz !</p>
</div>
<a href="/login" class="vtk-btn-play" style="display:block;text-decoration:none">Se connecter →</a>
</div>
</div>`)
	}

	partBody, _ := h.db.Select("personal_quiz_participations",
		fmt.Sprintf("quiz_id=eq.%s&participant_id=eq.%s&select=score,correct_count,total_count", quizID, user.ID), true)
	var parts []map[string]interface{}
	json.Unmarshal(partBody, &parts)
	if len(parts) > 0 {
		score := database.DBInt(parts[0]["score"])
		correct := database.DBInt(parts[0]["correct_count"])
		total := database.DBInt(parts[0]["total_count"])

		rankBody, _ := h.db.RPC("rpc_get_personal_quiz_rank", map[string]interface{}{
			"p_quiz_id": quizID,
			"p_score":   score,
		})
		rank := 0
		if rankBody != nil {
			var rankResult []int
			json.Unmarshal(rankBody, &rankResult)
			if len(rankResult) > 0 {
				rank = rankResult[0]
			}
		}

		countBody, _ := h.db.Select("personal_quiz_participations",
			fmt.Sprintf("quiz_id=eq.%s&select=id", quizID), true)
		var allParts []map[string]interface{}
		json.Unmarshal(countBody, &allParts)
		totalPlayers := len(allParts)

		return renderPage(c, title, `
<div class="vtk-page">
<div class="vtk-card">
<div class="vtk-hero">
<h1>🎉 ` + title + `</h1>
<p class="vtk-hero-sub">Tu as déjà participé !</p>
</div>
<div class="vtk-score-card">
<div class="vtk-score-num">` + fmt.Sprintf("%d", score) + `</div>
<div class="vtk-score-label">points</div>
<div class="vtk-score-sub">` + fmt.Sprintf("%d/%d bonnes réponses", correct, total) + `</div>
</div>
<div class="vtk-rank-badge">#` + fmt.Sprintf("%d", rank) + ` / ` + fmt.Sprintf("%d", totalPlayers) + `</div>
<div style="display:flex;gap:10px;margin-top:20px">
<a href="/quiz/personal/` + token + `/results" class="vtk-btn-play" style="flex:1;display:block;text-decoration:none;padding:12px 0;font-size:.95rem">🏆 Classement</a>
</div>
</div>
</div>`)
	}

	questionsBody, _ := h.db.Select("personal_quiz_questions",
		fmt.Sprintf("quiz_id=eq.%s&select=*&order=sort_order.asc", quizID), false)
	var questions []map[string]interface{}
	json.Unmarshal(questionsBody, &questions)

	if len(questions) == 0 {
		return renderPage(c, title, `
<div class="vtk-page">
<div class="vtk-card">
<h1 style="margin:0 0 8px">🧐 ` + title + `</h1>
<p style="color:rgba(255,255,255,.5)">Ce quiz n'a pas encore de questions.</p>
<a href="/dashboard" class="vtk-btn-play" style="display:block;text-decoration:none;margin-top:20px">Retour →</a>
</div>
</div>`)
	}

	questionsJSON, _ := json.Marshal(questions)
	questionsStr := strings.ReplaceAll(string(questionsJSON), "</", "<\\/")

	content := `
<div class="vtk-page">
<div class="vtk-card" id="intro-card">
<div class="vtk-hero">
` + avatarHTML + `
<h1>` + title + `</h1>
<p class="vtk-hero-sub">Quiz de ` + creatorName + `</p>
</div>
<div class="vtk-intro-body">
<p style="font-size:3rem;margin:16px 0">🎯</p>
<p style="font-size:1.1rem"><strong>` + fmt.Sprintf("%d", len(questions)) + ` questions</strong> sur ` + creatorName + `</p>
<p style="color:rgba(255,255,255,.5);font-size:.9rem;margin-top:8px">⏱ ` + fmt.Sprintf("%d", len(questions)*15) + `s chrono</p>
<div style="display:flex;gap:4px;justify-content:center;margin:16px 0"><div style="width:8px;height:8px;border-radius:50%;background:rgba(255,255,255,.15);display:inline-block"></div><div style="width:8px;height:8px;border-radius:50%;background:rgba(255,255,255,.15);display:inline-block"></div><div style="width:8px;height:8px;border-radius:50%;background:rgba(255,255,255,.15);display:inline-block"></div><div style="width:8px;height:8px;border-radius:50%;background:rgba(255,255,255,.15);display:inline-block"></div><div style="width:8px;height:8px;border-radius:50%;background:rgba(255,255,255,.15);display:inline-block"></div></div>
<p style="color:rgba(255,255,255,.5);font-size:.85rem;line-height:1.5">Teste tes connaissances et découvre qui connaît le mieux ` + creatorName + ` !</p>
</div>
<button class="vtk-btn-play" onclick="startQuiz()">🏁 Commencer</button>
</div>

<div id="quiz-area" style="display:none">
<div id="vtk-dots" style="display:flex;gap:6px;justify-content:center;margin-bottom:16px"></div>
<div class="vtk-progress" id="vtk-progress"></div>
<div id="question-container" class="vtk-q-wrap"></div>
</div>
</div>
<style>
@keyframes vtkFadeIn{from{opacity:0;transform:translateY(16px)}to{opacity:1;transform:translateY(0)}}
@keyframes vtkPulse{0%{transform:scale(1)}50%{transform:scale(1.03)}100%{transform:scale(1)}}
@keyframes vtkPop{0%{transform:scale(.8);opacity:0}100%{transform:scale(1);opacity:1}}
@keyframes vtkSlideIn{from{opacity:0;transform:translateX(40px)}to{opacity:1;transform:translateX(0)}}
.vtk-page{min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px;background:linear-gradient(135deg,#0f0c29,#302b63,#24243e);box-sizing:border-box}
.vtk-card{background:rgba(255,255,255,.06);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);border:1px solid rgba(255,255,255,.1);border-radius:24px;padding:36px 28px;width:100%;max-width:460px;text-align:center;box-shadow:0 25px 50px -12px rgba(0,0,0,.5)}
.vtk-hero{margin-bottom:20px}
.vtk-hero h1{margin:12px 0 4px;font-size:1.6rem;font-weight:700}
.vtk-hero-sub{color:rgba(255,255,255,.4);margin:0;font-size:.95rem}
.vtk-intro-body{margin-top:16px}
.vtk-intro-body p{margin:0;font-size:1rem}
.vtk-avatar{width:72px;height:72px;border-radius:50%;background:linear-gradient(135deg,#667eea,#764ba2);display:flex;align-items:center;justify-content:center;font-size:1.8rem;font-weight:700;color:#fff;margin:0 auto;box-shadow:0 8px 24px rgba(102,126,234,.3);transition:transform .3s}
.vtk-avatar:hover{transform:scale(1.05)}
.vtk-avatar-img{width:72px;height:72px;border-radius:50%;object-fit:cover;display:block;box-shadow:0 8px 24px rgba(102,126,234,.3)}
.vtk-btn-play{background:linear-gradient(135deg,#667eea,#764ba2);border:none;color:#fff;padding:16px 48px;border-radius:16px;font-size:1.1rem;font-weight:600;cursor:pointer;margin-top:24px;transition:transform .15s,opacity .2s;width:100%;font-family:inherit}
.vtk-btn-play:hover{opacity:.9;transform:translateY(-1px)}
.vtk-btn-play:active{transform:scale(.97)}
.vtk-progress{margin-bottom:16px}
.vtk-bar{border-radius:8px;height:6px;overflow:hidden;background:rgba(255,255,255,.06)}
.vtk-bar-fill{height:100%;background:linear-gradient(90deg,#667eea,#764ba2);border-radius:8px;transition:width .5s cubic-bezier(.4,0,.2,1)}
.vtk-q-wrap{animation:vtkSlideIn .35s ease-out}
.vtk-q-card{background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:20px;padding:24px 20px}
.vtk-q-num{color:rgba(255,255,255,.3);font-size:.75rem;text-transform:uppercase;letter-spacing:.5px;margin-bottom:8px}
.vtk-q-text{font-size:1.2rem;font-weight:600;margin:0 0 20px;line-height:1.4}
.vtk-opt{display:block;width:100%;padding:14px 18px;margin-bottom:8px;border-radius:14px;border:1px solid rgba(255,255,255,.08);background:rgba(255,255,255,.04);color:#fff;font-size:.95rem;cursor:pointer;text-align:left;transition:all .2s;font-family:inherit;position:relative;overflow:hidden}
.vtk-opt:hover{border-color:rgba(102,126,234,.4);background:rgba(102,126,234,.08);transform:translateX(4px)}
.vtk-opt:active{transform:scale(.98)}
.vtk-opt.selected{border-color:#667eea;background:rgba(102,126,234,.15);font-weight:600;animation:vtkPulse .3s ease}
.vtk-opt input{display:none}
.vtk-grid2{display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:8px}
.vtk-opt-sq{aspect-ratio:1;border-radius:16px;border:1px solid rgba(255,255,255,.08);background:rgba(255,255,255,.04);color:#fff;font-size:.95rem;cursor:pointer;text-align:center;transition:all .2s;font-family:inherit;display:flex;align-items:center;justify-content:center;padding:8px;word-break:break-word;line-height:1.3}
.vtk-opt-sq:hover{border-color:rgba(102,126,234,.4);background:rgba(102,126,234,.08);transform:translateY(-2px)}
.vtk-opt-sq:active{transform:scale(.95)}
.vtk-opt-sq.selected{border-color:#667eea;background:rgba(102,126,234,.15);font-weight:600;animation:vtkPulse .3s ease}
.vtk-opt-sq input{display:none}
.vtk-input{width:100%;padding:14px 18px;border-radius:14px;border:1px solid rgba(255,255,255,.1);background:rgba(255,255,255,.04);color:#fff;font-size:1rem;outline:none;box-sizing:border-box;font-family:inherit;transition:border-color .2s}
.vtk-input:focus{border-color:rgba(102,126,234,.6)}
.vtk-btn-next{background:linear-gradient(135deg,#667eea,#764ba2);border:none;color:#fff;padding:14px 0;border-radius:14px;font-size:1rem;font-weight:600;cursor:pointer;margin-top:16px;transition:opacity .2s,transform .15s;width:100%;font-family:inherit}
.vtk-btn-next:hover{opacity:.9;transform:translateY(-1px)}
.vtk-btn-next:active{transform:scale(.97)}
.vtk-tbar{height:4px;background:rgba(255,255,255,.08);border-radius:4px;overflow:hidden;margin-bottom:2px}
.vtk-tfill{height:100%;background:#667eea;border-radius:4px;transition:width 1s linear}
.vtk-dot{width:10px;height:10px;border-radius:50%;background:rgba(255,255,255,.1);transition:all .3s;cursor:default;display:inline-block}
.vtk-dot.done{background:rgba(102,126,234,.5)}
.vtk-dot.curr{background:#667eea;box-shadow:0 0 8px rgba(102,126,234,.5);transform:scale(1.2)}
</style>
<script id="pq-data" type="application/json">` + questionsStr + `</script>
<script>
var qs=JSON.parse(document.getElementById('pq-data').textContent);
var current=0;var answers=[];
var timer;var TIME_PER_Q=15;var timeLeft;

function startQuiz(){
document.getElementById('intro-card').style.display='none';
document.getElementById('quiz-area').style.display='block';
renderDots();
renderQuestion();
}

function renderDots(){
var d='';
for(var i=0;i<qs.length;i++){
var cls='vtk-dot';
if(i<current)cls+=' done';
else if(i==current)cls+=' curr';
d+='<div class="'+cls+'"></div>';
}
document.getElementById('vtk-dots').innerHTML=d;
}

function renderQuestion(){
var q=qs[current];
timeLeft=TIME_PER_Q;
if(timer)clearInterval(timer);
var timerColor=timeLeft>3?'#667eea':'#ef4444';
var tbar='<div class="vtk-timer"><div class="vtk-tbar"><div class="vtk-tfill" id="tfill" style="width:100%;background:'+timerColor+'"></div></div><span id="tsec" style="font-size:.75rem;color:rgba(255,255,255,.3);display:block;text-align:right;margin-bottom:10px">'+timeLeft+'s</span></div>';
var h=tbar+'<div class="vtk-q-card"><p class="vtk-q-num">Question '+(current+1)+'/'+qs.length+'</p><p class="vtk-q-text">'+q.question_text+'</p>';
if(q.type==='vrai_faux'){
h+='<button class="vtk-opt" onclick="selectVF(this,\'Vrai\')"><input type="radio" name="ans"> ✅ Vrai</button><button class="vtk-opt" onclick="selectVF(this,\'Faux\')"><input type="radio" name="ans"> ❌ Faux</button>';
}else if(q.type==='qcm'&&q.options){
var opts=q.options;if(typeof opts==='string')opts=JSON.parse(opts);
h+='<div class="vtk-grid2">';
opts.forEach(function(o,i){h+='<button class="vtk-opt-sq" onclick="selectQCM(this,\''+o.replace(/'/g,"\\'")+'\')"><input type="radio" name="ans"><span>'+o+'</span></button>';});
h+='</div>';
}else{
h+='<input class="vtk-input" id="pq-text-answer" placeholder="Écris ta réponse..." onkeydown="if(event.key==\'Enter\')nextQuestion()">';
}
h+='<button class="vtk-btn-next" onclick="nextQuestion()">'+(current<qs.length-1?'Suivant →':'Terminer ✓')+'</button></div>';
var wrap=document.getElementById('question-container');
wrap.style.animation='none';void wrap.offsetWidth;
wrap.style.animation='vtkSlideIn .35s ease-out';
wrap.innerHTML=h;
document.getElementById('vtk-progress').innerHTML='<div class="vtk-bar"><div class="vtk-bar-fill" style="width:'+((current+1)*100/qs.length)+'%"></div></div><p style="color:rgba(255,255,255,.3);font-size:.8rem;margin-top:6px">'+(current+1)+'/'+qs.length+'</p>';
timer=setInterval(function(){
timeLeft--;
var f=document.getElementById('tfill');if(f)f.style.width=((timeLeft/TIME_PER_Q)*100)+'%';
var s=document.getElementById('tsec');if(s)s.textContent=timeLeft+'s';
if(timeLeft<=3&&f)f.style.background='#ef4444';
if(timeLeft<=0){clearInterval(timer);if(!answers[current])answers[current]='';current++;if(current>=qs.length)submitQuiz();else{renderDots();renderQuestion();}}
},1000);
}

function selectVF(el,val){
document.querySelectorAll('.vtk-opt').forEach(function(e){e.classList.remove('selected')});el.classList.add('selected');answers[current]=val;
if(current<qs.length-1){setTimeout(function(){nextQuestion();},300);}
}
function selectQCM(el,val){
document.querySelectorAll('.vtk-opt').forEach(function(e){e.classList.remove('selected')});el.classList.add('selected');answers[current]=val;
if(current<qs.length-1){setTimeout(function(){nextQuestion();},300);}
}

function nextQuestion(){
if(!answers[current]){var ta=document.getElementById('pq-text-answer');if(ta&&ta.value){answers[current]=ta.value;}else{return;}}
clearInterval(timer);
current++;
if(current>=qs.length){submitQuiz();return;}
renderDots();
renderQuestion();
}

function submitQuiz(){
var score=0;var correct=0;
for(var i=0;i<qs.length;i++){
var a=answers[i]||'';var ca=qs[i].correct_answer;
if(a.toLowerCase().trim()===ca.toLowerCase().trim()){score+=10;correct++;}
}
fetch('/api/personal-quiz/` + token + `/submit',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({score:score,correct_count:correct,total_count:qs.length,answers:answers})}).then(function(r){return r.json()}).then(function(d){
if(d.success){window.location='/quiz/personal/` + token + `/results';}
});
}
</script>`

	return renderPage(c, title, content)
}

// ============================================================
// PERSONAL QUIZ — SUBMIT
// ============================================================

func (h *Handler) PersonalQuizSubmit(c *fiber.Ctx) error {
	token := c.Params("token")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("token=eq.%s&status=eq.active&select=id,creator_id", token), true)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)
	if len(quizzes) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Quiz introuvable"})
	}
	quizID := database.DBValue(quizzes[0]["id"])
	creatorID := database.DBValue(quizzes[0]["creator_id"])

	type Request struct {
		Score        int           `json:"score"`
		CorrectCount int           `json:"correct_count"`
		TotalCount   int           `json:"total_count"`
		Answers      []interface{} `json:"answers"`
	}
	var req Request
	c.BodyParser(&req)

	checkBody, _ := h.db.Select("personal_quiz_participations",
		fmt.Sprintf("quiz_id=eq.%s&participant_id=eq.%s&select=id", quizID, user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Déjà participé"})
	}

	partData, _ := json.Marshal(map[string]interface{}{
		"quiz_id":       quizID,
		"participant_id": user.ID,
		"score":         req.Score,
		"correct_count": req.CorrectCount,
		"total_count":   req.TotalCount,
		"answers":       req.Answers,
	})
	h.db.Insert("personal_quiz_participations", partData, true)

	if creatorID != user.ID {
		notifData, _ := json.Marshal(map[string]interface{}{
			"user_id":    creatorID,
			"type":       "personal_quiz_participation",
			"title":      "Nouvelle participation !",
			"message":    fmt.Sprintf("%s a participé à votre quiz !", getDisplayName(user)),
			"linked_id":  quizID,
			"is_read":    false,
		})
		h.db.Insert("notifications", notifData, true)
	}

	return c.JSON(fiber.Map{"success": true})
}

// ============================================================
// PERSONAL QUIZ — RESULTS (ranking + messages)
// ============================================================

func (h *Handler) PersonalQuizResults(c *fiber.Ctx) error {
	token := c.Params("token")
	user := h.getUserFromSession(c)

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("token=eq.%s&status=eq.active&select=id,creator_id,title", token), false)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)
	if len(quizzes) == 0 {
		return c.Status(404).SendString("Quiz introuvable")
	}
	quiz := quizzes[0]
	quizID := database.DBValue(quiz["id"])
	title, _ := quiz["title"].(string)
	creatorID := database.DBValue(quiz["creator_id"])

	creatorBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=nickname,username,avatar_url", creatorID), false)
	var creators []map[string]interface{}
	json.Unmarshal(creatorBody, &creators)
	creatorName := "Anonyme"
	if len(creators) > 0 {
		if nn, ok := creators[0]["nickname"].(string); ok && nn != "" {
			creatorName = nn
		} else if un, ok := creators[0]["username"].(string); ok {
			creatorName = un
		}
	}
	fullURL := "https://otakuquiz.africa/quiz/personal/" + token

	partsBody, _ := h.db.Select("personal_quiz_participations",
		fmt.Sprintf("quiz_id=eq.%s&select=*,participant:participant_id(nickname,username)", quizID),
		false)
	var parts []map[string]interface{}
	json.Unmarshal(partsBody, &parts)

	type RankEntry struct {
		Nickname    string
		Score       int
		Correct     int
		Total       int
		PartDate    string
		IsMe        bool
	}
	rankings := []RankEntry{}
	for _, p := range parts {
		pid := database.DBValue(p["participant_id"])
		pName := "Anonyme"
		if profArr, ok := p["participant"].([]interface{}); ok && len(profArr) > 0 {
			if prof, ok := profArr[0].(map[string]interface{}); ok {
				if nn, ok := prof["nickname"].(string); ok && nn != "" {
					pName = nn
				} else if un, ok := prof["username"].(string); ok {
					pName = un
				}
			}
		} else if prof, ok := p["participant"].(map[string]interface{}); ok {
			if nn, ok := prof["nickname"].(string); ok && nn != "" {
				pName = nn
			} else if un, ok := prof["username"].(string); ok {
				pName = un
			}
		}
		rankings = append(rankings, RankEntry{
			Nickname: pName,
			Score:    database.DBInt(p["score"]),
			Correct:  database.DBInt(p["correct_count"]),
			Total:    database.DBInt(p["total_count"]),
			PartDate: database.DBValue(p["participated_at"]),
			IsMe:     user != nil && pid == user.ID,
		})
	}

	rankHTML := ""
	for i, r := range rankings {
		nameStyle := "color:rgba(255,255,255,.7)"
		meClass := ""
		if r.IsMe {
			nameStyle = "color:#667eea;font-weight:700"
			meClass = " vtk-rank-me"
		}
		rankHTML += fmt.Sprintf(`
<div class="vtk-rank-row%s">
<span class="vtk-rank-num">%d</span>
<span style="%s;flex:1">%s</span>
<span style="color:rgba(255,255,255,.35);font-size:.8rem">%d/%d</span>
<span style="color:#667eea;font-weight:700;margin-left:10px">%d pts</span>
</div>`,
			meClass, i+1, nameStyle, r.Nickname, r.Correct, r.Total, r.Score)
	}

	questionsBody, _ := h.db.Select("personal_quiz_questions",
		fmt.Sprintf("quiz_id=eq.%s&select=question_text,type,correct_answer,options&order=sort_order.asc", quizID), false)
	var questions []map[string]interface{}
	json.Unmarshal(questionsBody, &questions)

	msgsBody, _ := h.db.Select("personal_quiz_messages",
		fmt.Sprintf("quiz_id=eq.%s&select=*&order=sent_at.desc", token), false)
	var msgs []map[string]interface{}
	json.Unmarshal(msgsBody, &msgs)

	isCreator := user != nil && user.ID == creatorID

	msgsHTML := ""
	for _, m := range msgs {
		mText := database.DBValue(m["message_text"])
		isAnon := database.DBBool(m["is_anonymous"])
		dispName := "Anonyme"
		if !isAnon {
			if dn := database.DBValue(m["display_nickname"]); dn != "" {
				dispName = dn
			}
		}
		if isAnon && !isCreator {
			continue
		}
		msgsHTML += fmt.Sprintf(`
<div class="pq-msg">
<div class="pq-msg-name">%s</div>
<div class="pq-msg-text">%s</div>
<div class="pq-msg-time">%s</div>
</div>`, dispName, mText, h.timeAgo(database.DBValue(m["sent_at"])))
	}
	if msgsHTML == "" {
		msgsHTML = `<p style="color:#94a3b8;text-align:center;padding:20px">Aucun message pour l'instant</p>`
	}

	content := `
<div class="vtk-page" style="align-items:flex-start;padding-top:30px">
<div class="vtk-card" style="max-width:520px">

<div style="text-align:center">
<h2 style="margin:0;font-size:1.8rem">Bravo, ` + creatorName + ` ! 🤩</h2>
<h3 style="margin:8px 0 0;color:rgba(255,255,255,.5);font-weight:400">Ton quiz est prêt !</h3>
</div>

<div style="margin:24px 0 16px;text-align:center">
<h4 style="margin:0 0 8px;font-size:.9rem;color:rgba(255,255,255,.6)">Partage ton quiz avec tes amis :</h4>
<div style="font-size:2rem;margin-bottom:12px">👇👇👇👇</div>

<a href="whatsapp://send?text=` + url.QueryEscape("*"+creatorName+"* a lancé un défi ! 😍\n\nClique MAINTENANT 👇👇👇\n"+fullURL) + `" style="display:flex;align-items:center;justify-content:center;gap:10px;width:100%;padding:15px;margin-bottom:10px;border-radius:12px;background:#25D366;color:#fff;font-size:1.1rem;font-weight:600;text-decoration:none;font-family:inherit;animation:pulse-share 2s infinite">
<span style="font-size:1.4rem">💬</span> WhatsApp</a>

<div onclick="shareInstagram()" style="display:flex;align-items:center;justify-content:center;gap:10px;width:100%;padding:15px;margin-bottom:10px;border-radius:12px;background:linear-gradient(45deg,#f09433,#e6683c,#dc2743,#cc2366,#bc1888);color:#fff;font-size:1.1rem;font-weight:600;cursor:pointer;animation:pulse-share 2s infinite">
<span style="font-size:1.4rem">📸</span> Partager en Story</div>

<a id="snapShareBtn" href="#" data-share-url="` + fullURL + `" data-share-text="` + creatorName + ` a lancé un défi !" style="display:flex;align-items:center;justify-content:center;gap:10px;width:100%;padding:15px;margin-bottom:10px;border-radius:12px;background:#FFFC00;color:#000;font-size:1.1rem;font-weight:600;text-decoration:none;font-family:inherit;animation:pulse-share 2s infinite">
<span style="font-size:1.4rem">👻</span> Snapchat</a>

<a href="fb-messenger://share?link=` + url.QueryEscape(fullURL) + `" style="display:flex;align-items:center;justify-content:center;gap:10px;width:100%;padding:15px;margin-bottom:10px;border-radius:12px;background:#8E44AD;color:#fff;font-size:1.1rem;font-weight:600;text-decoration:none;font-family:inherit;animation:pulse-share 2s infinite">
<span style="font-size:1.4rem">💭</span> Messenger</a>

<a id="smsShareBtn" href="#" data-share-url="` + fullURL + `" data-share-text="` + creatorName + ` a lancé un défi ! 😍" style="display:flex;align-items:center;justify-content:center;gap:10px;width:100%;padding:15px;margin-bottom:10px;border-radius:12px;background:#34c759;color:#fff;font-size:1.1rem;font-weight:600;text-decoration:none;font-family:inherit;animation:pulse-share 2s infinite">
<span style="font-size:1.4rem">✉️</span> SMS</a>
</div>

<h4 style="margin:16px 0 8px;font-size:.9rem;color:rgba(255,255,255,.6);text-align:center">Autres réseaux :</h4>

<div style="display:flex;gap:8px;margin-bottom:10px">
<textarea readonly id="share-link" style="flex:1;padding:10px 14px;border-radius:10px;border:1px solid rgba(255,255,255,.1);background:rgba(255,255,255,.04);color:#fff;font-size:.85rem;outline:none;font-family:monospace;resize:none" rows="1">` + fullURL + `</textarea>
<button onclick="navigator.clipboard.writeText(document.getElementById('share-link').value);this.textContent='✅ Copié !';var b=this;setTimeout(function(){b.textContent='📋 Copier'},1500)" style="padding:10px 16px;border-radius:10px;border:1px solid rgba(255,255,255,.15);background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;font-size:.9rem;font-weight:600;cursor:pointer;font-family:inherit;white-space:nowrap">📋 Copier</button>
</div>

<div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:16px">
<a href="https://www.facebook.com/sharer/sharer.php?u=` + url.QueryEscape(fullURL) + `" target="_blank" style="display:flex;align-items:center;justify-content:center;gap:6px;padding:12px;border-radius:10px;background:#1877F2;color:#fff;font-size:.9rem;font-weight:600;text-decoration:none;font-family:inherit">📘 Facebook</a>
<a href="https://t.me/share/url?url=` + url.QueryEscape(fullURL) + `&text=` + url.QueryEscape("*"+creatorName+"* a lancé un défi ! 😍") + `" target="_blank" style="display:flex;align-items:center;justify-content:center;gap:6px;padding:12px;border-radius:10px;background:#0088cc;color:#fff;font-size:.9rem;font-weight:600;text-decoration:none;font-family:inherit">✈️ Telegram</a>
<div onclick="shareTikTok()" style="display:flex;align-items:center;justify-content:center;gap:6px;padding:12px;border-radius:10px;background:#f1db2f;color:#000;font-size:.9rem;font-weight:600;cursor:pointer;font-family:inherit">🎵 TikTok</div>
<a href="https://twitter.com/intent/tweet?text=` + url.QueryEscape("*"+creatorName+"* a lancé un défi ! 😍 "+fullURL) + `" target="_blank" style="display:flex;align-items:center;justify-content:center;gap:6px;padding:12px;border-radius:10px;background:#1DA1F2;color:#fff;font-size:.9rem;font-weight:600;text-decoration:none;font-family:inherit">🐦 Twitter</a>
</div>

<h4 style="margin:16px 0 8px;font-size:.9rem;color:rgba(255,255,255,.6);text-align:center">Après le partage, vérifie qui a répondu :</h4>
<a href="/quiz/personal/` + token + `/results" style="display:block;width:100%;padding:14px;border-radius:14px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;font-size:1rem;font-weight:600;text-decoration:none;text-align:center;font-family:inherit">📋 Voir les réponses</a>

<div style="margin:24px 0 0;text-align:center">
<p style="color:rgba(255,255,255,.5);font-size:.85rem;margin:0 0 12px">Crée ton propre quiz et partage-le !</p>
<a href="/personal-quiz/create" style="display:inline-block;padding:14px 32px;border-radius:14px;background:linear-gradient(135deg,#f093fb,#f5576c);color:#fff;font-size:1rem;font-weight:600;text-decoration:none;font-family:inherit;animation:pulse-share 2s infinite">🎯 Créer mon quiz</a>
</div>

</div>
</div>

<!-- INSTAGRAM POPUP -->
<div id="insta-popup" onclick="if(event.target===this)closePopup()" style="display:none;position:fixed;inset:0;z-index:999;background:rgba(0,0,0,.7);backdrop-filter:blur(4px);align-items:center;justify-content:center">
<div style="background:rgba(255,255,255,.1);backdrop-filter:blur(20px);border:1px solid rgba(255,255,255,.15);border-radius:20px;padding:24px;max-width:380px;width:90%;position:relative">
<button onclick="closePopup()" style="position:absolute;top:12px;right:16px;background:none;border:none;color:#fff;font-size:1.5rem;cursor:pointer">✕</button>
<h3 style="margin:0 0 12px;text-align:center">📸 Partager en Story</h3>
<p style="color:rgba(255,255,255,.6);font-size:.9rem;margin:0 0 12px">1️⃣ Enregistre cette image</p>
<div style="background:rgba(255,255,255,.06);border-radius:12px;padding:16px;text-align:center;margin-bottom:12px">
<p style="font-size:1.2rem;margin:0">🎌 Otaku Quiz 🎌</p>
<p style="color:rgba(255,255,255,.5);margin:4px 0 0;font-size:.8rem">` + creatorName + ` t'a lancé un défi !</p>
</div>
<p style="color:rgba(255,255,255,.6);font-size:.9rem;margin:0 0 8px">2️⃣ Crée une Story et ajoute le lien en sticker</p>
<div style="display:flex;gap:8px;align-items:center;background:rgba(255,255,255,.06);border-radius:10px;padding:8px 12px">
<input readonly value="` + fullURL + `" style="flex:1;background:transparent;border:none;color:#fff;font-size:.8rem;outline:none;font-family:monospace">
<button onclick="navigator.clipboard.writeText(this.previousElementSibling.value);this.textContent='✅'" style="background:rgba(255,255,255,.1);border:none;color:#fff;padding:6px 10px;border-radius:8px;cursor:pointer;font-size:.85rem">📋</button>
</div>
</div>
</div>

<style>
@keyframes pulse-share{0%,100%{transform:scale(1)}50%{transform:scale(1.01)}}
</style>
<script>
function shareInstagram(){document.getElementById('insta-popup').style.display='flex';}
function closePopup(){document.getElementById('insta-popup').style.display='none';}
function shareTikTok(){var u='` + fullURL + `';if(navigator.share){navigator.share({title:'Otaku Quiz',text:'`+creatorName+` t\\'a lancé un défi !',url:u}).catch(function(){});}else{navigator.clipboard.writeText(u);alert('Lien copié ! Colle-le dans ton bio TikTok.');}}
var s=document.getElementById('snapShareBtn');if(s){s.addEventListener('click',function(e){e.preventDefault();var u=this.getAttribute('data-share-url');var t=this.getAttribute('data-share-text');if(navigator.share&&u){navigator.share({title:document.title,text:t,url:u}).catch(function(){});}else{window.open('https://www.snapchat.com/scan?attachment_url='+encodeURIComponent(u),'_blank');}});}
var m=document.getElementById('smsShareBtn');if(m){m.addEventListener('click',function(e){e.preventDefault();var u=this.getAttribute('data-share-url');var t=this.getAttribute('data-share-text');var body=t+' '+u;if(navigator.share&&navigator.canShare&&navigator.canShare({text:body})){navigator.share({text:body}).catch(function(){});}else{window.location.href='sms:?&body='+encodeURIComponent(body);}});}
</script>`

	return renderPage(c, title+" — Résultats", content)
}

func (h *Handler) PersonalQuizSendMessage(c *fiber.Ctx) error {
	token := c.Params("token")
	user := h.getUserFromSession(c)

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("token=eq.%s&status=eq.active&select=id", token), false)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)
	if len(quizzes) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Quiz introuvable"})
	}
	quizID := database.DBValue(quizzes[0]["id"])

	type Request struct {
		Message      string `json:"message"`
		IsAnonymous  bool   `json:"is_anonymous"`
	}
	var req Request
	c.BodyParser(&req)

	if req.Message == "" || len(req.Message) > 500 {
		return c.Status(400).JSON(fiber.Map{"error": "Message invalide"})
	}

	var senderID *string
	var displayNickname *string

	if !req.IsAnonymous && user != nil {
		senderID = &user.ID
		if user.Nickname != nil && *user.Nickname != "" {
			displayNickname = user.Nickname
		} else {
			displayNickname = &user.Username
		}
	}

	var partID *string
	if user != nil {
		partBody, _ := h.db.Select("personal_quiz_participations",
			fmt.Sprintf("quiz_id=eq.%s&participant_id=eq.%s&select=id", quizID, user.ID), true)
		var parts []map[string]interface{}
		json.Unmarshal(partBody, &parts)
		if len(parts) > 0 {
			v := database.DBValue(parts[0]["id"])
			partID = &v
		}
	}

	quizCreatorBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&select=creator_id", quizID), true)
	var qcs []map[string]interface{}
	json.Unmarshal(quizCreatorBody, &qcs)
	creatorID := ""
	if len(qcs) > 0 {
		creatorID = database.DBValue(qcs[0]["creator_id"])
	}

	msgData := map[string]interface{}{
		"quiz_id":      quizID,
		"message_text": req.Message,
		"is_anonymous": req.IsAnonymous,
	}
	if partID != nil {
		msgData["participation_id"] = *partID
	}
	if senderID != nil {
		msgData["sender_id"] = *senderID
	}
	if displayNickname != nil {
		msgData["display_nickname"] = *displayNickname
	}
	data, _ := json.Marshal(msgData)
	h.db.Insert("personal_quiz_messages", data, true)

	if creatorID != "" && senderID != nil && *senderID != creatorID {
		notifData, _ := json.Marshal(map[string]interface{}{
			"user_id":    creatorID,
			"type":       "personal_quiz_message",
			"title":      "Nouveau message !",
			"message":    fmt.Sprintf("Quelqu'un vous a laissé un message sur votre quiz !"),
			"linked_id":  quizID,
			"is_read":    false,
		})
		h.db.Insert("notifications", notifData, true)
	}

	return c.JSON(fiber.Map{"success": true})
}

// ============================================================
// API CRUD QUESTIONS
// ============================================================

func (h *Handler) APIPersonalQuizAddQuestion(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	quizID := c.Params("id")

	checkBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&creator_id=eq.%s&select=id", quizID, user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	countBody, _ := h.db.Select("personal_quiz_questions",
		fmt.Sprintf("quiz_id=eq.%s&select=id", quizID), true)
	var all []map[string]interface{}
	json.Unmarshal(countBody, &all)

	type Request struct {
		QuestionText string      `json:"question_text"`
		Type         string      `json:"type"`
		CorrectAnswer string    `json:"correct_answer"`
		Options      interface{} `json:"options"`
	}
	var req Request
	c.BodyParser(&req)

	if req.Type == "" {
		req.Type = "qcm"
	}

	data, _ := json.Marshal(map[string]interface{}{
		"quiz_id":        quizID,
		"question_text":  req.QuestionText,
		"type":           req.Type,
		"correct_answer": req.CorrectAnswer,
		"options":        req.Options,
		"sort_order":     len(all),
	})
	h.db.Insert("personal_quiz_questions", data, true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIPersonalQuizUpdateQuestion(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	quizID := c.Params("id")
	qID := c.Params("qid")

	checkBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&creator_id=eq.%s&select=id", quizID, user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	var updates map[string]interface{}
	c.BodyParser(&updates)

	data, _ := json.Marshal(updates)
	h.db.Update("personal_quiz_questions",
		fmt.Sprintf("id=eq.%s&quiz_id=eq.%s", qID, quizID), data, true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIPersonalQuizDeleteQuestion(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	quizID := c.Params("id")
	qID := c.Params("qid")

	checkBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&creator_id=eq.%s&select=id", quizID, user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	h.db.Delete("personal_quiz_questions",
		fmt.Sprintf("id=eq.%s&quiz_id=eq.%s", qID, quizID), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIPersonalQuizArchive(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	quizID := c.Params("id")

	checkBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&creator_id=eq.%s&select=id", quizID, user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	questionsBody, _ := h.db.Select("personal_quiz_questions",
		fmt.Sprintf("quiz_id=eq.%s&select=correct_answer", quizID), true)
	var questions []map[string]interface{}
	json.Unmarshal(questionsBody, &questions)
	for _, q := range questions {
		ca, _ := q["correct_answer"].(string)
		if ca == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Toutes les questions doivent avoir une bonne réponse"})
		}
	}

	h.db.Update("personal_quizzes",
		fmt.Sprintf("id=eq.%s", quizID),
		[]byte(`{"status":"archived"}`), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIPersonalQuizUpdate(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	quizID := c.Params("id")

	checkBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("id=eq.%s&creator_id=eq.%s&select=id,token", quizID, user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		Token  string `json:"token"`
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	updates := map[string]interface{}{}
	if req.Token != "" {
		tokenBody, _ := h.db.Select("personal_quizzes",
			fmt.Sprintf("token=eq.%s&id=neq.%s&select=id", req.Token, quizID), true)
		var tokenExisting []map[string]interface{}
		json.Unmarshal(tokenBody, &tokenExisting)
		if len(tokenExisting) > 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Ce lien est déjà utilisé"})
		}
		updates["token"] = req.Token
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if len(updates) == 0 {
		return c.JSON(fiber.Map{"success": true})
	}

	jsonData, _ := json.Marshal(updates)
	h.db.Update("personal_quizzes", fmt.Sprintf("id=eq.%s", quizID), jsonData, true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) PersonalQuizView(c *fiber.Ctx) error {
	token := c.Params("token")

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("token=eq.%s&status=eq.active&select=id", token), false)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)
	if len(quizzes) == 0 {
		return c.Status(404).SendString("Quiz introuvable")
	}

	quizID := database.DBValue(quizzes[0]["id"])

	user := h.getUserFromSession(c)
	if user != nil {
		checkBody, _ := h.db.Select("personal_quiz_participations",
			fmt.Sprintf("quiz_id=eq.%s&participant_id=eq.%s&select=id", quizID, user.ID), true)
		var existing []map[string]interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			return c.Redirect("/quiz/personal/" + token + "/results")
		}
	}

	return c.Redirect("/quiz/personal/" + token + "/play")
}

func (h *Handler) PersonalQuizDashboard(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	quizBody, _ := h.db.Select("personal_quizzes",
		fmt.Sprintf("creator_id=eq.%s&select=*&order=created_at.desc", user.ID), true)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)

	content := `<div class="pq-page"><h1>🎯 Mon Quiz Personnel</h1>`

	if len(quizzes) == 0 {
		content += `
<p style="color:#94a3b8;margin-bottom:16px">Tu n'as pas encore créé de quiz personnel.</p>
<form method="POST" action="/personal-quiz/create">
<button type="submit" class="btn-primary">Créer mon quiz personnel</button>
</form>`
	} else {
		for _, q := range quizzes {
			qID := database.DBValue(q["id"])
			qTitle, _ := q["title"].(string)
			qToken, _ := q["token"].(string)
			qStatus, _ := q["status"].(string)
			qCreatedAt := database.DBValue(q["created_at"])

			statusBadge := `<span style="color:#22c55e;font-weight:600">Actif</span>`
			if qStatus == "archived" {
				statusBadge = `<span style="color:#fbbf24;font-weight:600">Archivé</span>`
			}

			partBody, _ := h.db.Select("personal_quiz_participations",
				fmt.Sprintf("quiz_id=eq.%s&select=id", qID), true)
			var parts []map[string]interface{}
			json.Unmarshal(partBody, &parts)

			qCountBody, _ := h.db.Select("personal_quiz_questions",
				fmt.Sprintf("quiz_id=eq.%s&select=id", qID), true)
			var qRows []map[string]interface{}
			json.Unmarshal(qCountBody, &qRows)

			content += fmt.Sprintf(`
<div class="pq-dash-card">
<div style="flex:1">
<div style="font-weight:600;margin-bottom:4px">%s</div>
<div style="font-size:.8rem;color:#94a3b8">%d questions • %d participations • %s • Créé %s</div>
</div>
<div style="display:flex;gap:8px">
%s
<button class="btn-sm btn-outline" onclick="copyPQLink('%s')">🔗 Lien</button>
<a href="/quiz/personal/%s/results" class="btn-sm btn-ghost">🏆 Résultats</a>
%s
</div>
</div>`, qTitle, len(qRows), len(parts), statusBadge, h.timeAgo(qCreatedAt),
				func() string { if qStatus == "active" || qStatus == "archived" { return `<a href="/personal-quiz/` + qID + `/edit" class="btn-sm btn-outline">✏️ Modifier</a>` }; return "" }(),
				qToken, qToken,
				func() string {
    if qStatus == "active" { return `<button class="btn-sm btn-danger" onclick="archivePQ('` + qID + `')">🗑️ Archiver</button>` }
    if qStatus == "archived" { return `<button class="btn-sm btn-primary" onclick="publishPQ('` + qID + `')">Publier</button>` }
    return ""
}())
		}
	}

	content += `
</div>
<script>
function archivePQ(id){
if(!confirm('Archiver ce quiz ? Le lien de partage sera désactivé.'))return;
fetch('/api/personal-quiz/'+id+'/archive',{method:'POST'}).then(function(r){return r.json()}).then(function(d){if(d.success)location.reload();});
}
function publishPQ(id){
fetch('/api/personal-quiz/'+id,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'active'})}).then(function(r){return r.json()}).then(function(d){if(d.success)location.reload();});
}
function copyPQLink(token){
var full=window.location.origin+'/quiz/personal/'+token;
navigator.clipboard.writeText(full).then(function(){alert('Lien copié !');}).catch(function(){var tmp=document.createElement('input');tmp.value=full;document.body.appendChild(tmp);tmp.select();document.execCommand('copy');document.body.removeChild(tmp);alert('Lien copié !');});
}
</script>
<style>
.pq-page{max-width:700px;margin:0 auto}
.pq-dash-card{display:flex;align-items:center;justify-content:space-between;background:#16213e;border:1px solid #2d2d44;border-radius:10px;padding:16px;margin-bottom:8px;flex-wrap:wrap;gap:8px}
.btn-danger{background:none;border:1px solid #ef4444;color:#ef4444;border-radius:6px;padding:4px 10px;cursor:pointer;font-size:.8rem}
.btn-danger:hover{background:rgba(239,68,68,.1)}
</style>`

	return renderPage(c, "Mon Quiz Personnel", content)
}

func (h *Handler) PersonalQuizCreatePost(c *fiber.Ctx) error {
	// POST is handled in PersonalQuizCreate's POST branch
	return h.PersonalQuizCreate(c)
}
