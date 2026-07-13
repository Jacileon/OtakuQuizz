package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"otaku-quiz-africa/internal/database"
)

func (h *Handler) QuizResults(c *fiber.Ctx) error {
	id := c.Params("id")
	sessionID := c.Params("session")

	sessionBody, _ := h.db.Select("game_sessions",
		fmt.Sprintf("id=eq.%s&select=*", sessionID), true)
	var sessions []map[string]interface{}
	json.Unmarshal(sessionBody, &sessions)

	if len(sessions) == 0 {
		return c.Redirect("/quiz/" + id)
	}

	s := sessions[0]
	score := database.DBInt(s["score"])
	total := database.DBInt(s["total_questions"])
	acc := database.DBFloat(s["accuracy_rate"])

	return renderWithContent(c, "Résultats", fmt.Sprintf(`
<div class="results-page">
    <h1>Quiz Terminé ! 🎉</h1>
    <div class="results-card">
        <div class="results-score">
            <div class="rs-number">%d</div>
            <p class="text-muted">sur %d questions</p>
        </div>
        <div class="results-divider"></div>
        <div class="results-pct">
            <div class="rs-number">%.0f%%</div>
            <p class="text-muted">de bonnes réponses</p>
        </div>
    </div>
    <div class="results-actions">
        <a href="/quiz/%s/play" class="btn-primary btn-lg">🔄 Recommencer</a>
        <a href="/explore" class="btn-outline btn-lg">← Retour</a>
    </div>
</div>`, score, total, acc, id))
}

func (h *Handler) QuizCreate(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	if !user.IsAdmin && !user.CanCreateQuiz {
		return renderPage(c, "Non autorisé", `
<div style="text-align:center;padding:60px 20px;">
<h1 style="margin-bottom:16px;">🔒 Accès restreint</h1>
<p style="color:#94a3b8;margin-bottom:24px;">Seuls les utilisateurs d'un rang défini par un admin et les admins peuvent créer des quiz.</p>
<a href="/explore" class="btn-primary">Retour à l'exploration</a>
</div>`)
	}

	categories := []string{"Anime/Manga", "Culture Générale", "Autre"}
	catOptions := ""
	for _, cat := range categories {
		catOptions += fmt.Sprintf(`<option value="%s">%s</option>`, cat, cat)
	}

	return renderWithContent(c, "Créer un quiz", fmt.Sprintf(`
<div class="quiz-create-page">
    <h1 class="page-title">📝 Créer un quiz</h1>
    <div class="pe-card">
        <form method="POST" action="/quiz/create">
            <div class="pe-field">
                <label>Titre du quiz</label>
                <input type="text" name="title" placeholder="Ex: Quiz Naruto Shippuden" class="pe-input" required>
            </div>
            <div class="pe-field">
                <label>Description (optionnelle)</label>
                <textarea name="description" rows="3" placeholder="Décris ton quiz..." class="pe-input"></textarea>
            </div>
            <div class="pe-field">
                <label>Série/Sujet</label>
                <input type="text" name="series" placeholder="Ex: Naruto, One Piece, Dragon Ball..." class="pe-input" required>
            </div>
            <div class="pe-field">
                <label>Catégorie</label>
                <select name="category" class="pe-input" required>
                    <option value="">Sélectionner...</option>
                    %s
                </select>
            </div>
            <button type="submit" class="btn-primary btn-lg w-full">Créer le quiz</button>
        </form>
    </div>
</div>`, catOptions))
}

func (h *Handler) QuizCreatePost(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	if !user.IsAdmin && !user.CanCreateQuiz {
		return c.Redirect("/quiz/create")
	}

	title := c.FormValue("title")
	description := c.FormValue("description")
	quizSeries := c.FormValue("series")
	category := c.FormValue("category")
	subcategory := c.FormValue("subcategory")

	if title == "" || quizSeries == "" || category == "" {
		return c.Redirect("/quiz/create")
	}
	if len(title) < 5 {
		return c.Redirect("/quiz/create")
	}
	if subcategory == "" {
		subcategory = category
	}

	quizData, _ := json.Marshal(map[string]interface{}{
		"creator_id":  user.ID,
		"title":       title,
		"description": description,
		"series":      quizSeries,
		"category":    category,
		"subcategory": subcategory,
		"quiz_type":   "community",
		"is_visible":  false,
		"status":      "draft",
	})

	body, err := h.db.Insert("quizzes", quizData, true)
	if err != nil {
		fmt.Printf(">>> QUIZ CREATE ERROR: %v\n", err)
		return c.Redirect("/explore")
	}
	var quizzes []map[string]interface{}
	json.Unmarshal(body, &quizzes)
	fmt.Printf(">>> QUIZ CREATE OK: %d quizzes, body=%s\n", len(quizzes), string(body))

	if len(quizzes) > 0 {
		if id, ok := quizzes[0]["id"].(string); ok {
			return c.Redirect("/quiz/" + id + "/edit")
		}
	}
	return c.Redirect("/profil")
}

func (h *Handler) QuizEdit(c *fiber.Ctx) error {
	id := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	quiz, err := h.db.GetQuiz(id)
	if err != nil {
		return c.Redirect("/explore")
	}
	dbCreator := database.DBValue(quiz["creator_id"])
	if dbCreator != user.ID {
		return c.Redirect("/explore")
	}

	title, _ := quiz["title"].(string)
	description := database.DBValue(quiz["description"])
	isVisible := database.DBBool(quiz["is_visible"])
	questionCount := database.DBInt(quiz["question_count"])

	questionsBody, _ := h.db.Select("questions",
		fmt.Sprintf("quiz_id=eq.%s&order=order_index.asc&select=*,answers(*)", id), true)
	var questions []map[string]interface{}
	json.Unmarshal(questionsBody, &questions)

	questionsJSON, _ := json.Marshal(questions)

	publishBtn := `<button type="button" class="btn-primary" onclick="publishQuiz()" id="publish-btn">🚀 Publier le quiz</button>`
	if isVisible {
		publishBtn = `<button type="button" class="btn-outline" onclick="unpublishQuiz()" id="publish-btn" style="border-color:#fbbf24;color:#fbbf24">⏸️ Dépublier</button>`
	}

	quizEditHTML := `
<div class="qe-page">
    <a href="/quiz/mine" style="color:#6366f1;display:inline-block;margin-bottom:16px">← Retour</a>
    <div class="qe-header-row">
        <h1>✏️ Modifier le quiz</h1>
        <span class="qe-status-badge ` + func() string { if isVisible { return "qe-published" }; return "qe-draft" }() + `">` + func() string { if isVisible { return "Publié" }; return "Brouillon" }() + `</span>
    </div>

    <div class="qe-card">
        <form id="quiz-form" onsubmit="saveQuizMeta(event)">
            <div class="qe-field"><label>Titre</label><input type="text" id="quiz-title" value="` + htmlAttr(title) + `" class="pe-input"></div>
            <div class="qe-field"><label>Description</label><textarea id="quiz-desc" class="pe-input" rows="3">` + htmlAttr(description) + `</textarea></div>
            <button type="submit" class="btn-primary btn-sm">💾 Sauvegarder</button>
        </form>
    </div>

    <div class="qe-actions-bar">
        <h2>📋 Questions (` + fmt.Sprintf("%d", questionCount) + `)</h2>
        <div class="qe-actions-bar-btns">
            <button class="btn-outline btn-sm" onclick="showAddQuestion()">+ Ajouter</button>
            <button class="btn-outline btn-sm" onclick="showAIImport()">📝 Importer IA</button>
        </div>
    </div>

    <div id="add-question-panel" style="display:none" class="qe-card">
        <h3>Nouvelle question</h3>
        <div class="qe-field">
            <label>Type</label>
            <select id="new-q-type" class="pe-input" onchange="toggleNewQOptions()">
                <option value="text">Texte (QCM)</option>
                <option value="true_false">Vrai / Faux</option>
                <option value="image">Image (QCM)</option>
                <option value="gif">GIF (QCM)</option>
                <option value="audio">Audio (QCM)</option>
                <option value="matching">🔗 Relier (Matching)</option>
                <option value="fill_in">📝 Texte à trous</option>
            </select>
        </div>

        <div id="new-q-standard-fields">
            <div class="qe-field"><label>Question</label><textarea id="new-q-text" class="pe-input" rows="2" placeholder="Texte de la question..."></textarea></div>
            <div id="new-q-media-field" class="qe-field" style="display:none"><label>URL média</label><input type="text" id="new-q-media" class="pe-input" placeholder="https://..."></div>
            <div class="qe-field"><label>Temps limite (secondes)</label><input type="number" id="new-q-timer" class="pe-input" value="30" min="5" max="120" style="width:100px"></div>
            <div class="qe-field">
                <label>Réponses <span style="color:#94a3b8;font-weight:normal">(cocher la bonne)</span></label>
                <div id="new-q-answers">
                    <div class="qe-answer-row"><input type="radio" name="new-q-correct" value="0" checked><input type="text" class="pe-input" placeholder="Réponse 1"></div>
                    <div class="qe-answer-row"><input type="radio" name="new-q-correct" value="1"><input type="text" class="pe-input" placeholder="Réponse 2"></div>
                    <div class="qe-answer-row"><input type="radio" name="new-q-correct" value="2"><input type="text" class="pe-input" placeholder="Réponse 3"></div>
                    <div class="qe-answer-row"><input type="radio" name="new-q-correct" value="3"><input type="text" class="pe-input" placeholder="Réponse 4"></div>
                </div>
                <button type="button" class="btn-sm btn-ghost" onclick="addNewAnswerRow()">+ Ajouter réponse</button>
            </div>
            <div class="qe-field">
                <label>Distractors (mauvaises réponses alternatives) <span style="color:#94a3b8;font-weight:normal">jusqu'à 6, séparés par virgule</span></label>
                <input type="text" id="new-q-distractors" class="pe-input" placeholder="Faux1, Faux2, Faux3...">
            </div>
        </div>

        <div id="new-q-matching-fields" style="display:none">
            <div class="qe-field"><label>Question</label><textarea id="new-q-match-text" class="pe-input" rows="2" placeholder="Reliez les éléments..."></textarea></div>
            <div class="qe-field"><label>Temps limite (secondes)</label><input type="number" id="new-q-match-timer" class="pe-input" value="60" min="10" max="120" style="width:100px"></div>
            <div class="qe-field">
                <label>Paires <span style="color:#94a3b8;font-weight:normal">(élément → cible, séparés par ;)</span></label>
                <div id="matching-pairs-container">
                    <div class="qe-pair-row"><input type="text" class="pe-input" placeholder="Gauche (ex: Paris)" style="flex:1"> → <input type="text" class="pe-input" placeholder="Droite (ex: France)" style="flex:1"></div>
                </div>
                <button type="button" class="btn-sm btn-ghost" onclick="addMatchingPair()">+ Ajouter paire</button>
            </div>
        </div>

        <div id="new-q-fillin-fields" style="display:none">
            <div class="qe-field"><label>Phrase avec trous <span style="color:#94a3b8;font-weight:normal">(utilise {0}, {1}... pour les trous)</span></label>
                <textarea id="new-q-fillin-template" class="pe-input" rows="2" placeholder="Ex: Soleil et {0}"></textarea></div>
            <div class="qe-field"><label>Temps limite (secondes)</label><input type="number" id="new-q-fillin-timer" class="pe-input" value="30" min="5" max="120" style="width:100px"></div>
            <div class="qe-field">
                <label>Réponses pour chaque trou</label>
                <div id="fillin-blanks-container">
                    <div class="qe-fillin-row"><span style="color:#6366f1;font-weight:600;margin-right:8px">{0}:</span> <input type="text" class="pe-input" placeholder="Réponse" style="width:200px"> <input type="number" class="pe-input" placeholder="Lettres" style="width:80px" min="1"> <label style="display:inline-flex;align-items:center;gap:4px;font-size:.8rem;color:#94a3b8"><input type="checkbox" checked> Majuscules</label> <label style="display:inline-flex;align-items:center;gap:4px;font-size:.8rem;color:#94a3b8"><input type="checkbox" checked> Sans accents</label></div>
                </div>
                <button type="button" class="btn-sm btn-ghost" onclick="addFillInBlank()">+ Ajouter un trou</button>
            </div>
        </div>

        <div class="qe-form-actions">
            <button class="btn-outline btn-sm" onclick="hideAddQuestion()">Annuler</button>
            <button class="btn-primary btn-sm" onclick="addQuestion()">✅ Ajouter</button>
        </div>
    </div>

    <div id="ai-import-panel" style="display:none" class="qe-card">
        <h3>📝 Importer depuis une IA</h3>
        <p style="color:#94a3b8;font-size:.85rem;margin-bottom:12px">Colle le texte généré par ChatGPT, Claude ou une autre IA. Les questions seront parsées automatiquement.</p>
        <div class="qe-field"><label>Texte brut</label><textarea id="ai-raw-text" class="pe-input" rows="8" placeholder="Colle le texte ici..."></textarea></div>
        <div class="qe-form-actions">
            <button class="btn-outline btn-sm" onclick="hideAIImport()">Annuler</button>
            <button class="btn-primary btn-sm" onclick="parseImportedText()" id="ai-parse-btn">📝 Analyser le texte</button>
        </div>
        <div id="ai-loading" style="display:none;color:#94a3b8;text-align:center;padding:16px">Analyse en cours...</div>
        <div id="ai-parse-result" style="margin-top:12px"></div>
    </div>

    <div id="ai-format-guide" class="qe-card" style="margin-top:12px">
        <h4 style="font-size:.9rem;color:#94a3b8;margin-bottom:8px">Formats supportés</h4>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;font-size:.8rem">
            <div style="background:#0f172a;border:1px solid rgba(34,197,94,.3);border-radius:8px;padding:10px">
                <p style="color:#22c55e;font-weight:600;margin-bottom:4px">QCM</p>
                <pre style="color:#94a3b8;font-family:monospace;font-size:.7rem;white-space:pre-wrap;margin:0">1. Question ?
A) Réponse ✓
B) Autre</pre>
            </div>
            <div style="background:#0f172a;border:1px solid rgba(59,130,246,.3);border-radius:8px;padding:10px">
                <p style="color:#3b82f6;font-weight:600;margin-bottom:4px">Vrai / Faux</p>
                <pre style="color:#94a3b8;font-family:monospace;font-size:.7rem;white-space:pre-wrap;margin:0">2. Naruto est Hokage. Vrai</pre>
            </div>
            <div style="background:#0f172a;border:1px solid rgba(168,85,247,.3);border-radius:8px;padding:10px">
                <p style="color:#a855f7;font-weight:600;margin-bottom:4px">MATCHING</p>
                <pre style="color:#94a3b8;font-family:monospace;font-size:.7rem;white-space:pre-wrap;margin:0">Relier: Paris→France
Berlin→Allemagne</pre>
            </div>
            <div style="background:#0f172a;border:1px solid rgba(239,68,68,.3);border-radius:8px;padding:10px">
                <p style="color:#ef4444;font-weight:600;margin-bottom:4px">FILL IN</p>
                <pre style="color:#94a3b8;font-family:monospace;font-size:.7rem;white-space:pre-wrap;margin:0">{0} et {1}
Rép: Lune, Étoile</pre>
            </div>
        </div>
    </div>

    <div id="questions-container">
        <div id="questions-list"></div>
    </div>

    <div class="qe-publish-bar">
        ` + publishBtn + `
    </div>
</div>

<style>
.qe-page{max-width:700px;margin:0 auto}
.qe-header-row{display:flex;align-items:center;gap:12px;margin-bottom:20px}
.qe-header-row h1{margin:0}
.qe-status-badge{font-size:.75rem;padding:4px 10px;border-radius:4px;font-weight:600}
.qe-published{background:rgba(34,197,94,.15);color:#22c55e}
.qe-draft{background:rgba(251,191,36,.15);color:#fbbf24}
.qe-card{background:#16213e;border:1px solid #2d2d44;border-radius:10px;padding:16px;margin-bottom:16px}
.qe-field{margin-bottom:12px}
.qe-field label{display:block;font-size:.85rem;color:#94a3b8;margin-bottom:4px}
.qe-answer-row{display:flex;gap:8px;align-items:center;margin-bottom:6px}
.qe-answer-row input[type="radio"]{accent-color:#6366f1}
.qe-answer-row input[type="text"]{flex:1}
.qe-form-actions{display:flex;gap:8px;justify-content:flex-end}
.qe-actions-bar{display:flex;align-items:center;justify-content:space-between;margin-bottom:12px}
.qe-actions-bar h2{margin:0;font-size:1.1rem}
.qe-actions-bar-btns{display:flex;gap:8px}
.qe-question-card{background:#16213e;border:1px solid #2d2d44;border-radius:10px;padding:14px;margin-bottom:8px}
.qe-q-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:8px}
.qe-q-num{font-weight:700;color:#6366f1;font-size:.85rem}
.qe-q-type{font-size:.7rem;color:#94a3b8;background:#1e293b;padding:2px 8px;border-radius:4px}
.qe-q-actions{display:flex;gap:4px}
.qe-q-actions button{background:none;border:none;cursor:pointer;font-size:.85rem;padding:2px 6px;border-radius:4px}
.qe-q-actions button:hover{background:#1e293b}
.qe-answer-item{display:flex;gap:8px;align-items:center;margin-bottom:4px;font-size:.85rem}
.qe-answer-correct{color:#22c55e;font-weight:700}
.qe-answer-wrong{color:#64748b}
.qe-publish-bar{text-align:center;margin-top:24px;padding-top:16px;border-top:1px solid #2d2d44}
.qe-pair-row{display:flex;gap:8px;align-items:center;margin-bottom:6px}
.qe-pair-row input[type="text"]{flex:1}
.qe-fillin-row{display:flex;gap:8px;align-items:center;margin-bottom:6px;flex-wrap:wrap}
</style>
<script>
var QUIZ_ID='` + id + `';
var questions=` + string(questionsJSON) + `;

function renderQuestions(){
    var el=document.getElementById('questions-list');
    if(!questions||questions.length===0){el.innerHTML='<p style="color:#94a3b8;text-align:center;padding:20px">Aucune question. Ajoutez-en une !</p>';return;}
    var h='';
    for(var i=0;i<questions.length;i++){
        var q=questions[i];
        var answers=q.answers||[];
        var xp=q.xp_reward||10;
        h+='<div class="qe-question-card" id="qcard-'+q.id+'">';
        h+='<div class="qe-q-header"><span class="qe-q-num">#'+(i+1)+'</span><span class="qe-q-type">'+q.question_type+'</span>';
        h+='<span style="font-size:.75rem;color:#fbbf24;background:#1e293b;padding:2px 8px;border-radius:4px">';
        h+='⭐ <input type="number" value="'+xp+'" min="0" max="999" id="xp-input-'+q.id+'" style="width:50px;background:transparent;border:1px solid #334155;color:#fbbf24;font-size:.75rem;text-align:center;border-radius:3px;padding:1px" onchange="updateQuestionXp(\''+q.id+'\')"> XP';
        h+='</span>';
        h+='<div class="qe-q-actions"><button title="Modifier" onclick="editQuestion(\''+q.id+'\')">✏️</button><button title="Supprimer" onclick="deleteQuestion(\''+q.id+'\')">🗑️</button></div></div>';
        h+='<div id="qview-'+q.id+'">';
        h+='<div style="font-weight:600;margin-bottom:6px">'+escapeHtml(q.question_text)+'</div>';
        if(q.media_url){h+='<div style="font-size:.8rem;color:#94a3b8;margin-bottom:6px">📷 '+q.media_url+'</div>';}
        if(q.question_type==='matching'){
            var opts=typeof q.options==='string'?JSON.parse(q.options):(q.options||{});
            var pairs=opts.pairs||[];
            h+='<div style="font-size:.8rem;color:#94a3b8">';
            for(var j=0;j<pairs.length;j++){
                h+='<div>🔗 '+escapeHtml(pairs[j].x)+' → '+escapeHtml(pairs[j].y)+'</div>';
            }
            h+='</div>';
        } else if(q.question_type==='fill_in'){
            var opts=typeof q.options==='string'?JSON.parse(q.options):(q.options||{});
            h+='<div style="font-size:.8rem;color:#94a3b8">📝 '+escapeHtml(opts.template||'')+'</div>';
        } else {
            h+='<div class="qe-answers">';
            for(var j=0;j<answers.length;j++){
                var a=answers[j];
                var cls=a.is_correct?'qe-answer-correct':'qe-answer-wrong';
                var icon=a.is_correct?'✅':'○';
                h+='<div class="qe-answer-item"><span class="'+cls+'">'+icon+'</span> '+escapeHtml(a.answer_text)+'</div>';
            }
            h+='</div>';
        }
        h+='</div>';
        h+='<div id="qedit-'+q.id+'" style="display:none"></div>';
        h+='</div>';
    }
    el.innerHTML=h;
}

function updateQuestionXp(qid){
    var inp=document.getElementById('xp-input-'+qid);
    if(!inp)return;
    var val=parseInt(inp.value,10);
    if(isNaN(val)||val<0)val=0;
    if(val>999)val=999;
    inp.value=val;
    fetch('/api/quiz/'+QUIZ_ID+'/questions/'+qid+'/xp',{
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body:JSON.stringify({xp_reward:val})
    }).then(function(r){return r.json();}).then(function(d){
        if(d.success){inp.style.borderColor='#22c55e';setTimeout(function(){inp.style.borderColor='#334155';},1000);}
    });
}

function editQuestion(qid){
    var q=getQById(qid);
    if(!q)return;
    document.getElementById('qview-'+qid).style.display='none';
    var editEl=document.getElementById('qedit-'+qid);
    var answers=q.answers||[];
    var opts=typeof q.options==='string'?JSON.parse(q.options):(q.options||{});
    var h='<div style="margin-bottom:8px"><input type="text" id="eq-text-'+qid+'" value="'+escapeAttr(q.question_text)+'" class="pe-input" placeholder="Texte de la question"></div>';
    h+='<div style="margin-bottom:8px"><input type="text" id="eq-media-'+qid+'" value="'+(q.media_url?escapeAttr(q.media_url):'')+'" class="pe-input" placeholder="URL média (optionnel)"></div>';
    if(q.question_type==='matching'){
        var pairs=opts.pairs||[];
        h+='<div id="eq-pairs-'+qid+'">';
        for(var j=0;j<pairs.length;j++){
            h+='<div class="qe-pair-row">';
            h+='<input type="text" class="pe-input" id="eq-px-'+qid+'-'+j+'" value="'+escapeAttr(pairs[j].x)+'" placeholder="Gauche" style="flex:1">';
            h+=' → ';
            h+='<input type="text" class="pe-input" id="eq-py-'+qid+'-'+j+'" value="'+escapeAttr(pairs[j].y)+'" placeholder="Droite" style="flex:1">';
            h+='<button onclick="removeEditPair(\''+qid+'\','+j+')" style="background:none;border:none;color:#ef4444;cursor:pointer">✕</button>';
            h+='</div>';
        }
        h+='</div>';
        h+='<button class="btn-outline btn-sm" onclick="addEditPair(\''+qid+'\')">+ Ajouter paire</button>';
    } else if(q.question_type==='fill_in'){
        h+='<div style="margin-bottom:8px"><input type="text" id="eq-template-'+qid+'" value="'+escapeAttr(opts.template||'')+'" class="pe-input" placeholder="Template (ex: {0} est un {1})"></div>';
        var blanks=opts.blanks||[];
        h+='<div style="font-size:.8rem;color:#94a3b8;margin-bottom:4px">Trous ({'+qid+'}...): éditez directement dans le template ci-dessus.</div>';
    } else {
        h+='<div id="eq-answers-'+qid+'">';
        for(var j=0;j<answers.length;j++){
            var a=answers[j];
            var checked=a.is_correct?'checked':'';
            h+='<div class="qe-answer-row">';
            h+='<input type="radio" name="eq-correct-'+qid+'" value="'+j+'" '+checked+'>';
            h+='<input type="text" class="pe-input" id="eq-atxt-'+qid+'-'+j+'" value="'+escapeAttr(a.answer_text)+'" placeholder="Réponse '+(j+1)+'" style="flex:1">';
            h+='<button onclick="removeEditAnswer(\''+qid+'\','+j+')" style="background:none;border:none;color:#ef4444;cursor:pointer;font-size:.85rem">✕</button>';
            h+='</div>';
        }
        h+='</div>';
        h+='<button class="btn-outline btn-sm" onclick="addEditAnswer(\''+qid+'\')">+ Ajouter réponse</button>';
    }
    h+='<div style="display:flex;gap:8px;margin-top:10px;justify-content:flex-end">';
    h+='<button class="btn-ghost btn-sm" onclick="cancelEdit(\''+qid+'\')">Annuler</button>';
    h+='<button class="btn-primary btn-sm" onclick="saveQuestion(\''+qid+'\')">💾 Enregistrer</button>';
    h+='</div>';
    editEl.innerHTML=h;
    editEl.style.display='block';
}

function getQById(qid){
    for(var i=0;i<questions.length;i++){
        if(questions[i].id===qid)return questions[i];
    }
    return null;
}

function escapeAttr(s){
    if(!s)return'';
    return s.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/'/g,'&#39;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function cancelEdit(qid){
    document.getElementById('qedit-'+qid).style.display='none';
    document.getElementById('qview-'+qid).style.display='block';
}

function saveQuestion(qid){
    var q=getQById(qid);
    if(!q)return;
    var text=document.getElementById('eq-text-'+qid).value.trim();
    if(!text){alert('Écris la question');return;}
    var media=document.getElementById('eq-media-'+qid).value.trim();
    var payload={question_text:text,media_url:media,time_limit_seconds:q.time_limit_seconds||30,answers:[],options:null,distractors:[]};
    if(q.question_type==='matching'){
        var pairRows=document.getElementById('eq-pairs-'+qid).children;
        var pairs=[];
        for(var j=0;j<pairRows.length;j++){
            var px=document.getElementById('eq-px-'+qid+'-'+j);
            var py=document.getElementById('eq-py-'+qid+'-'+j);
            if(px&&py&&px.value.trim()&&py.value.trim())pairs.push({id:'p'+(j+1),x:px.value.trim(),y:py.value.trim()});
        }
        if(pairs.length<2){alert('Ajoute au moins 2 paires');return;}
        payload.options=JSON.stringify({type:'matching',pairs:pairs});
    } else if(q.question_type==='fill_in'){
        var template=document.getElementById('eq-template-'+qid).value.trim();
        if(!template){alert('Écris le template');return;}
        payload.question_text=template;
        var blanks=[];var phIdx=0;
        var tpl=template;var newBlanks=[];
        while(tpl.indexOf('{'+phIdx+'}')!==-1){phIdx++;}
        // Reconstruct blanks from existing options
        var oldBlanks=(typeof q.options==='string'?JSON.parse(q.options):q.options||{}).blanks||[];
        for(var bi=0;bi<oldBlanks.length;bi++){
            newBlanks.push({id:bi,answer:oldBlanks[bi].answer||'',char_count:oldBlanks[bi].char_count||1,case_sensitive:oldBlanks[bi].case_sensitive!==false,accept_without_accents:oldBlanks[bi].accept_without_accents!==false});
        }
        payload.options=JSON.stringify({type:'fill_in',template:tpl,blanks:newBlanks});
    } else {
        var answerRows=document.getElementById('eq-answers-'+qid).children;
        var ans=[];
        for(var j=0;j<answerRows.length;j++){
            var radio=answerRows[j].querySelector('input[type="radio"]');
            var inp=document.getElementById('eq-atxt-'+qid+'-'+j);
            if(inp&&inp.value.trim())ans.push({answer_text:inp.value.trim(),is_correct:radio&&radio.checked,order_index:j});
        }
        if(ans.length<2){alert('Ajoute au moins 2 réponses');return;}
        payload.answers=ans;
    }
    fetch('/api/quiz/'+QUIZ_ID+'/questions/'+qid,{
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body:JSON.stringify(payload)
    }).then(function(r){return r.json();}).then(function(d){
        if(d.success){location.reload();}else{alert(d.error||'Erreur');}
    });
}

function addEditAnswer(qid){
    var el=document.getElementById('eq-answers-'+qid);
    var n=el.children.length;
    var div=document.createElement('div');div.className='qe-answer-row';
    div.innerHTML='<input type="radio" name="eq-correct-'+qid+'" value="'+n+'"><input type="text" class="pe-input" id="eq-atxt-'+qid+'-'+n+'" placeholder="Réponse '+(n+1)+'" style="flex:1"><button onclick="removeEditAnswer(\''+qid+'\','+n+')" style="background:none;border:none;color:#ef4444;cursor:pointer;font-size:.85rem">✕</button>';
    el.appendChild(div);
}

function removeEditAnswer(qid,idx){
    var el=document.getElementById('eq-answers-'+qid);
    // Rebuild without the removed row
    var rows=el.children;
    var h='';var newIdx=0;
    for(var j=0;j<rows.length;j++){
        if(j===idx)continue;
        var radio=rows[j].querySelector('input[type="radio"]');
        var inp=rows[j].querySelector('input[type="text"]');
        if(inp){
            var checked=radio&&radio.checked?'checked':'';
            h+='<div class="qe-answer-row">';
            h+='<input type="radio" name="eq-correct-'+qid+'" value="'+newIdx+'" '+checked+'>';
            h+='<input type="text" class="pe-input" id="eq-atxt-'+qid+'-'+newIdx+'" value="'+escapeAttr(inp.value)+'" placeholder="Réponse '+(newIdx+1)+'" style="flex:1">';
            h+='<button onclick="removeEditAnswer(\''+qid+'\','+newIdx+')" style="background:none;border:none;color:#ef4444;cursor:pointer;font-size:.85rem">✕</button>';
            h+='</div>';
            newIdx++;
        }
    }
    el.innerHTML=h;
}

function addEditPair(qid){
    var el=document.getElementById('eq-pairs-'+qid);
    var n=el.children.length;
    var div=document.createElement('div');div.className='qe-pair-row';
    div.innerHTML='<input type="text" class="pe-input" id="eq-px-'+qid+'-'+n+'" placeholder="Gauche" style="flex:1"> → <input type="text" class="pe-input" id="eq-py-'+qid+'-'+n+'" placeholder="Droite" style="flex:1"><button onclick="removeEditPair(\''+qid+'\','+n+')" style="background:none;border:none;color:#ef4444;cursor:pointer">✕</button>';
    el.appendChild(div);
}

function removeEditPair(qid,idx){
    var el=document.getElementById('eq-pairs-'+qid);
    var rows=el.children;
    var h='';var newIdx=0;
    for(var j=0;j<rows.length;j++){
        if(j===idx)continue;
        var px=rows[j].querySelector('input[placeholder="Gauche"]');
        var py=rows[j].querySelector('input[placeholder="Droite"]');
        if(px&&py){
            h+='<div class="qe-pair-row">';
            h+='<input type="text" class="pe-input" id="eq-px-'+qid+'-'+newIdx+'" value="'+escapeAttr(px.value)+'" placeholder="Gauche" style="flex:1">';
            h+=' → ';
            h+='<input type="text" class="pe-input" id="eq-py-'+qid+'-'+newIdx+'" value="'+escapeAttr(py.value)+'" placeholder="Droite" style="flex:1">';
            h+='<button onclick="removeEditPair(\''+qid+'\','+newIdx+')" style="background:none;border:none;color:#ef4444;cursor:pointer">✕</button>';
            h+='</div>';
            newIdx++;
        }
    }
    el.innerHTML=h;
}

function escapeHtml(t){var d=document.createElement('div');d.textContent=t;return d.innerHTML;}

function showAddQuestion(){document.getElementById('add-question-panel').style.display='block';}
function hideAddQuestion(){document.getElementById('add-question-panel').style.display='none';}
function showAIImport(){document.getElementById('ai-import-panel').style.display='block';}
function hideAIImport(){document.getElementById('ai-import-panel').style.display='none';}

function toggleNewQOptions(){
    var t=document.getElementById('new-q-type').value;
    document.getElementById('new-q-media-field').style.display=(t==='image'||t==='gif'||t==='audio')?'block':'none';
    document.getElementById('new-q-standard-fields').style.display=(t==='matching'||t==='fill_in')?'none':'block';
    document.getElementById('new-q-matching-fields').style.display=t==='matching'?'block':'none';
    document.getElementById('new-q-fillin-fields').style.display=t==='fill_in'?'block':'none';
}

function addNewAnswerRow(){
    var el=document.getElementById('new-q-answers');
    var n=el.children.length;
    var row=document.createElement('div');row.className='qe-answer-row';
    row.innerHTML='<input type="radio" name="new-q-correct" value="'+n+'"><input type="text" class="pe-input" placeholder="Réponse '+(n+1)+'">';
    el.appendChild(row);
}

function addMatchingPair(){
    var el=document.getElementById('matching-pairs-container');
    var row=document.createElement('div');row.className='qe-pair-row';
    row.innerHTML='<input type="text" class="pe-input" placeholder="Gauche" style="flex:1"> → <input type="text" class="pe-input" placeholder="Droite" style="flex:1">';
    el.appendChild(row);
}

function addFillInBlank(){
    var el=document.getElementById('fillin-blanks-container');
    var n=el.children.length;
    var row=document.createElement('div');row.className='qe-fillin-row';
    row.innerHTML='<span style="color:#6366f1;font-weight:600;margin-right:8px">{'+n+'}:</span> <input type="text" class="pe-input" placeholder="Réponse" style="width:200px"> <input type="number" class="pe-input" placeholder="Lettres" style="width:80px" min="1"> <label style="display:inline-flex;align-items:center;gap:4px;font-size:.8rem;color:#94a3b8"><input type="checkbox" checked> Majuscules</label> <label style="display:inline-flex;align-items:center;gap:4px;font-size:.8rem;color:#94a3b8"><input type="checkbox" checked> Sans accents</label>';
    el.appendChild(row);
}

function addQuestion(){
    var type=document.getElementById('new-q-type').value;

    if(type==='matching'){
        return addMatchingQuestion();
    }
    if(type==='fill_in'){
        return addFillInQuestion();
    }

    var text=document.getElementById('new-q-text').value.trim();
    if(!text){alert('Écris la question');return;}
    var media=document.getElementById('new-q-media')?document.getElementById('new-q-media').value.trim():'';
    var timer=parseInt(document.getElementById('new-q-timer').value)||30;
    var answerEls=document.querySelectorAll('#new-q-answers .qe-answer-row');
    var answers=[];
    for(var i=0;i<answerEls.length;i++){
        var radio=answerEls[i].querySelector('input[type="radio"]');
        var input=answerEls[i].querySelector('input[type="text"]');
        if(input&&input.value.trim())answers.push({answer_text:input.value.trim(),is_correct:radio&&radio.checked,order_index:i});
    }
    if(answers.length<2){alert('Ajoute au moins 2 réponses');return;}

    var distractorStr=document.getElementById('new-q-distractors')?document.getElementById('new-q-distractors').value.trim():'';
    var distractors=distractorStr?distractorStr.split(',').map(function(s){return s.trim();}).filter(function(s){return s;}):[];

    fetch('/api/quiz/'+QUIZ_ID+'/questions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({
        question_text:text,question_type:type,media_url:media,time_limit_seconds:timer,answers:answers,distractors:distractors
    })}).then(function(r){return r.json()}).then(function(d){
        if(d.success){location.reload();}else{alert(d.error||'Erreur');}
    });
}

function addMatchingQuestion(){
    var text=document.getElementById('new-q-match-text').value.trim();
    if(!text){alert('Écris la question');return;}
    var timer=parseInt(document.getElementById('new-q-match-timer').value)||60;
    var pairEls=document.querySelectorAll('#matching-pairs-container .qe-pair-row');
    var pairs=[];
    for(var i=0;i<pairEls.length;i++){
        var inputs=pairEls[i].querySelectorAll('input[type="text"]');
        var x=inputs[0]?inputs[0].value.trim():'';
        var y=inputs[1]?inputs[1].value.trim():'';
        if(x&&y)pairs.push({id:'p'+(i+1),x:x,y:y});
    }
    if(pairs.length<2){alert('Ajoute au moins 2 paires');return;}

    fetch('/api/quiz/'+QUIZ_ID+'/questions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({
        question_text:text,question_type:'matching',time_limit_seconds:timer,answers:[],
        options:JSON.stringify({type:'matching',pairs:pairs})
    })}).then(function(r){return r.json()}).then(function(d){
        if(d.success){location.reload();}else{alert(d.error||'Erreur');}
    });
}

function addFillInQuestion(){
    var template=document.getElementById('new-q-fillin-template').value.trim();
    if(!template){alert('Écris la phrase avec des trous');return;}
    var timer=parseInt(document.getElementById('new-q-fillin-timer').value)||30;
    var blankEls=document.querySelectorAll('#fillin-blanks-container .qe-fillin-row');
    var blanks=[];
    var phIdx=0;
    for(var i=0;i<blankEls.length;i++){
        var answerInput=blankEls[i].querySelector('input[type="text"]');
        var countInput=blankEls[i].querySelector('input[type="number"]');
        var checkboxes=blankEls[i].querySelectorAll('input[type="checkbox"]');
        var answer=answerInput?answerInput.value.trim():'';
        var charCount=parseInt(countInput?countInput.value:'0')||answer.length;
        var caseSensitive=checkboxes.length>0?!checkboxes[0].checked:true;
        var acceptNoAccents=checkboxes.length>1?checkboxes[1].checked:true;
        if(answer)blanks.push({id:phIdx,answer:answer,char_count:charCount,case_sensitive:caseSensitive,accept_without_accents:acceptNoAccents});
        phIdx++;
    }
    if(blanks.length===0){alert('Ajoute au moins un trou avec une réponse');return;}
    // Auto-replace ? with {0}, {1}... in template
    var tpl=template;
    for(var bi=0;bi<blanks.length;bi++){
        var idx=tpl.indexOf('?');
        if(idx===-1)break;
        tpl=tpl.substring(0,idx)+'{'+bi+'}'+tpl.substring(idx+1);
    }
    fetch('/api/quiz/'+QUIZ_ID+'/questions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({
        question_text:template,question_type:'fill_in',time_limit_seconds:timer,answers:[],
        options:JSON.stringify({type:'fill_in',template:tpl,blanks:blanks})
    })}).then(function(r){return r.json()}).then(function(d){
        if(d.success){location.reload();}else{alert(d.error||'Erreur');}
    });
}

function deleteQuestion(qid){
    if(!confirm('Supprimer cette question ?'))return;
    fetch('/api/quiz/'+QUIZ_ID+'/questions/'+qid,{method:'DELETE'}).then(function(r){return r.json()}).then(function(d){
        if(d.success)location.reload();
    });
}

function saveQuizMeta(e){
    e.preventDefault();
    var t=document.getElementById('quiz-title').value.trim();
    var d=document.getElementById('quiz-desc').value.trim();
    fetch('/api/quiz/'+QUIZ_ID+'/meta',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:t,description:d})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){alert('Sauvegardé !');}else{alert(d.error||'Erreur');}
    });
}

function publishQuiz(){
    fetch('/api/quiz/'+QUIZ_ID+'/publish',{method:'POST'}).then(function(r){return r.json()}).then(function(d){
        if(d.success)location.reload();
    });
}
function unpublishQuiz(){
    fetch('/api/quiz/'+QUIZ_ID+'/publish',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({visible:false})}).then(function(r){return r.json()}).then(function(d){
        if(d.success)location.reload();
    });
}

function parseImportedText(){
    var text=document.getElementById('ai-raw-text').value.trim();
    if(!text){alert('Colle d\'abord le texte à analyser');return;}
    document.getElementById('ai-loading').style.display='block';
    document.getElementById('ai-parse-btn').disabled=true;
    document.getElementById('ai-parse-result').innerHTML='';
    var questions=[];
    var blocks=text.split(/\n(?=\d+[).]\s*)/);
    for(var b=0;b<blocks.length;b++){
        var block=blocks[b].trim();
        if(!block)continue;
        var tf=block.match(/(?:\d+[).]\s*)?(.+?)\.?\s*(Vrai|Faux|True|False)\s*$/i);
        if(tf){
            var isTrue=tf[2].toLowerCase()==='vrai'||tf[2].toLowerCase()==='true';
            questions.push({question_text:tf[1].replace(/^\d+[).]\s*/,'').trim(),question_type:'true_false',answers:[{answer_text:'Vrai',is_correct:isTrue,order_index:0},{answer_text:'Faux',is_correct:!isTrue,order_index:1}]});
            continue;
        }
        var lines=block.split('\n').filter(function(l){return l.trim();});
        if(lines.length<2)continue;
        var qt=lines[0].replace(/^\d+[).]\s*/,'').trim();
        var ans=[];
        for(var i=1;i<lines.length;i++){
            var m=lines[i].trim().match(/^[A-Za-z][).]\s*(.+)/);
            if(m){
                var t=m[1].trim();
                var ok=t.indexOf('✓')>=0||t.indexOf('✔')>=0||t.toLowerCase().indexOf('(correct)')>=0||t.toLowerCase().indexOf('[correct]')>=0;
                t=t.replace(/[✓✔]/g,'').replace(/\(correct\)/gi,'').replace(/\[correct\]/gi,'').trim();
                if(t)ans.push({answer_text:t,is_correct:ok,order_index:ans.length});
            }
        }
        if(ans.length>=2&&ans.some(function(a){return a.is_correct;}))questions.push({question_text:qt,question_type:'text',answers:ans});
    }
    if(questions.length===0){
        document.getElementById('ai-loading').style.display='none';
        document.getElementById('ai-parse-btn').disabled=false;
        document.getElementById('ai-parse-result').innerHTML='<div style="padding:12px;background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:8px;color:#ef4444;font-size:.85rem">Aucune question detectee. Verifie le format.</div>';
        return;
    }
    var total=questions.length;
    var done=0;
    var resultEl=document.getElementById('ai-parse-result');
    resultEl.innerHTML='<div style="padding:12px;background:rgba(34,197,94,.1);border:1px solid rgba(34,197,94,.3);border-radius:8px;color:#22c55e;font-size:.85rem">0/'+total+' questions importees...</div>';
    function importNext(idx){
        if(idx>=questions.length){
            document.getElementById('ai-loading').style.display='none';
            document.getElementById('ai-parse-btn').disabled=false;
            resultEl.innerHTML='<div style="padding:12px;background:rgba(34,197,94,.1);border:1px solid rgba(34,197,94,.3);border-radius:8px;color:#22c55e;font-size:.85rem">'+total+' question(s) importee(s) !</div>';
            setTimeout(function(){location.reload();},1500);
            return;
        }
        var q=questions[idx];
        fetch('/api/quiz/'+QUIZ_ID+'/questions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({question_text:q.question_text,question_type:q.question_type,time_limit_seconds:30,answers:q.answers})}).then(function(r){return r.json()}).then(function(d){
            if(d.success){resultEl.innerHTML='<div style="padding:12px;background:rgba(34,197,94,.1);border:1px solid rgba(34,197,94,.3);border-radius:8px;color:#22c55e;font-size:.85rem">'+(done+1)+'/'+total+' questions importees...</div>';done++;importNext(idx+1);}
            else{resultEl.innerHTML='<div style="padding:12px;background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:8px;color:#ef4444;font-size:.85rem">Erreur: '+(d.error||'inconnue')+'</div>';document.getElementById('ai-loading').style.display='none';document.getElementById('ai-parse-btn').disabled=false;}
        });
    }
    importNext(0);
}

renderQuestions();
</script>
`
	return renderWithContent(c, "Modifier le quiz", quizEditHTML)
}

func (h *Handler) APIQuizAddQuestion(c *fiber.Ctx) error {
	quizID := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type AnswerReq struct {
		AnswerText string `json:"answer_text"`
		IsCorrect  bool   `json:"is_correct"`
		OrderIndex int    `json:"order_index"`
	}
	type Request struct {
		QuestionText     string          `json:"question_text"`
		QuestionType     string          `json:"question_type"`
		MediaURL         string          `json:"media_url"`
		TimeLimitSeconds int             `json:"time_limit_seconds"`
		Answers          []AnswerReq     `json:"answers"`
		Options          json.RawMessage `json:"options"`
		Distractors      []string        `json:"distractors"`
	}
	var req Request
	c.BodyParser(&req)

	if req.QuestionText == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Question requise"})
	}
	if req.QuestionType == "" {
		req.QuestionType = "text"
	}
	if req.TimeLimitSeconds < 5 {
		req.TimeLimitSeconds = 30
	}

	xpDefaults := map[string]int{
		"text": 10, "true_false": 10, "image": 10, "image_shadow": 10,
		"gif": 10, "audio": 10, "character_guess": 50, "impostor": 50,
		"fill_in": 20, "matching": 30,
	}
	xpReward := xpDefaults[req.QuestionType]
	if xpReward == 0 {
		xpReward = 10
	}

	countBody, _ := h.db.Select("questions",
		fmt.Sprintf("quiz_id=eq.%s&select=id", quizID), true)
	var existing []map[string]interface{}
	json.Unmarshal(countBody, &existing)
	nextOrder := len(existing)

	qDataMap := map[string]interface{}{
		"quiz_id":            quizID,
		"question_text":      req.QuestionText,
		"question_type":      req.QuestionType,
		"media_url":          req.MediaURL,
		"time_limit_seconds": req.TimeLimitSeconds,
		"order_index":        nextOrder,
		"xp_reward":          xpReward,
	}
	if len(req.Options) > 0 {
		var optsObj interface{}
		if err := json.Unmarshal(req.Options, &optsObj); err == nil {
			qDataMap["options"] = optsObj
		} else {
			qDataMap["options"] = string(req.Options)
		}
	}
	if len(req.Distractors) > 0 {
		qDataMap["distractors"] = req.Distractors
	}
	qData, _ := json.Marshal(qDataMap)

	qBody, err := h.db.Insert("questions", qData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var qRows []map[string]interface{}
	json.Unmarshal(qBody, &qRows)
	if len(qRows) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur insertion"})
	}
	qID := database.DBValue(qRows[0]["id"])

	for i, a := range req.Answers {
		if a.AnswerText == "" {
			continue
		}
		aData, _ := json.Marshal(map[string]interface{}{
			"question_id": qID,
			"answer_text": a.AnswerText,
			"is_correct":  a.IsCorrect,
			"order_index": i,
		})
		h.db.Insert("answers", aData, true)
	}

	h.db.Update("quizzes",
		fmt.Sprintf("id=eq.%s", quizID),
		[]byte(fmt.Sprintf(`{"question_count":%d}`, nextOrder+1)), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIQuizDeleteQuestion(c *fiber.Ctx) error {
	quizID := c.Params("id")
	qID := c.Params("qid")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	h.db.Delete("answers", fmt.Sprintf("question_id=eq.%s", qID), true)
	h.db.Delete("questions", fmt.Sprintf("id=eq.%s&quiz_id=eq.%s", qID, quizID), true)

	countBody, _ := h.db.Select("questions",
		fmt.Sprintf("quiz_id=eq.%s&select=id", quizID), true)
	var remaining []map[string]interface{}
	json.Unmarshal(countBody, &remaining)
	h.db.Update("quizzes",
		fmt.Sprintf("id=eq.%s", quizID),
		[]byte(fmt.Sprintf(`{"question_count":%d}`, len(remaining))), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIQuizUpdateQuestionXp(c *fiber.Ctx) error {
	quizID := c.Params("id")
	qID := c.Params("qid")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		XpReward int `json:"xp_reward"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.XpReward < 0 {
		req.XpReward = 0
	}
	if req.XpReward > 999 {
		req.XpReward = 999
	}

	updateData, _ := json.Marshal(map[string]interface{}{
		"xp_reward": req.XpReward,
	})
	h.db.Update("questions", fmt.Sprintf("id=eq.%s&quiz_id=eq.%s", qID, quizID), updateData, true)

	return c.JSON(fiber.Map{"success": true, "xp_reward": req.XpReward})
}

func (h *Handler) APIQuizUpdateQuestion(c *fiber.Ctx) error {
	quizID := c.Params("id")
	qID := c.Params("qid")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type AnswerReq struct {
		AnswerText string `json:"answer_text"`
		IsCorrect  bool   `json:"is_correct"`
		OrderIndex int    `json:"order_index"`
	}
	type Request struct {
		QuestionText     string          `json:"question_text"`
		MediaURL         string          `json:"media_url"`
		TimeLimitSeconds int             `json:"time_limit_seconds"`
		Options          json.RawMessage `json:"options"`
		Distractors      []string        `json:"distractors"`
		Answers          []AnswerReq     `json:"answers"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.QuestionText == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Question requise"})
	}
	if req.TimeLimitSeconds < 5 {
		req.TimeLimitSeconds = 30
	}

	qData := map[string]interface{}{
		"question_text":      req.QuestionText,
		"media_url":          req.MediaURL,
		"time_limit_seconds": req.TimeLimitSeconds,
	}
	if len(req.Options) > 0 {
		var optsObj interface{}
		if err := json.Unmarshal(req.Options, &optsObj); err == nil {
			qData["options"] = optsObj
		} else {
			qData["options"] = string(req.Options)
		}
	}
	if len(req.Distractors) > 0 {
		qData["distractors"] = req.Distractors
	} else {
		qData["distractors"] = nil
	}
	qJSON, _ := json.Marshal(qData)
	h.db.Update("questions", fmt.Sprintf("id=eq.%s&quiz_id=eq.%s", qID, quizID), qJSON, true)

	h.db.Delete("answers", fmt.Sprintf("question_id=eq.%s", qID), true)
	for i, a := range req.Answers {
		if a.AnswerText == "" {
			continue
		}
		aData, _ := json.Marshal(map[string]interface{}{
			"question_id": qID,
			"answer_text": a.AnswerText,
			"is_correct":  a.IsCorrect,
			"order_index": i,
		})
		h.db.Insert("answers", aData, true)
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIQuizUpdateMeta(c *fiber.Ctx) error {
	quizID := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}
	_ = quiz

	type Request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	var req Request
	c.BodyParser(&req)

	updateData, _ := json.Marshal(map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
	})
	h.db.Update("quizzes", fmt.Sprintf("id=eq.%s", quizID), updateData, true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIQuizPublish(c *fiber.Ctx) error {
	quizID := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		Visible *bool `json:"visible"`
	}
	var req Request
	c.BodyParser(&req)

	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}

	questionCount := database.DBInt(quiz["question_count"])
	if visible && questionCount == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Ajoute au moins une question"})
	}

	updateData, _ := json.Marshal(map[string]interface{}{
		"is_visible": visible,
		"status":     func() string { if visible { return "published" }; return "draft" }(),
	})
	h.db.Update("quizzes", fmt.Sprintf("id=eq.%s", quizID), updateData, true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIQuizArchive(c *fiber.Ctx) error {
	quizID := c.Params("id")
	log.Printf("[APIQuizArchive] user=%s quiz=%s", c.Locals("user_id"), quizID)
	user := h.getUserFromSession(c)
	if user == nil {
		log.Printf("[APIQuizArchive] user not authenticated, redirecting to login")
		return c.Redirect("/login")
	}
	log.Printf("[APIQuizArchive] user=%s authenticated", user.ID)

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil {
		log.Printf("[APIQuizArchive] GetQuiz error: %v", err)
		return c.Redirect("/quiz/mine")
	}
	if database.DBValue(quiz["creator_id"]) != user.ID {
		log.Printf("[APIQuizArchive] forbidden: creator=%s != user=%s", database.DBValue(quiz["creator_id"]), user.ID)
		return c.Redirect("/quiz/mine")
	}

	updateData, _ := json.Marshal(map[string]interface{}{
		"status":     "archived",
		"is_visible": false,
	})
	h.db.Update("quizzes", fmt.Sprintf("id=eq.%s", quizID), updateData, true)
	log.Printf("[APIQuizArchive] quiz %s archived successfully", quizID)

	return c.Redirect("/quiz/mine")
}

func (h *Handler) APIQuizDelete(c *fiber.Ctx) error {
	quizID := c.Params("id")
	log.Printf("[APIQuizDelete] user=%s quiz=%s", c.Locals("user_id"), quizID)
	user := h.getUserFromSession(c)
	if user == nil {
		log.Printf("[APIQuizDelete] user not authenticated, redirecting to login")
		return c.Redirect("/login")
	}
	log.Printf("[APIQuizDelete] user=%s authenticated", user.ID)

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil {
		log.Printf("[APIQuizDelete] GetQuiz error: %v", err)
		return c.Redirect("/quiz/mine")
	}
	if database.DBValue(quiz["creator_id"]) != user.ID {
		log.Printf("[APIQuizDelete] forbidden: creator=%s != user=%s", database.DBValue(quiz["creator_id"]), user.ID)
		return c.Redirect("/quiz/mine")
	}

	// Supprimer les réponses associées aux questions du quiz
	qBody, _ := h.db.Select("questions", fmt.Sprintf("quiz_id=eq.%s&select=id", quizID), true)
	var questions []map[string]interface{}
	json.Unmarshal(qBody, &questions)
	log.Printf("[APIQuizDelete] quiz %s has %d questions", quizID, len(questions))
	for _, q := range questions {
		qid := database.DBValue(q["id"])
		h.db.Delete("answers", fmt.Sprintf("question_id=eq.%s", qid), true)
	}
	h.db.Delete("questions", fmt.Sprintf("quiz_id=eq.%s", quizID), true)
	h.db.Delete("quizzes", fmt.Sprintf("id=eq.%s", quizID), true)
	log.Printf("[APIQuizDelete] quiz %s permanently deleted", quizID)

	return c.Redirect("/quiz/mine")
}

func (h *Handler) APIQuizGenerateQuestions(c *fiber.Ctx) error {
	quizID := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Non connecté"})
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil || database.DBValue(quiz["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		Text string `json:"text"`
	}
	var req Request
	c.BodyParser(&req)

	if req.Text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Texte requis"})
	}

	questions := parseQuizText(req.Text)

	countBody, _ := h.db.Select("questions",
		fmt.Sprintf("quiz_id=eq.%s&select=id", quizID), true)
	var existing []map[string]interface{}
	json.Unmarshal(countBody, &existing)
	nextOrder := len(existing)

	added := 0
	for _, q := range questions {
		qData, _ := json.Marshal(map[string]interface{}{
			"quiz_id":            quizID,
			"question_text":      q["question_text"],
			"question_type":      q["question_type"],
			"time_limit_seconds": 30,
			"order_index":        nextOrder,
		})
		qBody, qErr := h.db.Insert("questions", qData, true)
		if qErr != nil {
			continue
		}
		var qRows []map[string]interface{}
		json.Unmarshal(qBody, &qRows)
		if len(qRows) == 0 {
			continue
		}
		qID := database.DBValue(qRows[0]["id"])
		nextOrder++

		answers, _ := q["answers"].([]map[string]interface{})
		for i, a := range answers {
			aData, _ := json.Marshal(map[string]interface{}{
				"question_id": qID,
				"answer_text": a["answer_text"],
				"is_correct":  a["is_correct"],
				"order_index": i,
			})
			h.db.Insert("answers", aData, true)
		}
		added++
	}

	h.db.Update("quizzes",
		fmt.Sprintf("id=eq.%s", quizID),
		[]byte(fmt.Sprintf(`{"question_count":%d}`, nextOrder)), true)

	return c.JSON(fiber.Map{"success": true, "count": added})
}

func parseQuizText(text string) []map[string]interface{} {
	var questions []map[string]interface{}

	blocks := strings.Split(text, "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) < 2 {
			continue
		}

		questionText := strings.TrimSpace(lines[0])
		questionText = strings.TrimPrefix(questionText, "Q:")
		questionText = strings.TrimPrefix(questionText, "q:")
		questionText = strings.TrimPrefix(questionText, "Question:")
		questionText = strings.TrimSpace(questionText)

		if questionText == "" {
			continue
		}

		var answers []map[string]interface{}
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			isCorrect := false
			if strings.HasPrefix(line, "*A:") || strings.HasPrefix(line, "*a:") {
				isCorrect = true
				line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "*A:"), "*a:"))
			} else if strings.HasPrefix(line, "A:") || strings.HasPrefix(line, "a:") {
				line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "A:"), "a:"))
			} else if strings.HasPrefix(line, "- ") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
			} else {
				continue
			}

			if line != "" {
				answers = append(answers, map[string]interface{}{
					"answer_text": line,
					"is_correct":  isCorrect,
				})
			}
		}

		if len(answers) >= 2 {
			qType := "text"
			if len(answers) == 2 {
				if (answers[0]["answer_text"] == "Vrai" || answers[0]["answer_text"] == "True") &&
					(answers[1]["answer_text"] == "Faux" || answers[1]["answer_text"] == "False") {
					qType = "true_false"
				}
			}
			questions = append(questions, map[string]interface{}{
				"question_text": questionText,
				"question_type": qType,
				"answers":       answers,
			})
		}
	}

	return questions
}

func (h *Handler) OfficialLeaderboard(c *fiber.Ctx) error {
	return c.Redirect("/leaderboard")
}

func (h *Handler) OfficialLeaderboardDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.Redirect("/leaderboard/quiz/" + id)
}

func (h *Handler) MyQuizzes(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	quizBody, _ := h.db.Select("quizzes",
		fmt.Sprintf("creator_id=eq.%s&order=created_at.desc&select=*", user.ID), true)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)

	var cards string
	for _, q := range quizzes {
		title := database.DBValue(q["title"])
		qType := database.DBValue(q["quiz_type"])
		qCount := database.DBInt(q["question_count"])
		isVisible := database.DBBool(q["is_visible"])
		qID := database.DBValue(q["id"])
		status := "Brouillon"
		statusClass := "qe-draft"
		if isVisible {
			status = "Publié"
			statusClass = "qe-published"
		}
		cards += fmt.Sprintf(`		<div class="qe-card">
			<div style="display:flex;justify-content:space-between;align-items:center">
				<div><strong>%s</strong><br><span style="color:#94a3b8;font-size:.8rem">%s · %d questions · %s</span></div>
				<div style="display:flex;align-items:center;gap:8px">
					<span class="qe-status-badge %s">%s</span>
					<div style="display:flex;gap:4px">
						<a href="/quiz/%s/edit" class="qe-icon-btn" title="Modifier">✏️</a>
						<form action="/quiz/%s/archive" method="POST" style="display:inline" onsubmit="return confirm('Archiver ce quiz ?')">
							<button type="submit" class="qe-icon-btn" title="Archiver">📦</button>
						</form>
						<form action="/quiz/%s/delete" method="POST" style="display:inline" onsubmit="return confirm('Supprimer définitivement ce quiz ?')">
							<button type="submit" class="qe-icon-btn qe-icon-btn-danger" title="Supprimer">🗑️</button>
						</form>
					</div>
				</div>
			</div>
		</div>`, htmlAttr(title), qType, qCount, qID[:8], statusClass, status, qID, qID, qID)
	}

	if cards == "" {
		cards = `<div style="text-align:center;padding:40px;color:#94a3b8">
			<p>Tu n'as encore créé aucun quiz.</p>
			<a href="/quiz/create" class="btn-primary" style="margin-top:12px">Créer un quiz</a>
		</div>`
	}

	return renderWithContent(c, "Mes Quiz", fmt.Sprintf(`
<style>
.qe-card{background:#16213e;border:1px solid #2d2d44;border-radius:10px;padding:16px;margin-bottom:16px}
.qe-status-badge{font-size:.75rem;padding:4px 10px;border-radius:4px;font-weight:600}
.qe-published{background:rgba(34,197,94,.15);color:#22c55e}
.qe-draft{background:rgba(251,191,36,.15);color:#fbbf24}
.qe-icon-btn{display:inline-flex;align-items:center;justify-content:center;width:32px;height:32px;border:none;border-radius:6px;background:transparent;cursor:pointer;font-size:1rem;transition:background .15s}
.qe-icon-btn:hover{background:rgba(255,255,255,.1)}
.qe-icon-btn-danger:hover{background:rgba(239,68,68,.15)}
</style>
<div style="max-width:700px;margin:0 auto">
	<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:20px">
		<h1>📚 Mes Quiz</h1>
		<a href="/quiz/create" class="btn-primary btn-sm">+ Créer</a>
	</div>
	%s
</div>`, cards))
}

func (h *Handler) QuizEditPost(c *fiber.Ctx) error {
	return c.Redirect("/quiz/" + c.Params("id") + "/edit")
}

func (h *Handler) ExploreCategory(c *fiber.Ctx) error {
	category := c.Params("category")

	quizBody, _ := h.db.Select("quizzes",
		fmt.Sprintf("category=eq.%s&is_visible=eq.true&order=created_at.desc&select=*", category), true)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizBody, &quizzes)

	var cards string
	for _, q := range quizzes {
		title := database.DBValue(q["title"])
		qID := database.DBValue(q["id"])
		series := database.DBValue(q["series"])
		playCount := database.DBInt(q["play_count"])
		qCount := database.DBInt(q["question_count"])
		cards += fmt.Sprintf(`<div class="explore-card" onclick="location.href='/quiz/%s'">
			<h3>%s</h3>
			<p style="color:#94a3b8;font-size:.85rem">%s</p>
			<p style="color:#6366f1;font-size:.8rem">📋 %d questions · ▶ %d parties</p>
		</div>`, qID, htmlAttr(title), htmlAttr(series), qCount, playCount)
	}

	if cards == "" {
		cards = `<div style="text-align:center;padding:40px;color:#94a3b8">Aucun quiz dans cette catégorie.</div>`
	}

	return renderWithContent(c, category, fmt.Sprintf(`
<div style="max-width:700px;margin:0 auto">
	<h1 style="margin-bottom:20px">%s</h1>
	<div style="display:grid;gap:12px">%s</div>
	<a href="/explore" style="display:inline-block;margin-top:16px;color:#6366f1">← Retour à l'exploration</a>
</div>`, htmlAttr(category), cards))
}

func (h *Handler) CompleteProfile(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	return renderWithContent(c, "Compléter le profil", `
<div style="max-width:500px;margin:40px auto;text-align:center">
	<h1 style="margin-bottom:16px">👋 Bienvenue !</h1>
	<p style="color:#94a3b8;margin-bottom:24px">Choisis un pseudo pour commencer.</p>
	<form method="POST" action="/complete-profile">
		<div class="pe-field">
			<input type="text" name="nickname" class="pe-input" placeholder="Ton pseudo..." required maxlength="20">
		</div>
		<button type="submit" class="btn-primary btn-lg w-full" style="margin-top:12px">Commencer !</button>
	</form>
</div>`)
}

func (h *Handler) CompleteProfilePost(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	nickname := strings.TrimSpace(c.FormValue("nickname"))
	if nickname == "" {
		return c.Redirect("/complete-profile")
	}

	updateData, _ := json.Marshal(map[string]interface{}{
		"nickname": nickname,
	})
	h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", user.ID), updateData, true)

	return c.Redirect("/")
}
